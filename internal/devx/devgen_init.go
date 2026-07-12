package devx

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	"github.com/xoctopus/x/misc/cleanup"
)

var CmdInit = cmdx.NewCommand("init", &Init{}).Cmd()

// Init initializes devgen configuration
type Init struct{}

func (d *Init) Exec(cmd *cobra.Command, args ...string) (err error) {
	if FileCheck(ConfigFilename, false) {
		cmd.Printf("devx configuration: %s\n", ConfigFilename)
		return nil
	}

	if err = os.MkdirAll(filepath.Dir(ConfigFilename), 0755); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(DefaultConfig, "", "\t")
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	co := cleanup.NewCollector()
	defer func() { err = co.JoinTo(nil) }()

	f, err := os.OpenFile(ConfigFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	co.Collect(f.Close)

	if _, err = io.Copy(f, bytes.NewReader(data)); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	cmd.Printf("config init to %s\n%s\n", ConfigFilename, string(data))
	return nil
}
