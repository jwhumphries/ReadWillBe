# Stage 1: Go Base & Tool Builder
FROM golang:1.27-alpine@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523 AS gobase
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
# Install Air and Templ (templ pinned to the go.mod version so generated code matches the runtime)
RUN go install github.com/air-verse/air@latest && \
    go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{.Version}}' github.com/a-h/templ)

# Stage 2: Development Environment
# Uses custom frontend image (Bun + Tailwind) as base
FROM ghcr.io/jwhumphries/frontend:latest@sha256:7270e8fbe7f4629906764604b12b77182e3714474ccee5b14aa01ccc6b8a5b53 AS dev
WORKDIR /app

# Install system dependencies (git/curl needed for dev tools)
RUN apk add --no-cache git curl ca-certificates

# Copy Go installation from the official Go image
COPY --from=gobase /usr/local/go /usr/local/go

# Copy pre-built Go tools (Air, Templ)
COPY --from=gobase /go/bin/air /usr/local/bin/air
COPY --from=gobase /go/bin/templ /usr/local/bin/templ

# Setup Go Environment Variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Dev environment config
ENV APP_NAME=readwillbe
ENV GOCACHE=/go-build-cache
ENV GOMODCACHE=/go-mod-cache

# Copy startup script
COPY scripts/develop.sh /develop.sh
RUN chmod +x /develop.sh

EXPOSE 8080 7331

CMD ["/develop.sh"]