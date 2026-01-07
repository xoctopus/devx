NAME=devgen
VERSION_PATH=main

LDFLAGS="-s -w -X ${VERSION_PATH}.Name=${NAME} \
	-X ${VERSION_PATH}.Branch=${GIT_BRANCH}   \
	-X ${VERSION_PATH}.Version=${GIT_TAG}      \
	-X ${VERSION_PATH}.CommitID=${GIT_COMMIT}  \
	-X ${VERSION_PATH}.BuildTime=${BUILD_AT} "

build:
	@echo building $(NAME)...
	@go build -ldflags ${LDFLAGS} -o ${NAME}
	@echo DONE


install: build
	@echo install to dist...
	@mkdir -pv ../../dist && mv devgen ../../dist
	@echo DONE

