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
	//go:embed static/ci.yml
	gWorkflowCI []byte
	//go:embed static/dependabot.yml
	gDependabot []byte

	CmdCI = cmdx.NewCommand("ci", &CI{}).Cmd()
)

// CI generates ci configuration
type CI struct {
	// CI if enable ci configuration generating
	CI bool `cmd:",default=true,noopdef=true"`
}

func (c *CI) Exec(cmd *cobra.Command, args ...string) (err error) {
	if c.CI == false {
		return nil
	}

	if err := os.MkdirAll("./.github/workflows", 0o755); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}

	f1, err := os.OpenFile("./.github/workflows/ci.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	defer defers.Collect(f1.Close, &err)

	if _, err = io.Copy(f1, bytes.NewReader(gWorkflowCI)); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	cmd.Println("==> generated .github/workflows/ci.yml")

	f2, err := os.OpenFile("./.github/dependabot.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	defer defers.Collect(f2.Close, &err)

	if _, err = io.Copy(f2, bytes.NewReader(gDependabot)); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
	cmd.Println("==> generated .github/dependabot.yml")

	return nil
}
