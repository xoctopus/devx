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
		GoModDir: "",
		Depends:  Depends{tool},
	}

	var buf bytes.Buffer
	Expect(t, gProjectTplMakefile.Execute(&buf, data), Succeed())

	out := buf.String()
	Expect(t, out, ContainsSubString("wildcard vendor/modules.txt"))
	Expect(t, out, ContainsSubString("GO_INSTALL := GOFLAGS=-mod=mod go install"))
	Expect(t, out, ContainsSubString("META_TZ := Asia/Shanghai"))
	Expect(t, out, ContainsSubString("TZ=$(META_TZ) date \"+%Y%m%d%H%M%SCST\""))
	Expect(t, out, ContainsSubString("date=format:%Y%m%d%H%M%SCST"))
}

func TestTargetMakefileTemplateVendorMode(t *testing.T) {
	var buf bytes.Buffer
	Expect(t, gTargetTplMakefile.Execute(&buf, map[string]any{
		"Image":    false,
		"GoModDir": "../../",
	}), Succeed())

	out := buf.String()
	Expect(t, out, ContainsSubString("wildcard ../../vendor/modules.txt"))
	Expect(t, out, ContainsSubString("GO_MOD_FLAG := -mod=vendor"))
	Expect(t, out, ContainsSubString("go build $(GO_MOD_FLAG)"))
	Expect(t, out, ContainsSubString("export GOFLAGS := $(GOFLAGS) -buildvcs=false"))
	Expect(t, out, ContainsSubString("cat ../../go.mod"))
	Expect(t, out, ContainsSubString("export GIT_COMMIT"))
	Expect(t, out, ContainsSubString("export BUILD_AT"))
}

func TestTargetDockerfileTemplateRunArgs(t *testing.T) {
	var buf bytes.Buffer
	Expect(t, gTargetTplDockerfile.Execute(&buf, map[string]any{
		"Name":       "demo",
		"GoVersion":  "1.27.0",
		"Runtime":    "alpine:latest",
		"CgoEnabled": "0",
		"Expose":     "80",
		"RunArgs":    []string{"run"},
	}), Succeed())
	Expect(t, buf.String(), ContainsSubString("ENTRYPOINT [\"/app/app\", \"run\"]"))

	buf.Reset()
	Expect(t, gTargetTplDockerfile.Execute(&buf, map[string]any{
		"Name":       "demo",
		"GoVersion":  "1.27.0",
		"Runtime":    "alpine:latest",
		"CgoEnabled": "0",
	}), Succeed())
	Expect(t, buf.String(), ContainsSubString("ENTRYPOINT [\"/app/app\"]"))
	Expect(t, buf.String(), Not(ContainsSubString("ENTRYPOINT [\"/app/app\", ")))
}
