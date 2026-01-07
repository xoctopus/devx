NAME=devgen
VERSION_PATH=main

LDFLAGS="-s -w -X ${VERSION_PATH}.Name=${NAME} \
	-X ${VERSION_PATH}.Branch=${GIT_BRANCH}   \
	-X ${VERSION_PATH}.Version=${GIT_TAG}      \
	-X ${VERSION_PATH}.CommitID=${GIT_COMMIT}  \
	-X ${VERSION_PATH}.BuildTime=${BUILD_AT} "

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

OUT := ${NAME}-${GOOS}-${GOARCH}
ifeq ($(GOOS),"windows")
OUT := ${OUT}.exe
endif

build:
	@echo building $(NAME)...
	@go build -ldflags ${LDFLAGS} -o ${NAME}
	@echo "$(NAME):$(GIT_BRANCH)@$(GIT_TAG)#$(GIT_COMMIT)_$(BUILD_AT)" > version
	@cp ${NAME} ${OUT}
	@echo DONE

install: build
	@echo install to dist...
	@mkdir -pv ../../dist && mv ${NAME} ${OUT} ../../dist
	@echo DONE
