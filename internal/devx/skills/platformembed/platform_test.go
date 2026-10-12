package platformembed_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xoctopus/devx/internal/devx/skills/platformembed"
)

func TestNames(t *testing.T) {
	names, err := platformembed.Names()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("expected embedded platform skills")
	}
}

func TestMaterializeSkill(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "base")
	if err := platformembed.MaterializeSkill("base", dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md: %v", err)
	}
}
