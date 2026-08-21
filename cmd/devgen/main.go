package main

import (
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xoctopus/devx/internal/devx"
)

var (
	Name     string
	Feature  string
	Version  string
	CommitID string
	CommitAt string
	BuildAt  string
)

var CmdVersion = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		if Name != "" {
			cmd.Printf("%s:%s@%s#%s()[commit=%s|build=%s]\n",
				Name, Feature, Version, CommitID, CommitAt, BuildAt)
		}
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, dep := range info.Deps {
				if strings.HasPrefix(dep.Path, "github.com/xoctopus/") &&
					dep.Path != info.Main.Path {
					cmd.Printf("%s %s\n", dep.Path, dep.Version)
				}
			}
		}
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
