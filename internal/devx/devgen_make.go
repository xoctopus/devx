package devx

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	"github.com/xoctopus/x/misc/cleanup"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/slicex"
	"github.com/xoctopus/x/stringsx"
)

var (
	CmdMakefile = cmdx.NewCommand("make", &Makefile{}).Cmd()
	//go:embed static/project.tpl.mk
	gProjectMakefile []byte
	//go:embed static/meta.tpl.mk
	gMetaMakefile []byte
	//go:embed static/target.tpl.mk
	gTargetMakefile []byte
	//go:embed static/target.tpl.dockerfile
	gTargetImage []byte

	gProjectTplMakefile  = parseTemplates("project", gMetaMakefile, gProjectMakefile)
	gTargetTplMakefile   = parseTemplates("target", gMetaMakefile, gTargetMakefile)
	gTargetTplDockerfile = must.NoErrorV(template.New("image").Parse(string(gTargetImage)))
)

func parseTemplates(name string, parts ...[]byte) *template.Template {
	t := template.New(name)
	for _, part := range parts {
		t = must.NoErrorV(t.Parse(string(part)))
	}
	return t
}

// Target cmd/<name> build target
type Target struct {
	// Name target binary name, entry is cmd/<name>
	Name string `json:"name"`
	// Image optional docker image options
	Image *ImageOption `json:"image"`
}

// ImageOption docker image build options
type ImageOption struct {
	// Runtime base image, default alpine:latest
	Runtime string `json:"runtime,omitempty"`
	// TimeZone container timezone
	TimeZone string `json:"timezone,omitempty"`
	// CgoEnabled enable CGO for image build
	CgoEnabled bool `json:"cgo_enabled,omitempty"`
	// GoProxy GOPROXY for image build
	GoProxy string `json:"go_proxy,omitempty"`
	// Expose container expose port
	Expose string `json:"expose,omitempty"`
	// RunArgs image run arguments
	RunArgs []string `json:"run_args,omitempty"`
}

type EnvVar struct {
	Key   string
	Value string
}

type EnvVars []EnvVar

type EnvShow struct {
	Key     string
	Padding string
}

type ProjectTarget struct {
	Name  string
	Entry string
	Image bool
}

type ProjectMakeData struct {
	GoModDir      string
	TestIgnores   string
	FormatIgnores string
	Envs          []EnvVar
	EnvShows      []EnvShow
	GoTools       string
	DepFlags      string
	Depends       Depends
	TestDep       string
	BenchFlag     string
	HackTest      bool
	Targets       []ProjectTarget
	TargetNames   string
	ImageNames    string
	HasTargets    bool
	HasImages     bool
}

// Makefile generates go project Makefile
type Makefile struct {
	// TestIgnore patterns unit testing and coverage ignores
	TestIgnore []string `json:"test_ignore" cmd:""`
	// FormatIgnore patterns code formating ignores
	FormatIgnore []string `json:"format_ignore" cmd:""`
	// Env global env variables
	Env []string `json:"env" cmd:""`
	// HackTest hack test entry
	HackTest bool `json:"hack_test" cmd:",default=false"`
	// Depends dependent tools info
	Depends Depends `json:"depends" cmd:"depends"`
	// Target assigns build targets under cmd/<name>; optional image generates Dockerfile
	Target []Target `json:"target" cmd:"target"`
	// EnableBenchCover if enable bench in cover
	EnableBenchCover bool `json:"enable_bench_cover" cmd:",default=false"`

	envs [][2]string
}

