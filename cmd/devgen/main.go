package main

import (
	_ "embed"
	"runtime/debug"
	"strings"

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
		if Name != "" {
			cmd.Printf("%s:%s@%s#%s_%s\n", Name, Branch, Version, CommitID, BuildTime)
		} else {
			if version != "" {
				cmd.Println(strings.TrimSpace(version))
			}
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
