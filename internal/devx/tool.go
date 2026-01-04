package devx

import "github.com/xoctopus/x/stringsx"

// ToolType make depends tool type
// +genx:enum
type ToolType int

const (
	TOOL_TYPE_UNKNOWN   ToolType = iota
	TOOL_TYPE__LINTER            // code linting
	TOOL_TYPE__DOC               // doc generating
	TOOL_TYPE__FORMATER          // code formating
	TOOL_TYPE__DEVGEN            // devgen
)

type Tool struct {
	Name        string   `json:"name"`
	Repo        string   `json:"repo"`
	Version     string   `json:"version"`
	Type        ToolType `json:"type"`
	Description string   `json:"description"`
}

func (t *Tool) SetDefault() {
	if t.Version == "" {
		t.Version = "latest"
	}
	if t.Description == "" {
		t.Description = t.Type.Text()
	}
}

func (t *Tool) DepKey() string {
	return "DEP_" + stringsx.UpperSnakeCase(t.Name)
}

type Depends []Tool

var DefaultDepends = Depends{
	{
		Name:    "golangci-lint",
		Repo:    "github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
		Version: "latest",
		Type:    TOOL_TYPE__LINTER,
	},
	{
		Name:    "goimports-reviser",
		Repo:    "github.com/incu6us/goimports-reviser/v3",
		Version: "latest",
		Type:    TOOL_TYPE__FORMATER,
	},
	{
		Name:        "git-chglog",
		Repo:        "github.com/git-chglog/git-chglog/cmd/git-chglog",
		Version:     "latest",
		Type:        TOOL_TYPE__DOC,
		Description: "generating changelog",
	},
	{
		Name:        "devgen",
		Repo:        "github.com/xoctopus/devgen/cmd/devgen",
		Version:     "main",
		Type:        TOOL_TYPE__DEVGEN,
		Description: "dev configuration generating",
	},
}
