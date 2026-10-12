package platformembed

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
)

// Names lists top-level skill directory names under platform/.
func Names() ([]string, error) {
	entries, err := fs.ReadDir(FS, Root)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// MaterializeSkill writes one embedded skill tree into destDir (replaces existing).
func MaterializeSkill(name, destDir string) error {
	// embed.FS paths always use '/' (not filepath.Join).
	srcRoot := path.Join(Root, name)
	if _, err := fs.Stat(FS, srcRoot); err != nil {
		return fmt.Errorf("platform skill %q: %w", name, err)
	}
	_ = os.RemoveAll(destDir)
	return fs.WalkDir(FS, srcRoot, func(embedPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(filepath.FromSlash(srcRoot), filepath.FromSlash(embedPath))
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		src, err := FS.Open(embedPath)
		if err != nil {
			return err
		}
		defer func() { _ = src.Close() }()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, src)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}
