#!/bin/sh
set -e

# Cleanup function
cleanup() {
    echo "Stopping all processes..."
    kill $(jobs -p)
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "Starting development environment..."

# Create necessary directories
mkdir -p data tmp

# 1. Start Templ Proxy & Watcher
# We use --open-browser=false because we are in docker
# We proxy to port 8080 (where air/go will listen)
echo "Starting Templ..."
templ generate --watch --proxy="http://localhost:8080" --proxyport=7331 --proxybind="0.0.0.0" --open-browser=false &

# 2. Start JS Watcher
echo "Starting JS Watcher..."
bun run watch:js &

# 3. Start CSS Watcher
echo "Starting CSS Watcher..."
bun run watch:css &

# 4. Start Air (Go Hot Reload)
echo "Starting Air..."
air

# Wait for any process to exit
wait
