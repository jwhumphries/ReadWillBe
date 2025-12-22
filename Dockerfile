# =============================================================================
# ReadWillBe Dockerfile
# Multi-target build pipeline for development and production
# =============================================================================

# Base Images
FROM golang:1.25-alpine AS gobase
FROM ghcr.io/jwhumphries/frontend:latest AS frontend

# =============================================================================
# Development Stage
# All-in-one development environment with hot reload
# Usage: docker run -v $(pwd):/app -p 8080:8080 readwillbe:dev
# =============================================================================
FROM frontend AS develop
ARG APP_NAME=readwillbe
ENV APP_NAME=${APP_NAME}
WORKDIR /app

# Install Go from gobase
COPY --from=gobase /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Install development tools
RUN apk add --no-cache git libc6-compat
RUN go install github.com/air-verse/air@latest
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install DaisyUI for TailwindCSS
RUN bun add -D daisyui@latest

# Copy development entrypoint script
COPY --chmod=755 scripts/develop.sh /develop.sh

EXPOSE 8080
ENTRYPOINT ["/develop.sh"]

# =============================================================================
# CSS Compilation Stage
# Compiles TailwindCSS with DaisyUI
# =============================================================================
FROM frontend AS css-compile
WORKDIR /app
RUN bun add -D daisyui@latest
COPY input.css ./input.css
COPY views/ ./views/
COPY web/ ./web/
RUN mkdir -p static/css
RUN tailwindcss -i ./input.css -o ./static/css/style.min.css --minify

# =============================================================================
# Go Modules Stage
# Downloads and caches Go dependencies
# =============================================================================
FROM gobase AS gomods
ENV GOCACHE=/go-build-cache
ENV GOMODCACHE=/go-mod-cache
ENV CGO_ENABLED=0
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    go mod download

# Install templ for code generation
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    go install github.com/a-h/templ/cmd/templ@latest

# =============================================================================
# Templ Generation Stage
# Generates Go code from .templ files
# =============================================================================
FROM gomods AS templer
COPY static ./static
COPY views ./views
RUN templ generate

# =============================================================================
# WASM Builder Stage
# Builds the WebAssembly binary for go-app frontend
# =============================================================================
FROM gomods AS wasm-builder
COPY . .
RUN mkdir -p web
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    GOOS=js GOARCH=wasm go build -o ./web/app.wasm ./web

# =============================================================================
# Server Builder Stage
# Builds the server binary
# =============================================================================
FROM templer AS server-builder
COPY cmd ./cmd
COPY api ./api
COPY types ./types
COPY version ./version
COPY --from=css-compile /app/static/css/style.min.css ./static/css/style.min.css
ARG VERSION="dev"
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    go build -ldflags "-X readwillbe/version.Tag=${VERSION}" -o /app/bin/server ./cmd/server/

# Create non-root user entry
RUN echo "nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin" > /etc_passwd

# =============================================================================
# Release Stage
# Minimal production image
# =============================================================================
FROM alpine:3.20 AS release
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

# Copy built artifacts
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
