FROM golang:1.27.0-alpine AS builder

# build env
ENV CGO_ENABLED=0 \
    GOFLAGS=-buildvcs=false

RUN apk add --no-cache make git

WORKDIR /go/src
COPY ./ ./

# Trust copied VCS metadata when repo ownership differs from the build user.
RUN git config --global --add safe.directory /go/src

RUN make target_demo

# Stage runtime artifacts with stable paths (fail build if binary missing).
RUN set -eux; \
	test -x "dist/demo/demo"; \
	install -D "dist/demo/demo" /out/app; \
	install -D "dist/demo/version" /out/version; \
	if [ -d "dist/demo/config" ]; then cp -a "dist/demo/config" /out/config; fi

FROM alpine:latest
RUN apk add --no-cache ca-certificates libcap
# timezone
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" > /etc/timezone

# permission
# -H: no /home/app; point HOME at /app so runtime (and any go tool cache) can write.
RUN adduser -D -H -u 65532 -s /sbin/nologin app
ENV HOME=/app

WORKDIR /app
COPY --from=builder /out/ /tmp/stage/
RUN set -eux; \
	test -x /tmp/stage/app; \
	mv /tmp/stage/app /app/app; \
	mv /tmp/stage/version /app/version; \
	if [ -d /tmp/stage/config ]; then mv /tmp/stage/config /app/config; fi; \
	rm -rf /tmp/stage; \
	chown -R app:app /app; \
	setcap 'cap_net_bind_service=+ep' /app/app; \
	apk del libcap

USER app
EXPOSE 80
ENTRYPOINT ["/app/app"]
