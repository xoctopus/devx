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

# Stage runtime artifacts with stable paths (fail build if binary missing).
RUN set -eux; \
	test -x "dist/{{.Name}}/{{.Name}}"; \
	install -D "dist/{{.Name}}/{{.Name}}" /out/app; \
	install -D "dist/{{.Name}}/version" /out/version; \
	if [ -d "dist/{{.Name}}/config" ]; then cp -a "dist/{{.Name}}/config" /out/config; fi

FROM {{.Runtime}}
RUN apk add --no-cache ca-certificates libcap{{if .TimeZone}}
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
{{if .Expose}}EXPOSE {{.Expose}}
{{end}}ENTRYPOINT ["/app/app"{{range .RunArgs}}, "{{.}}"{{end}}]
