// Package skills provides devgen skill subcommands.
// +genx:doc
package skills

import (
	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/pkg/cmdx"
)

// Cmd is `devgen skill`.
var Cmd *cobra.Command

func init() {
	Cmd = &cobra.Command{
		Use:   "skill",
		Short: "skills reviewing and installation for repository and user level",
	}
	Cmd.AddCommand(
		cmdx.NewCommand("install", &Install{}).Cmd(),
		// cmdx.NewCommand("list", &List{}).Cmd(),
	)
}

type AgentType string

const (
	CLAUDE AgentType = "claude"
	CURSOR AgentType = "cursor"
	CODEX  AgentType = "codex"
)

var (
	Agents = []AgentType{CLAUDE, CURSOR, CODEX}

	AgentDirs = map[AgentType]string{
		CLAUDE: ".claude",
		CURSOR: ".cursor",
		CODEX:  ".agents",
	}
)
