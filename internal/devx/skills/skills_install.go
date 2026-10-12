package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xoctopus/devx/internal/devx/skills/platformembed"
)

// Install links project .cursor/.claude to the module .agents and installs embedded platform skills to $HOME.
type Install struct {
	// Name the agent id: cursor, claude, codex
	Name string `cmd:"name"`
	// Path is where project agent symlinks are installed; default is the module root (same as cwd)
	Path string `cmd:"path"`

	cwd    string
	source string
	names  []AgentType
}

func (s *Install) Exec(cmd *cobra.Command, _ ...string) error {
	if err := s.precheck(); err != nil {
		cmd.Println(err)
		return err
	}

	if s.source == "" {
		cmd.Println("SKILLs source not found (.agents/)")
		return nil
	}

	if err := s.install(cmd); err != nil {
		cmd.Println(err)
		return err
	}
	return nil
}

func (s *Install) precheck() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	source, err := filepath.Abs(filepath.Join(cwd, ".agents"))
	if err != nil {
		return err
	}

	i, err := os.Stat(filepath.Join(cwd, "go.mod"))
	if err != nil || i.IsDir() {
		return fmt.Errorf("run skill install in a go module root (go.mod not found)")
	}

	i, err = os.Stat(source)
	if err != nil || !i.IsDir() {
		return nil
	}

	names := Agents
	if s.Name != "" {
		a := AgentType(strings.ToLower(strings.TrimSpace(s.Name)))
		if _, ok := AgentDirs[a]; !ok {
			return fmt.Errorf("invalid agent %q", s.Name)
		}
		names = []AgentType{a}
	}

	s.source, s.cwd, s.names = source, cwd, names
	return nil
}

func (s *Install) install(cmd *cobra.Command) error {
	cmd.Println("installing repository skills.")

	base := s.Path
	if len(base) == 0 {
		base = s.cwd
	}
	base, err := filepath.Abs(base)
	if err != nil {
		return err
	}

	for _, name := range s.names {
		tar, err := filepath.Abs(filepath.Join(base, AgentDirs[name]))
		if err != nil {
			return err
		}

		if tar == s.source {
			continue
		}

		rel, err := filepath.Rel(filepath.Dir(tar), s.source)
		if err != nil {
			return err
		}

		_ = os.Remove(tar)
		if err := os.Symlink(rel, tar); err != nil {
			return fmt.Errorf("link %s: %w", tar, err)
		}
		cmd.Printf("\t%s -> %s\n", tar, rel)
	}

	cmd.Println()
	cmd.Println("installing platform skills.")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("user home: %w", err)
	}

	skills, err := platformembed.Names()
	if err != nil {
		return err
	}
	if len(skills) == 0 {
		cmd.Println("no platform skills embedded in devgen")
		return nil
	}

	for _, skill := range skills {
		for _, agent := range s.names {
			dest := filepath.Join(home, AgentDirs[agent], "skills", skill)

			if err = platformembed.MaterializeSkill(skill, dest); err != nil {
				return fmt.Errorf("install %s: %w", dest, err)
			}
			cmd.Printf("\t[%s] ==> %s\n", skill, dest)
		}
	}
	return nil
}
