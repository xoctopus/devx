package main

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/xoctopus/devx/internal/devx"
)

var (
	Name      string
	Branch    string
	Version   string
	CommitID  string
	BuildTime string

	//go:embed version
	version string
)

var CmdVersion = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("%s:%s@%s#%s_%s\n", Name, Branch, Version, CommitID, BuildTime)
	},
	Short: "print the version of DevX/devgen",
}

func main() {
	cmd := devx.Command()
	cmd.AddCommand(CmdVersion)

	if err := cmd.Execute(); err != nil {
		cmd.Println(err)
		return
	}
}
