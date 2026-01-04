package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/xoctopus/genx/devpkg/codex"
	_ "github.com/xoctopus/genx/devpkg/docx"
	_ "github.com/xoctopus/genx/devpkg/enumx"
	"github.com/xoctopus/genx/pkg/genx"
	_ "github.com/xoctopus/sqlx/devpkg/sqlx/v1"
	"github.com/xoctopus/x/misc/must"
)

func main() {
	cwd := must.NoErrorV(os.Getwd())
	fmt.Println("Working DIR:", cwd)

	ctx := genx.NewContext(&genx.Args{
		Entrypoint: []string{
			filepath.Join(cwd, "..."),
		},
	})

	if err := ctx.Execute(context.Background(), genx.Get()...); err != nil {
		panic(err)
	}
}
