# ReadWillBe justfile - thin wrapper around Dagger
# Every build/test/lint runs inside the Dagger module at .dagger/

set shell := ["bash", "-uc"]

APP_NAME := "readwillbe"
DEV_IMAGE := APP_NAME + ":dev"
GIT_COMMIT := `git rev-parse --short HEAD`

# List available recipes
default:
    @just --list

# Build the dev Docker image (used by `just dev`)
_build-dev:
    docker build \
        --target dev \
        -t {{DEV_IMAGE}} \
        --build-arg APP_NAME={{APP_NAME}} \
        --build-arg VERSION={{GIT_COMMIT}} \
        .

# Start dev environment with hot reload at http://localhost:7331
dev: _build-dev
    exec docker run --rm -it \
        --name {{APP_NAME}}-dev \
        -p 8080:8080 -p 7331:7331 \
        -v $(pwd):/app \
        -v go-mod-cache:/go/pkg/mod \
        -v go-build-cache:/go-build-cache \
        -e READWILLBE_PORT=:8080 \
        -e READWILLBE_DB_PATH=/app/data/readwillbe.db \
        -e READWILLBE_COOKIE_SECRET=dev-only-local-secret-min-32-chars \
        -e READWILLBE_SEED_DB=true \
        -e READWILLBE_ALLOW_SIGNUP=true \
        -e READWILLBE_LOG_LEVEL=debug \
        -e TEMPL_EXPERIMENT=rawgo \
        -e READWILLBE_HOSTNAME=localhost:7331 \
        {{DEV_IMAGE}} \
        sh -c "bun install && /develop.sh"

# Run linter (Go via golangci-lint)
lint:
    dagger -m .dagger call lint --source=.

# Run Go tests
test:
    dagger -m .dagger call test --source=.

# Run TypeScript type checking
typecheck:
    dagger -m .dagger call typecheck --source=.

# Run lint + typecheck + test in parallel
check:
    dagger -m .dagger call check --source=.

# Compile CSS (Tailwind) and React/TypeScript
build-assets:
    dagger -m .dagger call build-assets --source=. export --path=./static

# Build production Docker image
build:
    dagger -m .dagger call release --source=. --version dev-release export --path ./readwillbe-dev.tar
    trap 'rm -f ./readwillbe-dev.tar' EXIT; \
        id=$(docker load -i ./readwillbe-dev.tar | sed -n 's/^Loaded image.*: //p'); \
        docker tag "$id" {{APP_NAME}}:latest

# Format Go files (goimports)
fmt:
    dagger -m .dagger call fmt --source=. export --path .

# Format templ files
templ-fmt:
    for file in $(find ./internal/views -type f -name '*.templ'); do templ fmt "$file"; done

# Format JS/TS/JSON/CSS with Prettier
format:
    bun run format

# Check Prettier formatting (used by CI)
format-check:
    dagger -m .dagger call prettier-check --source=.

# Remove build artifacts and node_modules
clean:
    rm -rf ./tmp ./bin ./node_modules
    rm -f bun.lock
