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
			return
		}
		inf, _ := debug.ReadBuildInfo()
		if inf != nil && len(inf.Main.Version) > 0 {
			cmd.Println(inf.Main.Version)
			return
		}
		if version != "" {
			cmd.Println(strings.TrimSpace(version))
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
