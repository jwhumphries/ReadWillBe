#!/bin/sh
set -e

echo "==> ReadWillBe Development Environment"
echo "==> Installing dependencies..."

bun install

go mod download

echo "==> Building initial WASM..."
mkdir -p ./static/js
cp $(go env GOROOT)/lib/wasm/wasm_exec.js ./static/js/wasm_exec.js
GOOS=js GOARCH=wasm go build -o ./web/app.wasm ./web

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
