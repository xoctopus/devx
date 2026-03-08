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

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	"github.com/xoctopus/x/misc/defers"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/slicex"
	"github.com/xoctopus/x/stringsx"
)

var (
	CmdMakefile = cmdx.NewCommand("make", &Makefile{}).Cmd()
	//go:embed static/target.mk
	gTargetMakefile []byte
	//go:embed static/target.dockerfile
	gTargetDockerfile []byte
)

type Target struct {
	Name    string `json:"name"`
	GenMake bool   `json:"gen_make"`
}

// Makefile generates go project Makefile
type Makefile struct {
	// TestIgnore patterns unit testing and coverage ignores
	// .pb.go _mock.go _genx_ testing.go example/ will be ignored default
	TestIgnore []string `json:"test_ignore" cmd:""`
	// FormatIgnore patterns code formating ignores
	// .git/ .xgo/ *.pb.go *_genx_* will be ignored default
	FormatIgnore []string `json:"format_ignore" cmd:""`
	// Env global env variables
	Env []string `json:"env" cmd:""`
	// HackTest hack test entry
	HackTest bool `json:"hack_test" cmd:",default=false"`
	// Depends dependent tools info
	Depends Depends `json:"depends" cmd:"depends"`
	// Target assigns target entries with name and entry.
	// eg: '{"name":"poc","entry":"cmd/poc"}'
	Target []Target `json:"target" cmd:"target"`
	// Target assigns image entries with name and entry.
	// eg: '{"name":"poc","entry":"cmd/poc"}'
	Image []Target `json:"image" cmd:"image"`
	// EnableBenchCover if enable bench in cover
	EnableBenchCover bool `json:"enable_bench_cover" cmd:",default=false"`

	envs [][2]string
}

