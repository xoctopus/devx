package main

import (
	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/appx"

	"github.com/xoctopus/devx/internal/devx"
)

var (
	Name     = "devgen"
	Feature  string
	Version  string
	CommitID string
	CommitAt string
	BuildAt  string

	meta appx.Meta
)

func init() {
	meta = appx.Meta{
		Name:     Name,
		Feature:  Feature,
		Version:  Version,
		CommitID: CommitID,
		CommitAt: CommitAt,
		BuildAt:  BuildAt,
		Runtime:  appx.GetRuntime(),
	}
}

var CmdVersion = &cobra.Command{
	Use:   "version",
	Short: "print the version of DevX/devgen",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(meta.String())
	},
}

func main() {
	cmd := devx.Command()
	cmd.AddCommand(CmdVersion)

	if err := cmd.Execute(); err != nil {
		cmd.Println(err)
		return
	}
}
