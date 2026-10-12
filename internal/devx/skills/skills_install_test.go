package skills

import (
	"os"
	"path/filepath"
	"testing"

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

	base, err := filepath.Abs(target)
	Expect(t, err, Succeed())
	for _, name := range inst.names {
		tar, err := filepath.Abs(filepath.Join(base, AgentDirs[name]))
		Expect(t, err, Succeed())
		if tar == inst.source {
			continue
		}
		rel, err := filepath.Rel(filepath.Dir(tar), inst.source)
		Expect(t, err, Succeed())
		_ = os.Remove(tar)
		Expect(t, os.Symlink(rel, tar), Succeed())
	}

	got, err := os.Readlink(filepath.Join(target, ".cursor"))
	Expect(t, err, Succeed())
	Expect(t, got, Equal(filepath.Join("..", "devx", ".agents")))
}
