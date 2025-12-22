#!/bin/sh
set -e

echo "==> ReadWillBe Development Environment"
echo "==> Installing dependencies..."

go install github.com/air-verse/air@latest
go install github.com/a-h/templ/cmd/templ@latest
bun add -D daisyui@latest

go mod download

echo "==> Generating templ files..."
templ generate

echo "==> Building initial CSS..."
tailwindcss -i ./input.css -o ./static/css/style.css

# Create tmp directory for air
mkdir -p tmp

echo "==> Starting tailwindcss --watch..."
tailwindcss -i ./input.css -o ./static/css/style.css --watch &
TAILWIND_PID=$!

cleanup() {
    echo "==> Shutting down..."
    kill $TAILWIND_PID 2>/dev/null || true
    exit 0
}
trap cleanup INT TERM

echo "==> Starting air (Go hot reload)..."
echo "==> Server will be available at http://localhost:8080"
air

# Wait for cleanup
wait
