# ReadWillBe

A GOTH stack (Go + Templ + HTMX + TailwindCSS) application for tracking progress through daily reading plans.

## Features

- User registration and authentication
- Upload reading plans via CSV files
- Dashboard view showing today's readings and overdue items
- History view of completed readings
- Manage multiple reading plans
- Configure notification settings
- Support for day, week, and month-based reading schedules

## Tech Stack

- **Go** - Backend server and API
- **Templ** - Type-safe HTML templating
- **HTMX** - Dynamic frontend interactions
- **TailwindCSS** - Styling with JetBrains Mono font
- **SQLite** - Database via GORM
- **Echo** - Web framework
- **Docker** - Containerized deployment
- **Air** - Hot reloading in development

## Prerequisites

- Go 1.24+
- Docker (optional, for containerized development)
- Air (for hot reloading): `go install github.com/air-verse/air@latest`
- Templ: `go install github.com/a-h/templ/cmd/templ@latest`
- TailwindCSS CLI

## Quick Start

1. Clone the repository
2. Copy `.env.example` to `.env` and configure:
   ```bash
   cp .env.example .env
   ```
3. Install dependencies:
   ```bash
   make deps
   make install-tools
   ```

## Development

### Option 1: Docker + Air (Recommended)
```bash
make dev
```
This will build a Docker container and run the app with hot reloading on port 8080.

### Option 2: Local Development
```bash
make run
```

### Option 3: Manual Steps
```bash
# Generate Templ templates
make templ-generate

# Compile TailwindCSS
make tailwind-dev

# Run the application
go run ./cmd/readwillbe/
```

## Building for Production

```bash
make build
```

The binary will be output to `./bin/readwillbe`.

## Testing

```bash
make test
```

## CSV Format

Upload reading plans as CSV files with the following format:

```csv
date,reading
2025-01-15,Read Chapter 1
2025-01-16,Read Chapter 2
January 2025,Read Book 1
2025-W01,Weekly Reading
```

Supported date formats:
- **Day**: `2025-01-15` or `01/15/2025`
- **Month**: `January 2025` or `Jan 2025` or `2025-01`
- **Week**: `2025-W01`

## Project Structure

```
.
├── cmd/
│   └── readwillbe/        # Application entrypoint and handlers
├── types/                 # Data models and configuration
├── views/                 # Templ templates
├── static/
│   └── css/              # TailwindCSS styles
├── version/              # Version information
├── Dockerfile            # Container definition
├── Makefile              # Build commands
└── .air.toml            # Hot reload configuration
```

## Environment Variables

- `DB_PATH` - SQLite database path (default: `./tmp/readwillbe.db`)
- `COOKIE_SECRET` - Secret for session cookies (required)
- `ALLOW_SIGNUP` - Enable user registration (default: `true`)
- `PORT` - Server port (default: `8080`)
- `TZ` - Timezone (default: `America/Chicago`)

## License

MIT
