# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS builder

# build env
ENV CGO_ENABLED=0

RUN apk add --no-cache make git

WORKDIR /go/src
COPY ./ ./

RUN make target_demo

FROM alpine:latest AS bundle
WORKDIR /app
COPY --from=builder /go/src/dist/demo/demo /app/app


# runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
# timezone
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" > /etc/timezone

# permission
RUN adduser -D -H -u 65532 -s /sbin/nologin app

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
