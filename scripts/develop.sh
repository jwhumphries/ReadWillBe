#!/bin/sh
set -e

echo "Starting development environment (with dev-runner)..."

# Create necessary directories
mkdir -p data tmp

# Run the Go-based dev runner
# This handles Air, Templ, JS, CSS, and the Proxy
go run ./cmd/dev