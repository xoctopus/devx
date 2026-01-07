// Package devx defines go project dev repository generators
// +genx:doc
package devx

import (
	"github.com/spf13/cobra"
)

var root = &cobra.Command{Use: "devgen"}

func init() {
	root.AddCommand(CmdAll)
	root.AddCommand(CmdMakefile)
	root.AddCommand(CmdLint)
	root.AddCommand(CmdCI)
	root.AddCommand(CmdInit)
	root.AddCommand(CmdShow)
	root.AddCommand(CmdCode)
}

func Command() *cobra.Command {
	return root
}
