package devx

import (
	"bytes"
	_ "embed"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
	"github.com/xoctopus/x/misc/defers"
)

var (
	//go:embed static/golangci.yml
	gLintConfig []byte

	CmdLint = cmdx.NewCommand("lint", &Lint{}).Cmd()
)

// Lint generates lint configuration
type Lint struct {
	// Lint if enable lint generating .golangci.yml
	Lint bool `json:"lint" cmd:",default=true,noopdef=true"`
}

func (l *Lint) Exec(cmd *cobra.Command, args ...string) (err error) {
	if !l.Lint {
		return nil
	}

	f, err := os.OpenFile(".golangci.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	defer defers.Collect(f.Close, &err)

	if _, err = io.Copy(f, bytes.NewReader(gLintConfig)); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	cmd.Println("==> generated .golangci.yml")
	return nil
}
