// Package platformembed holds devgen-bundled platform skills (devx, base, ...).
//
// Edit sources under platform/ only. Do not put devx/base under repo .agents/skills;
// that directory is for go tool skill-install (go.mod // +skill:) dependency skills.
package platformembed

import "embed"

//go:embed all:platform
var FS embed.FS

const Root = "platform"
