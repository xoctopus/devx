package devx

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
)

type Config struct {
	CI       `json:"ci"`
	Makefile `json:"make"`
	Lint     `json:"lint"`
}

var DefaultConfig = Config{
	CI:   CI{CI: true},
	Lint: Lint{Lint: true},
	Makefile: Makefile{
		TestIgnore:   DefaultTestIgnores,
		FormatIgnore: DefaultFormatIgnores,
		Env:          DefaultEnvs,
		HackTest:     false,
		Depends:      DefaultDepends,
	},
}

var CmdAll = cmdx.NewCommand("all", &All{}).Cmd()

// All generates all configurations
type All struct {
	// File devgen config file
	File string `cmd:",default=.devx/config.json"`
}

func (a *All) Exec(cmd *cobra.Command, args ...string) error {
	if err := (&Init{}).Exec(cmd, args...); err != nil {
		return err
	}

	data, err := os.ReadFile(a.File)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	c := &Config{}
	if err = json.Unmarshal(data, c); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	if err = c.Makefile.Exec(cmd, args...); err != nil {
		os.Exit(1)
	}
	if err = c.Lint.Exec(cmd, args...); err != nil {
		os.Exit(1)
	}
	if err = c.CI.Exec(cmd, args...); err != nil {
		os.Exit(1)
	}

	return nil
}
