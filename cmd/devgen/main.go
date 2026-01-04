package main

import (
	_ "github.com/xoctopus/genx/devpkg/docx"

	"github.com/xoctopus/devx/internal/devx"
)

func main() {
	cmd := devx.Command()
	if err := cmd.Execute(); err != nil {
		cmd.Println(err)
		return
	}
}
