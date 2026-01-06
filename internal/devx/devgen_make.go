package devx

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	"github.com/xoctopus/x/misc/defers"
	"github.com/xoctopus/x/slicex"
	"github.com/xoctopus/x/stringsx"
)

var CmdMakefile = cmdx.NewCommand("make", &Makefile{}).Cmd()

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
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
GIT_TAG    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
else
GIT_COMMIT := ""
GIT_TAG    := ""
GIT_BRANCH := ""
endif
BUILD_AT=$(shell date "+%s")
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

	text := fmt.Sprintf(`
test: %s
	@echo "==> run unit test"
	@$(GOTEST) test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"

cover: %s
	@echo "==> run unit test with coverage"
	@$(GOTEST) test ./... -failfast -parallel 1 -gcflags="all=-N -l" -covermode=count -coverprofile=cover.out
	@grep -vE $(TEST_IGNORES) cover.out > cover2.out && mv cover2.out cover.out

view-cover: cover
	@echo "==> run unit test with coverage and view results"
	@$(GOBUILD) tool cover -html cover.out

ci-cover: lint cover
`, dep, dep)
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

func (m *Makefile) check(w *os.File) {
	text := `
fmt: dep clean
	@echo "==> formating code"
	@goimports-reviser -rm-unused \
		-imports-order 'std,general,company,project' \
		-project-name ${MODULE_PATH} \
		-excludes $(FORMAT_IGNORES) ./...

lint: dep
	@echo "==> linting"
	@echo ">>>golangci-lint"
	@golangci-lint run
	@echo "done"

clean:
	@find . -name cover.out | xargs rm -rf
	@find . -name .xgo | xargs rm -rf
	@rm -rf build/*

changelog:
	@git chglog -o CHANGELOG.md || true

pre-commit: dep fmt lint view-cover changelog
`
	_, _ = w.WriteString(text)
}
