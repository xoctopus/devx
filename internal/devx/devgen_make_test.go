package devx

import (
	"bytes"
	"testing"

	. "github.com/xoctopus/x/testx"
)

func TestProjectMakefileTemplateVendorMode(t *testing.T) {
	tool := Tool{
		Name:        "devgen",
		Repo:        "github.com/xoctopus/devx/cmd/devgen",
		Version:     "main",
		Description: "dev configuration generating",
	}
	tool.SetDefault()

	data := ProjectMakeData{
		Depends: Depends{tool},
	}

	var buf bytes.Buffer
	Expect(t, gProjectTplMakefile.Execute(&buf, data), Succeed())

	out := buf.String()
	Expect(t, out, ContainsSubString("wildcard vendor/modules.txt"))
	Expect(t, out, ContainsSubString("GO_INSTALL := GOFLAGS=-mod=mod go install"))
	Expect(t, out, ContainsSubString("$(GO_INSTALL)"))
}
