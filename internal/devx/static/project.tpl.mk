
# go package info
MODULE_PATH    := $(shell cat go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
MODULE_NAME    := $(shell basename $(MODULE_PATH))
TEST_IGNORES   := "{{.TestIgnores}}"
FORMAT_IGNORES := "{{.FormatIgnores}}"

# git repository info
IS_GIT_REPO := $(shell git rev-parse --is-inside-work-tree >/dev/null 2>&1 && echo 1 || echo 0)
ifeq ($(IS_GIT_REPO),1)
export GIT_COMMIT_RAW := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
export GIT_COMMIT_AT  := $(shell git log -1 --format=%cd --date=format:%Y%m%d%H%M%S 2>/dev/null || echo "")
export GIT_TAG        := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
export GIT_BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
ifeq ($(shell git status --porcelain 2>/dev/null),)
export GIT_COMMIT := $(GIT_COMMIT_RAW)
else
export GIT_COMMIT := $(GIT_COMMIT_RAW)-dirty
endif
else
export GIT_COMMIT    := ""
export GIT_COMMIT_AT := ""
export GIT_TAG       := v0.0.0
export GIT_BRANCH    := ""
endif
export BUILD_AT := $(shell date "+%Y%m%d%H%M%S")
export MODULE_PATH

# global env variables
{{range .Envs}}{{.Key}} ?= {{.Value}}
export {{.Key}}
{{end}}
ifneq ($(wildcard vendor/modules.txt),)
export GOFLAGS := $(GOFLAGS) -mod=vendor
GO_INSTALL := GOFLAGS=-mod=mod go install
else
GO_INSTALL := go install
endif

# go build tools
{{.GoTools}}

# dependencies flags
{{.DepFlags}}

show:
	@echo "module:"
	@echo "	path=$(MODULE_PATH)"
	@echo "	module=$(MODULE_NAME)"
	@echo "git:"
	@echo "	commit_id=$(GIT_COMMIT)"
	@echo "	commit_at=$(GIT_COMMIT_AT)"
	@echo "	tag=$(GIT_TAG)"
	@echo "	branch=$(GIT_BRANCH)"
	@echo "	build_at=$(BUILD_AT)"
	@echo "	name=$(MODULE_NAME)"
	@echo "tools:"
	@echo "	build=$(GOBUILD)"
	@echo "	test=$(GOTEST)"
{{range .Depends}}	@echo "	{{.Name}}=$(shell which {{.Name}}) $({{.DepKey}})"
{{end}}	@echo "envs:"
{{range .EnvShows}}	@echo "	{{.Key}}: {{.Padding}}$({{.Key}})"
{{end}}	@echo "	GOFLAGS: $(GOFLAGS)"

dep:
	@echo "==> installing dependencies"
{{range .Depends}}	@if [ "${{"{"}}{{.DepKey}}{{"}"}}" != "0" ]; then \
		echo "	{{.Name}} for {{.Description}}"; \
		$(GO_INSTALL) {{.Repo}}@{{.Version}}; \
		echo "	DONE."; \
	fi
{{end}}
upgrade-dep:
	@echo "==> upgrading dependencies"
{{range .Depends}}	@echo "	{{.Name}} for {{.Description}}"
	@$(GO_INSTALL) {{.Repo}}@{{.Version}}
	@echo "	DONE."
{{end}}
tidy:
	@echo "==> go mod tidy"
	@go mod tidy
	@if [ -d vendor ]; then \
		echo "==> go mod vendor"; \
		go mod vendor; \
	fi
{{if .HackTest}}
hack_dep_run:
	@cd hack && docker compose up -d --remove-orphans

hack_dep_stop:
	@cd hack && docker compose down -v
{{end}}
test: {{.TestDep}}
	@echo "==> run unit test"
	@$(GOTEST) test ./... -race -failfast -parallel 1 -gcflags="all=-N -l"

cover: {{.TestDep}}
	@echo "==> run unit test with coverage"
	@$(GOTEST) test ./... {{.BenchFlag}}-failfast -parallel 1 -gcflags="all=-N -l" -covermode=count -coverprofile=cover.out
	@grep -vE $(TEST_IGNORES) cover.out > cover2.out && mv cover2.out cover.out

view-cover: cover
	@echo "==> run unit test with coverage and view results"
	@$(GOBUILD) tool cover -html cover.out

ci-cover: lint cover

{{range .Targets}}
target_{{.Name}}:
	@make -C {{.Entry}} --no-print-directory install
{{if .Image}}
image_{{.Name}}:
	@make -C {{.Entry}} --no-print-directory image
{{end}}{{end}}{{if .HasTargets}}
targets: {{.TargetNames}}
{{end}}{{if .HasImages}}
images: {{.ImageNames}}
{{end}}
fmt: dep clean
	@echo "==> formating code"
	@goimports-reviser -rm-unused \
		-imports-order 'std,general,company,project' \
		-project-name ${MODULE_PATH} \
		-excludes $(FORMAT_IGNORES) ./...

fmt-check: fmt
	@echo "==> checking code format"
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "code is not properly formatted."; \
		echo "==> git status --porcelain"; \
		git status --porcelain; \
		echo "==> git diff"; \
		git diff; \
		exit 1; \
	fi

lint: dep
	@echo "==> linting"
	@echo ">>>golangci-lint"
	@golangci-lint run
	@go vet ./...
	@echo "done"

clean:
	@find . -name cover.out | xargs rm -rf
	@find . -name .xgo | xargs rm -rf
	@rm -rf build/*

changelog:
	@git chglog --next-tag HEAD -o CHANGELOG.md || true

pre-commit: dep fmt lint view-cover changelog{{if .HasTargets}} targets{{end}}
