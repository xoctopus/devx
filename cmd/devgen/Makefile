
# git repository info
META_TZ := Asia/Shanghai
MODULE_PATH := $(shell cat ../../go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
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

NAME         := $(shell basename $(CURDIR))
DIST         := $(abspath ../../dist/${NAME})
DIST_NAME    := dist/${NAME}
VERSION_PATH := main

LDFLAGS="-s -w -X ${VERSION_PATH}.Name=${NAME} \
	-X ${VERSION_PATH}.Feature=${GIT_BRANCH}  \
	-X ${VERSION_PATH}.Version=${GIT_TAG}     \
	-X ${VERSION_PATH}.CommitID=${GIT_COMMIT} \
	-X ${VERSION_PATH}.CommitAt=${GIT_COMMIT_AT} \
	-X ${VERSION_PATH}.BuildAt=${BUILD_AT} "

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

OUT := ${NAME}-${GOOS}-${GOARCH}
ifeq ($(GOOS),"windows")
OUT := ${OUT}.exe
endif

export GOFLAGS := $(GOFLAGS) -buildvcs=false
ifneq ($(wildcard ../../vendor/modules.txt),)
GO_MOD_FLAG := -mod=vendor
export GOFLAGS := $(GOFLAGS) -mod=vendor
else
GO_MOD_FLAG :=
endif

build:
	@echo "==> building $(NAME)..."
	@go build $(GO_MOD_FLAG) -ldflags ${LDFLAGS} -o ${NAME}
	@cp ${NAME} ${OUT}
	@echo DONE

install: build
	@echo "==> installing to ${DIST_NAME}"
	@rm -rf ${DIST} && mkdir -p ${DIST} && mv ${NAME} ${OUT} ${DIST}
	@echo "$(NAME):$(GIT_BRANCH)@$(GIT_TAG)#$(GIT_COMMIT)[commit=$(GIT_COMMIT_AT)|build=$(BUILD_AT)]" > ${DIST}/version
	@echo "$(MODULE_PATH)" >> ${DIST}/version
	@if [ -d "config" ]; then \
		cp -r config ${DIST}; \
	fi
	@echo DONE

