package devx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	. "github.com/xoctopus/x/testx"
)

func TestHasVendorModules(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	Expect(t, err, Succeed())
	Expect(t, os.Chdir(dir), Succeed())
	t.Cleanup(func() { _ = os.Chdir(wd) })

	Expect(t, hasVendorModules(), Equal(false))

	mustMkdirAll(t, filepath.Join(dir, "vendor"))
	Expect(t, hasVendorModules(), Equal(false))

	mustWriteFile(t, filepath.Join(dir, "vendor", "modules.txt"), "")
	Expect(t, hasVendorModules(), Equal(true))
}

func TestProjectMakefileTemplateHasVendor(t *testing.T) {
	tool := Tool{
		Name:        "devgen",
		Repo:        "github.com/xoctopus/devx/cmd/devgen",
		Version:     "main",
		Description: "dev configuration generating",
	}
	tool.SetDefault()

	data := ProjectMakeData{
		HasVendor: true,
		Depends:   Depends{tool},
	}

	var buf bytes.Buffer
	Expect(t, gProjectTplMakefile.Execute(&buf, data), Succeed())

	out := buf.String()
	Expect(t, out, ContainsSubString("-mod=vendor"))
	Expect(t, out, ContainsSubString("GOFLAGS=-mod=mod go install"))
	Expect(t, out, Not(ContainsSubString("wildcard vendor/modules.txt")))
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	Expect(t, os.MkdirAll(path, 0o755), Succeed())
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	Expect(t, os.WriteFile(path, []byte(content), 0o644), Succeed())
}