func (m *Makefile) Exec(cmd *cobra.Command, args ...string) (err error) {
	m.init(cmd)

	var f *os.File

	f, err = os.OpenFile("Makefile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer defers.Collect(f.Close, &err)

	m.vars(f)
	m.show(f)
	m.dep(f)
	m.tidy(f)
	m.test(f)
	if len(m.Target) > 0 {
		m.target(cmd, f)
		m.image(f)
	}
	m.check(f)

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

func (m *Makefile) vars(w *os.File) {
	_, _ = fmt.Fprintf(w, `
# go package info
MODULE_PATH    := $(shell cat go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
MODULE_NAME    := $(shell basename $(MODULE_PATH))
TEST_IGNORES   := %q
FORMAT_IGNORES := %q

# git repository info
IS_GIT_REPO := $(shell git rev-parse --is-inside-work-tree >/dev/null 2>&1 && echo 1 || echo 0)
ifeq ($(IS_GIT_REPO),1)
export GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
export GIT_TAG    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
export GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
else
export GIT_COMMIT := ""
export GIT_TAG    := ""
export GIT_BRANCH := ""
endif
export BUILD_AT := $(shell date "+%s")
`, strings.Join(m.TestIgnore, "|"), strings.Join(m.FormatIgnore, ","), "%Y%m%d%H%M%S")

	_, _ = w.WriteString("\n# global env variables\n")
	_ = WriteKeyValAlign(w, "export ", ":=", m.envs)

	_, _ = w.WriteString("\n# go build tools\n")
	_ = WriteKeyValAlign(w, "", ":=", DefaultGoTools)

	kvs := make([][2]string, 0)
	for _, d := range m.Depends {
		kvs = append(kvs, [2]string{
			"DEP_" + stringsx.UpperSnakeCase(d.Name),
			"$(shell type " + d.Name + " > /dev/null 2>&1 && echo $$?)",
		})
	}
	_, _ = w.Write([]byte("\n# dependencies flags\n"))
	_ = WriteKeyValAlign(w, "", ":=", kvs)
}

func (m *Makefile) show(w *os.File) {
	text := `
show:
	@echo "module:"
	@echo "	path=$(MODULE_PATH)"
	@echo "	module=$(MODULE_NAME)"
	@echo "git:"
	@echo "	commit_id=$(GIT_COMMIT)"
	@echo "	tag=$(GIT_TAG)"
	@echo "	branch=$(GIT_BRANCH)"
	@echo "	build_time=$(BUILD_AT)"
	@echo "	name=$(MODULE_NAME)"
	@echo "tools:"
	@echo "	build=$(GOBUILD)"
	@echo "	test=$(GOTEST)"
`
	for _, d := range m.Depends {
		text += fmt.Sprintf(`	@echo "	%s=$(shell which %s) $(%s)"`, d.Name, d.Name, d.DepKey()) + "\n"
	}
	_, _ = w.WriteString(text)
}

func (m *Makefile) dep(w *os.File) {
	text := `
dep:
	@echo "==> installing dependencies"`
	for _, d := range m.Depends {
		text += fmt.Sprintf(`
	@if [ "${%s}" != "0" ]; then \
		echo "	%s for %s"; \
		go install %s@%s; \
		echo "	DONE."; \
	fi`, d.DepKey(), d.Name, d.Description, d.Repo, d.Version)
	}

	_, _ = w.WriteString(text + "\n")

	text = `
upgrade-dep:
	@echo "==> upgrading dependencies"`

	for _, d := range m.Depends {
		text += fmt.Sprintf(`
	@echo "	%s for %s"
	@go install %s@%s
	@echo "	DONE."`, d.Name, d.Description, d.Repo, d.Version)
	}

	_, _ = w.WriteString(text + "\n")
}

func (m *Makefile) tidy(w *os.File) {
	text := `
tidy:
	@echo "==> go mod tidy"
	@go mod tidy`
	_, _ = w.WriteString(text + "\n")
}

func (m *Makefile) test(w *os.File) {
	m.hack(w)

	dep := "dep tidy"
	if m.HackTest {
		dep += " hack_dep_run"
	}
	bench := ""
	if m.EnableBenchCover {
		bench = "-bench=. "
	}

	text := fmt.Sprintf(`
test: %s
	@echo "==> run unit test"
	@$(GOTEST) test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"

cover: %s
	@echo "==> run unit test with coverage"
	@$(GOTEST) test ./... %s-failfast -parallel 1 -gcflags="all=-N -l" -covermode=count -coverprofile=cover.out
	@grep -vE $(TEST_IGNORES) cover.out > cover2.out && mv cover2.out cover.out

view-cover: cover
	@echo "==> run unit test with coverage and view results"
	@$(GOBUILD) tool cover -html cover.out

ci-cover: lint cover
`, dep, dep, bench)
	_, _ = w.WriteString(text + "\n")
}

func (m *Makefile) hack(w *os.File) {
	if !m.HackTest {
		return
	}

	_, _ = w.WriteString(`
hack_dep_run:
	@cd hack && docker compose up -d --remove-orphans

hack_dep_stop:
	@cd hack && docker compose down -v
`)
}

func (m *Makefile) target(cmd *cobra.Command, w *os.File) {
	names := make([]string, 0, len(m.Target))
	for _, t := range m.Target {
		m.cmdMake(cmd, t)
		entry := filepath.Join("cmd", t.Name)
		fi, err := os.Stat(entry)
		if err != nil && os.IsNotExist(err) || !fi.IsDir() {
			fmt.Printf("WARN: target entry `%s` is not exists or not a folder\n", entry)
			continue
		}

		_, _ = fmt.Fprintf(w, `
target_%s:
	@make -C %s --no-print-directory install
`, t.Name, entry)
		names = append(names, "target_"+t.Name)
	}

	_, _ = fmt.Fprintf(w, `
targets: %s
`, strings.Join(names, " "))
}

func (m *Makefile) image(w *os.File) {
	names := make([]string, 0, len(m.Target))
	for _, t := range m.Image {
		entry := filepath.Join("cmd", t.Name)
		_, _ = fmt.Fprintf(w, `
image_%s:
	@make -C %s --no-print-directory image
`, t.Name, entry)
		names = append(names, "image_"+t.Name)
	}

	_, _ = fmt.Fprintf(w, `
images: %s
`, strings.Join(names, " "))
}

func (m *Makefile) cmdMake(cmd *cobra.Command, t Target) {
	if !t.GenMake {
		return
	}
	filename := filepath.Join("cmd", t.Name, "Makefile")
	f := must.NoErrorV(os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666))
	defer func() { _ = f.Close() }()
	must.NoErrorV(io.Copy(f, bytes.NewReader(gTargetMakefile)))
	cmd.Printf("==> generated %s\n", filename)
}

func (m *Makefile) check(w *os.File) {
	text := `
fmt: dep clean
	@echo "==> formating code"
	@goimports-reviser -rm-unused \
		-imports-order 'std,general,company,project' \
		-project-name ${MODULE_PATH} \
		-excludes $(FORMAT_IGNORES) ./...

fmt-check: fmt
	@echo "==> checking code format"
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "code is not properly formatted."; \
		exit 1; \
	fi

lint: dep
	@echo "==> linting"
	@echo ">>>golangci-lint"
	@golangci-lint run
	@go vet ./...
	@echo "done"

clean:
	@find . -name cover.out | xargs rm -rf
	@find . -name .xgo | xargs rm -rf
	@rm -rf build/*

changelog:
	@git chglog --next-tag HEAD -o CHANGELOG.md || true

pre-commit: dep fmt lint view-cover changelog`
	if len(m.Target) > 0 {
		text = text + " targets"
	}
	text = text + "\n"

	_, _ = w.WriteString(text)
}
