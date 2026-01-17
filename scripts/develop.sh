#!/bin/sh
set -e

echo "Starting development environment (with dev-runner)..."

# Create necessary directories
mkdir -p data tmp

# Use exec to replace shell with Go process
# This ensures signals (SIGINT, SIGTERM) go directly to the Go dev runner
exec go run -tags dev ./cmd/dev
