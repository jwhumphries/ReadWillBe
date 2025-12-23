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
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY --chmod=755 scripts/develop.sh /develop.sh
EXPOSE 8080
ENTRYPOINT ["/develop.sh"]

FROM frontend AS styler
WORKDIR /app
COPY package.json ./
RUN bun install
COPY input.css ./input.css
RUN mkdir -p static/css
RUN bun run build

FROM golangci/golangci-lint:v2.7.2 AS lint
WORKDIR /app
COPY . /app
RUN unformatted=$(find . -name "*.go" ! -name "*_templ.go" -type f -exec gofmt -l {} \;) && \
    if [ -n "$unformatted" ]; then \
        echo "The following files need formatting:"; \
        echo "$unformatted"; \
        exit 1; \
    fi
RUN golangci-lint run

FROM gobase AS gomods
ENV GOCACHE=/go-build-cache
ENV GOMODCACHE=/go-mod-cache
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
    go mod download

FROM gomods AS templer
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
   go install github.com/a-h/templ/cmd/templ@latest
COPY static ./static
COPY views ./views
RUN templ generate

FROM templer AS builder
COPY cmd ./cmd
COPY types ./types
COPY version ./version
COPY --from=styler /app/static/css/main.css ./static/css/main.css
ARG VERSION="dev"
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
  go build -ldflags "-X readwillbe/version.Tag=$VERSION" -o /readwillbe ./cmd/readwillbe/
RUN echo "nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin" > /etc/passwd

FROM alpine:3.20 AS release
WORKDIR /app
RUN apk add --no-cache tzdata
COPY --from=builder /readwillbe /readwillbe
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
ENV TZ=America/New_York
ENV PORT=:8080
EXPOSE 8080
USER 10001
ENTRYPOINT ["/readwillbe"]
