# Stage 1: Go Base & Tool Builder
FROM golang:1.25-alpine@sha256:d9b2e14101f27ec8d09674cd01186798d227bb0daec90e032aeb1cd22ac0f029 AS gobase
RUN apk add --no-cache git
# Install Air and Templ
RUN go install github.com/air-verse/air@latest && \
    go install github.com/a-h/templ/cmd/templ@latest

# Stage 2: Development Environment
# Uses custom frontend image (Bun + Tailwind) as base
FROM ghcr.io/jwhumphries/frontend:latest@sha256:a67b0171a111f70cb0ca54d7fdee6f6ac6382ea8b3151460fc8281faede61832 AS dev
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