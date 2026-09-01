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

IMAGE_NAME ?= $(NAME)
IMAGE_TAG  ?= $(if $(GIT_TAG),$(GIT_TAG),latest)-$(GIT_COMMIT)

image:
	@echo "==> building image $(IMAGE_NAME):$(IMAGE_TAG)"
	@docker build -f Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) ../..
	@echo "DONE $(IMAGE_NAME):$(IMAGE_TAG)"

