# ReadWillBe

A GOTH stack application for tracking progress through daily reading plans.
Heavily inspired by [fanks](https://github.com/oliverisaac/fanks).

## Features

- User registration and authentication
- Upload reading plans via CSV files
- Dashboard view showing today's readings and overdue items
- History view of completed readings
- Manage multiple reading plans
- Configure notification settings
- Support for day, week, and month-based reading schedules

### CSV Format Example

```csv
date,reading
2025-10-15,Read Genesis 1-3
2025-10-16,Read Genesis 4-6
2025-10-17,Read Genesis 7-9
January 2025,Read Mary Had A Little Lamb
2025-W42,Read Oliver Isaac's Blog
```

Supported date formats:
- **Day**: `2025-10-15` or `10/15/2025`
- **Month**: `January 2025` or `Jan 2025` or `2025-01`
- **Week**: `2025-W42`

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

## Environment Variables

- `DB_PATH` - SQLite database path (default: `./tmp/readwillbe.db`)
- `COOKIE_SECRET` - Secret for session cookies (required)
- `ALLOW_SIGNUP` - Enable user registration (default: `true`)
- `PORT` - Server port (default: `8080`)
- `TZ` - Timezone (default: `America/Chicago`)

