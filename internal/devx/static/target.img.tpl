# syntax=docker/dockerfile:1

FROM golang:{{.GoVersion}}-alpine AS builder

# build env
ENV CGO_ENABLED={{.CgoEnabled}}{{if .GoProxy}} \
    GOPROXY={{.GoProxy}}{{end}}

RUN apk add --no-cache make git

WORKDIR /go/src
COPY ./ ./

RUN make target_{{.Name}}

FROM {{.Runtime}} AS bundle
WORKDIR /app
COPY --from=builder /go/src/dist/{{.Name}}/{{.Name}} /app/app
{{if .Config}}COPY --from=builder /go/src/dist/{{.Name}}/config   /app/config
{{end}}

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
{{if .Expose}}EXPOSE {{.Expose}}
{{end}}ENTRYPOINT ["/app/app"]
