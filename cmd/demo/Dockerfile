# syntax=docker/dockerfile:1

FROM golang:1.27.0-alpine AS builder

# build env
ENV CGO_ENABLED=0

RUN apk add --no-cache make git

WORKDIR /go/src
COPY ./ ./

RUN make target_demo

# bundle
FROM alpine:latest AS bundle
WORKDIR /app
RUN --mount=type=bind,from=builder,source=/go/src/dist/demo,target=/mnt/dist <<'EOF'
set -eux
cp "/mnt/dist/demo" /app/app
if [ -f /mnt/dist/version ]; then
	cp /mnt/dist/version /app/version
fi
if [ -d /mnt/dist/config ]; then
	cp -a /mnt/dist/config /app/config
fi
EOF

# runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
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
RUN --mount=type=bind,from=bundle,source=/app,target=/mnt/app <<'EOF'
set -eux
apk add --no-cache libcap
cp -a /mnt/app/. /app/
chown -R app:app /app
setcap 'cap_net_bind_service=+ep' /app/app
apk del libcap
EOF

USER app
EXPOSE 80
ENTRYPOINT ["/app/app"]
