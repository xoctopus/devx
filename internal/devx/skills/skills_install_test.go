package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	. "github.com/xoctopus/x/testx"
)

func TestInstallRepositoryLinks(t *testing.T) {
	base := t.TempDir()
	module := filepath.Join(base, "devx")
	target := filepath.Join(base, "workspace")
	Expect(t, os.MkdirAll(filepath.Join(module, ".agents"), 0o755), Succeed())
	Expect(t, os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.com/m\n\ngo 1.27\n"), 0o644), Succeed())
	Expect(t, os.MkdirAll(target, 0o755), Succeed())

	t.Chdir(module)

	inst := &Install{Path: target, Name: "cursor"}
	Expect(t, inst.precheck(), Succeed())
	Expect(t, inst.install(&cobra.Command{}), Succeed())

	got, err := os.Readlink(filepath.Join(target, ".cursor"))
	Expect(t, err, Succeed())
	Expect(t, got, Equal(filepath.Join("..", "devx", ".agents")))
}
