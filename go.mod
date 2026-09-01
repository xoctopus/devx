module github.com/xoctopus/devx

go 1.27.0

tool (
	github.com/xoctopus/devx/internal/cmd/gen
	github.com/xoctopus/devx/internal/cmd/skill-install
)

require (
	github.com/spf13/cobra v1.10.2
	// +skill:appx
	github.com/xoctopus/confx v0.5.9
	// +skill:genx
	github.com/xoctopus/genx v0.3.8
	// +skill:testx
	github.com/xoctopus/x v0.5.8
)

require (
	github.com/fatih/color v1.19.0 // indirect
	github.com/go-think/openssl v1.22.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xoctopus/pkgx v0.4.4 // indirect
	github.com/xoctopus/typx v0.4.7 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
