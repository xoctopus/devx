package main

import (
	"context"

	_ "github.com/xoctopus/genx/devpkg"
	"github.com/xoctopus/genx/pkg/genx"
)

func main() {
	ctx := genx.NewContext(&genx.Args{
		Entrypoint: []string{"./..."},
	})

	if err := ctx.Execute(context.Background(), genx.Get()...); err != nil {
		panic(err)
	}
}
