package devx

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	_ "github.com/xoctopus/genx/devpkg/codex"
	_ "github.com/xoctopus/genx/devpkg/docx"
	_ "github.com/xoctopus/genx/devpkg/enumx"
	"github.com/xoctopus/genx/pkg/genx"
	_ "github.com/xoctopus/sqlx/devpkg/sqlx/v1"
)

var CmdCode = cmdx.NewCommand("code", &Code{}).Cmd()

// Code help code generating
type Code struct {
	// Entry generator entry, default use ./...
	Entry []string `cmd:""`
	// Name generator names
	Name []string `cmd:""`
}

func (g *Code) Exec(cmd *cobra.Command, args ...string) error {
	generators := make([]genx.Generator, 0, len(g.Name))
	for _, name := range g.Name {
		if gs := genx.Get(name); len(gs) > 0 {
			generators = append(generators, gs...)
		}
	}
	if len(generators) == 0 {
		generators = genx.Get()
	}
	if len(g.Entry) == 0 {
		g.Entry = append(g.Entry, "./...")
	}

	ctx := genx.NewContext(&genx.Args{
		Entrypoint: g.Entry,
	})

	if err := ctx.Execute(context.Background(), generators...); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	return nil
}
