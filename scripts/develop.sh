#!/bin/sh
set -e

echo "==> ReadWillBe Development Environment"
echo "==> Installing dependencies..."

# Install Go dependencies
go mod download

# Generate templ files initially
echo "==> Generating templ files..."
templ generate

# Build initial CSS
echo "==> Building initial CSS..."
tailwindcss -i ./input.css -o ./static/css/style.css

# Create tmp directory for air
mkdir -p tmp

# Start tailwindcss in watch mode (background)
echo "==> Starting tailwindcss --watch..."
tailwindcss -i ./input.css -o ./static/css/style.css --watch &
TAILWIND_PID=$!

# Trap to cleanup on exit
cleanup() {
    echo "==> Shutting down..."
    kill $TAILWIND_PID 2>/dev/null || true
    exit 0
}
trap cleanup INT TERM

# Start air for Go hot reload
echo "==> Starting air (Go hot reload)..."
echo "==> Server will be available at http://localhost:8080"
air

# Wait for cleanup
wait
