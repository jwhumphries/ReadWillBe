# Stage 1: Go Base & Tool Builder
FROM golang:1.26-alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS gobase
RUN apk add --no-cache git
# Install Air and Templ
RUN go install github.com/air-verse/air@latest && \
    go install github.com/a-h/templ/cmd/templ@latest

# Stage 2: Development Environment
# Uses custom frontend image (Bun + Tailwind) as base
FROM ghcr.io/jwhumphries/frontend:latest@sha256:bda0acd76fc710b9e7813f510aeb5cf13726ea4f7bde7ac551b36f0f95079527 AS dev
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