# syntax=docker/dockerfile:1

FROM golang:{{.GoVersion}}-alpine AS builder

# build env
ENV CGO_ENABLED={{.CgoEnabled}} \
    GOFLAGS=-buildvcs=false{{if .GoProxy}} \
    GOPROXY={{.GoProxy}}{{end}}

RUN apk add --no-cache make git

WORKDIR /go/src
COPY ./ ./

# Trust copied VCS metadata when repo ownership differs from the build user.
RUN git config --global --add safe.directory /go/src

RUN make target_{{.Name}}

# bundle
FROM {{.Runtime}} AS bundle
WORKDIR /app
RUN --mount=type=bind,from=builder,source=/go/src/dist/{{.Name}},target=/mnt/dist <<'EOF'
set -eux
cp "/mnt/dist/{{.Name}}" /app/app
if [ -f /mnt/dist/version ]; then
	cp /mnt/dist/version /app/version
fi
if [ -d /mnt/dist/config ]; then
	cp -a /mnt/dist/config /app/config
fi
EOF

# runtime
FROM {{.Runtime}}
RUN apk add --no-cache ca-certificates{{if .TimeZone}}
# timezone
ENV TZ={{.TimeZone}}
RUN apk add --no-cache tzdata
RUN ln -sf /usr/share/zoneinfo/{{.TimeZone}} /etc/localtime
RUN echo "{{.TimeZone}}" > /etc/timezone
{{end}}
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
{{if .Expose}}EXPOSE {{.Expose}}
{{end}}ENTRYPOINT ["/app/app"]
