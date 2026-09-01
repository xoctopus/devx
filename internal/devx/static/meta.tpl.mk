{{define "meta"}}
# git repository info
META_TZ := Asia/Shanghai
MODULE_PATH := $(shell cat {{.GoModDir}}go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
IS_GIT_REPO := $(shell git rev-parse --is-inside-work-tree >/dev/null 2>&1 && echo 1 || echo 0)
ifeq ($(IS_GIT_REPO),1)
export GIT_COMMIT_RAW := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
export GIT_COMMIT_AT  := $(shell TZ=$(META_TZ) git log -1 --format=%cd --date=format:%Y%m%d%H%M%SCST 2>/dev/null || echo "")
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
export BUILD_AT := $(shell TZ=$(META_TZ) date "+%Y%m%d%H%M%SCST")
export MODULE_PATH
{{end}}
