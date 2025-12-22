FROM ghcr.io/jwhumphries/frontend:latest AS frontend
WORKDIR /workdir
RUN [ "bun",  "add", "-D", "daisyui@latest" ]
COPY ./static/css/style.css /workdir/static/css/style.css

FROM frontend AS css-compile
RUN [ "tailwindcss", "-i", "./static/css/style.css", "-o", "./static/css/style.min.css", "--minify"]

FROM golang:1.25-alpine AS gomods
ENV GOCACHE=/go-build-cache
ENV GOMODCACHE=/go-mod-cache
ENV CGO_ENABLED=0
RUN apk add --no-cache git ca-certificates
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
   go install github.com/a-h/templ/cmd/templ@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
   go mod download

FROM gomods AS templer
COPY static ./static
COPY views ./views
RUN templ generate

FROM templer AS builder
COPY cmd ./cmd
COPY types ./types
COPY version ./version
COPY --from=css-compile /workdir/static/css/style.min.css ./static/css/style.min.css
ARG VERSION="dev"
RUN --mount=type=cache,target=/go-build-cache --mount=type=cache,target=/go-mod-cache \
  go build -ldflags "-X readwillbe/version.Tag=$VERSION" -o /readwillbe ./cmd/readwillbe/
RUN echo "nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin" > /etc/passwd

FROM alpine AS release
RUN apk add --no-cache tzdata
ENV TZ=America/New_York
COPY --from=builder /readwillbe /readwillbe
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
USER 10001
ENV PORT=:8080
EXPOSE 8080
ENTRYPOINT ["/readwillbe"]