func (m *Makefile) Exec(cmd *cobra.Command, args ...string) (err error) {
	m.init(cmd)

	co := cleanup.NewCollector()
	defer func() { err = co.JoinTo(&err) }()

	var f *os.File
	f, err = os.OpenFile("Makefile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	co.Collect(f.Close)

	data := m.buildProjectData(cmd)
	if err = gProjectTplMakefile.Execute(f, data); err != nil {
		return err
	}

	cmd.Println("==> generated Makefile")
	return nil
}

func (m *Makefile) init(cmd *cobra.Command) {
	if !FileCheck("go.mod", false) {
		cmd.Println("please run this command in a go project root")
		os.Exit(1)
	}

	if len(m.Env) > 0 && len(m.Env)%2 != 0 {
		cmd.Println("Env variables must be kv pairs")
		os.Exit(1)
	}

	m.TestIgnore = slicex.Unique(append(DefaultTestIgnores, m.TestIgnore...))
	m.FormatIgnore = slicex.Unique(append(DefaultFormatIgnores, m.FormatIgnore...))

	if len(m.Env) == 0 {
		m.Env = DefaultEnvs
	}

	m.envs = make([][2]string, 0, len(m.Env))
	if m.HackTest {
		m.envs = append(m.envs, [2]string{"HACK_TEST", "true"})
	}
	for i := 0; i < len(m.Env); i += 2 {
		m.envs = append(m.envs, [2]string{m.Env[i], m.Env[i+1]})
	}

	m.Depends = append(DefaultDepends, m.Depends...)
	m.Depends = slicex.UniqueValues(m.Depends, func(e Tool) string { return e.Name })
	slices.SortFunc(m.Depends, func(a, b Tool) int { return cmp.Compare(a.Name, b.Name) })
	for i := range m.Depends {
		m.Depends[i].SetDefault()
	}
}

func (m *Makefile) buildProjectData(cmd *cobra.Command) ProjectMakeData {
	depKVs := make([][2]string, 0, len(m.Depends))
	for _, d := range m.Depends {
		depKVs = append(depKVs, [2]string{
			"DEP_" + stringsx.UpperSnakeCase(d.Name),
			"$(shell type " + d.Name + " > /dev/null 2>&1 && echo $$?)",
		})
	}

	envPad := 0
	for _, v := range m.envs {
		if len(v[0]) > envPad {
			envPad = len(v[0])
		}
	}
	envShows := make([]EnvShow, 0, len(m.envs))
	for _, v := range m.envs {
		envShows = append(envShows, EnvShow{
			Key:     v[0],
			Padding: strings.Repeat(" ", envPad-len(v[0])),
		})
	}

	testDep := "dep tidy"
	if m.HackTest {
		testDep += " hack_dep_run"
	}

	benchFlag := ""
	if m.EnableBenchCover {
		benchFlag = "-bench=. "
	}

	targets, targetNames, imageNames := m.resolveTargets(cmd)

	envs := make([]EnvVar, 0, len(m.envs))
	for _, v := range m.envs {
		envs = append(envs, EnvVar{Key: v[0], Value: v[1]})
	}

	return ProjectMakeData{
		GoModDir:      "",
		TestIgnores:   strings.Join(m.TestIgnore, "|"),
		FormatIgnores: strings.Join(m.FormatIgnore, ","),
		Envs:          envs,
		EnvShows:      envShows,
		GoTools:       formatKeyValAlign("", ":=", DefaultGoTools),
		DepFlags:      formatKeyValAlign("", ":=", depKVs),
		Depends:       m.Depends,
		TestDep:       testDep,
		BenchFlag:     benchFlag,
		HackTest:      m.HackTest,
		Targets:       targets,
		TargetNames:   strings.Join(targetNames, " "),
		ImageNames:    strings.Join(imageNames, " "),
		HasTargets:    len(targetNames) > 0,
		HasImages:     len(imageNames) > 0,
	}
}

func (m *Makefile) resolveTargets(cmd *cobra.Command) ([]ProjectTarget, []string, []string) {
	targets := make([]ProjectTarget, 0, len(m.Target))
	targetNames := make([]string, 0, len(m.Target))
	imageNames := make([]string, 0, len(m.Target))

	for _, t := range m.Target {
		entry := filepath.Join("cmd", t.Name)
		fi, err := os.Stat(entry)
		if err != nil && os.IsNotExist(err) || !fi.IsDir() {
			fmt.Printf("WARN: target entry `%s` is not exists or not a folder\n", entry)
			continue
		}

		m.cmdMake(cmd, t)

		targets = append(targets, ProjectTarget{
			Name:  t.Name,
			Entry: entry,
			Image: t.Image != nil,
		})
		targetNames = append(targetNames, "target_"+t.Name)
		if t.Image != nil {
			imageNames = append(imageNames, "image_"+t.Name)
		}
	}

	return targets, targetNames, imageNames
}

func formatKeyValAlign(prefix, join string, kvs [][2]string) string {
	var b strings.Builder
	_ = WriteKeyValAlign(&b, prefix, join, kvs)
	return strings.TrimRight(b.String(), "\n")
}

func (m *Makefile) cmdMake(cmd *cobra.Command, t Target) {
	filename := filepath.Join("cmd", t.Name, "Makefile")
	f := must.NoErrorV(os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666))
	defer func() { _ = f.Close() }()

	must.NoError(gTargetTplMakefile.Execute(f, map[string]any{
		"Image":    t.Image != nil,
		"GoModDir": "../../",
	}))

	if t.Image != nil {
		m.cmdImage(cmd, t.Name, t.Image)
	}
	cmd.Printf("==> generated %s\n", filename)
}

func (m *Makefile) cmdImage(cmd *cobra.Command, name string, i *ImageOption) {
	goVersion := must.NoErrorV(goModVersion())
	runtime := cmp.Or(i.Runtime, "alpine:latest")
	cgoEnabled := "0"
	if i.CgoEnabled {
		cgoEnabled = "1"
	}

	var buf bytes.Buffer
	must.NoError(gTargetTplDockerfile.Execute(&buf, map[string]any{
		"Name":       name,
		"GoVersion":  goVersion,
		"Runtime":    runtime,
		"CgoEnabled": cgoEnabled,
		"GoProxy":    i.GoProxy,
		"TimeZone":   i.TimeZone,
		"Expose":     i.Expose,
		"RunArgs":    i.RunArgs,
	}))

	filename := filepath.Join("cmd", name, "Dockerfile")
	f := must.NoErrorV(os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666))
	defer func() { _ = f.Close() }()
	must.NoErrorV(io.Copy(f, &buf))
	cmd.Printf("==> generated %s\n", filename)
}

func goModVersion() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "go "); ok {
			if ver := strings.TrimSpace(after); ver != "" {
				return ver, nil
			}
		}
	}
	return "", fmt.Errorf("go version not found in go.mod")
}
