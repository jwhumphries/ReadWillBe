#!/bin/sh
set -e

echo "==> ReadWillBe Development Environment"
echo "==> Installing dependencies..."

bun install

go mod download

echo "==> Building initial CSS..."
bun run init

# Create tmp directory for air
mkdir -p tmp

echo "==> Starting templ generate --watch..."
templ generate --watch &
TEMPL_PID=$!

echo "==> Starting tailwindcss --watch..."
bun run dev &
TAILWIND_PID=$!

cleanup() {
    echo "==> Shutting down..."
    kill $TAILWIND_PID 2>/dev/null || true
    kill $TEMPL_PID 2>/dev/null || true
    exit 0
}
trap cleanup INT TERM

echo "==> Starting air (Go hot reload)..."
echo "==> Server will be available at http://localhost:8080"
air

# Wait for cleanup
wait
