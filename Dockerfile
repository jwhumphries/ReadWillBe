FROM golang:1.25-alpine AS gobase
FROM golangci/golangci-lint:v2.7.2 AS lintbase
FROM ghcr.io/jwhumphries/frontend:latest AS frontend

FROM gobase AS dev
ARG APP_NAME=readwillbe
ARG GOMODCACHE=/go-mod-cache
ARG GOBUILDCACHE=/go-build-cache
ENV GOCACHE=${GOBUILDCACHE}
ENV GOMODCACHE=${GOMODCACHE}
ENV GOLANGCI_LINT_CACHE=${GOMODCACHE}
ENV APP_NAME=${APP_NAME}
WORKDIR /app
RUN apk add --no-cache git ca-certificates
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
  go install github.com/air-verse/air@latest
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  go install github.com/a-h/templ/cmd/templ@latest

FROM gobase AS init
ARG APP_NAME=readwillbe
ARG GOMODCACHE=/go-mod-cache
ARG GOBUILDCACHE=/go-build-cache
ENV GOCACHE=${GOBUILDCACHE}
ENV GOMODCACHE=${GOMODCACHE}
ENV GOLANGCI_LINT_CACHE=${GOMODCACHE}
ENV APP_NAME=${APP_NAME}
WORKDIR /app
COPY --from=lintbase /bin/golangci-lint /bin/golangci-lint
COPY ./go.mod ./go.sum ./.golangci.yml /app/
RUN apk add --no-cache git ca-certificates
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  go mod download
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  go install github.com/a-h/templ/cmd/templ@latest
COPY cmd/ ./cmd/
COPY static/ ./static/
COPY types/ ./types/
COPY version/ ./version/
COPY views/ ./views/
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  go mod vendor

FROM init AS lint
SHELL ["/bin/sh", "-eo", "pipefail", "-c"]
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  [[ $( gofmt -s -l . | grep -v "^vendor/" | tee /dev/stderr | wc -l ) -eq 0 ]]
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  golangci-lint config verify && \
  golangci-lint run --color=always --issues-exit-code=1 --timeout=5m

FROM init AS test
SHELL ["/bin/sh", "-eo", "pipefail", "-c"]
RUN --mount=type=cache,target=/${GOBUILDCACHE} --mount=type=cache,target=/${GOMODCACHE} \
  go vet  ./... && \
  go test ./...

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
RUN go install github.com/air-verse/air@latest
RUN go install github.com/a-h/templ/cmd/templ@latest

FROM frontend AS uinit
WORKDIR /app
COPY ./package.json /app/
COPY --from=init /app/views/ ./views/
RUN bun install
COPY input.css ./input.css
RUN bun run init

FROM uinit AS minify
RUN mkdir -p static/css
RUN bun run build

FROM init AS builder
COPY --from=minify /app/static/css/main.css ./static/css/main.css
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
  templ generate
ARG VERSION="dev"
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
  go build -ldflags "-X readwillbe/version.Tag=$VERSION" -o /readwillbe ./cmd/readwillbe/
RUN echo "nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin" > /etc/passwd

FROM alpine:3.23 AS release
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
