FROM golang:1.25-alpine AS gobase
FROM ghcr.io/jwhumphries/frontend:latest AS frontend

FROM frontend AS develop
ARG APP_NAME=readwillbe
ENV APP_NAME=${APP_NAME}
WORKDIR /app
COPY --from=gobase /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
COPY --chmod=755 scripts/develop.sh /develop.sh
EXPOSE 8080
ENTRYPOINT ["/develop.sh"]

FROM frontend AS css-compile
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install
COPY input.css ./input.css
COPY web/ ./web/
RUN mkdir -p static/css
RUN bunx tailwindcss -i ./input.css -o ./static/css/style.min.css --minify

FROM gobase AS gomods
ENV GOCACHE=/go-build-cache
ENV GOMODCACHE=/go-mod-cache
ENV CGO_ENABLED=0
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    go mod download

FROM gomods AS wasm-builder
COPY api ./api
COPY types ./types
COPY web ./web
RUN mkdir -p web
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    GOOS=js GOARCH=wasm go build -ldflags "-s -w" -o ./web/app.wasm ./web

FROM gomods AS server-builder
COPY cmd ./cmd
COPY api ./api
COPY types ./types
COPY version ./version
COPY --from=css-compile /app/static/css/style.min.css ./static/css/style.min.css
ARG VERSION="dev"
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    go build -ldflags "-X readwillbe/version.Tag=${VERSION}" -o /app/bin/server ./cmd/server/
RUN echo "nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin" > /etc_passwd

FROM alpine:3.20 AS release
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=server-builder /app/bin/server ./server
COPY --from=server-builder /app/static ./static
COPY --from=wasm-builder /app/web/app.wasm ./web/app.wasm
COPY --from=gobase /usr/local/go/lib/wasm/wasm_exec.js ./static/js/wasm_exec.js
COPY --from=server-builder /etc_passwd /etc/passwd
ENV TZ=America/New_York
ENV PORT=:8080
EXPOSE 8080
USER 10001
ENTRYPOINT ["./server"]
