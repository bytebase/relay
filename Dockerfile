FROM golang:1.19-alpine as builder

ARG VERSION="development"
ARG GIT_COMMIT="unknown"
ARG BUILD_TIME="unknown"
ARG BUILD_USER="unknown"

# See https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.revision=${GIT_COMMIT}
LABEL org.opencontainers.image.created=${BUILD_TIME}
LABEL org.opencontainers.image.authors=${BUILD_USER}

WORKDIR /app

COPY . .

# -ldflags="-w -s" means omit DWARF symbol table and the symbol table and debug information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o relay \
    main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/relay /usr/bin/

ENTRYPOINT ["relay"]
