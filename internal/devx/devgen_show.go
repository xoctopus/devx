package devx

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	"github.com/xoctopus/x/misc/defers"
)

var CmdShow = cmdx.NewCommand("show", &Show{}).Cmd()

// Show shows devgen configuration
type Show struct{}

func (c *Show) Exec(cmd *cobra.Command, args ...string) (err error) {
	if FileCheck(ConfigFilename, false) {
		f, err := os.OpenFile(ConfigFilename, os.O_RDWR, 0644)
		if err != nil {
			cmd.Println(err)
			os.Exit(1)
		}
		defer defers.Collect(f.Close, &err)

		data, err := io.ReadAll(f)
		if err != nil {
			cmd.Println(err)
			os.Exit(1)
		}

		cmd.Printf("config file: %s\n%s\n", ConfigFilename, string(data))
	}

	return (&Init{}).Exec(cmd)
}
