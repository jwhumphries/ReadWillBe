# Quick Start Guide

## Get Started in 3 Steps

### 1. Install Dependencies

```bash
# Install Go tools
make install-tools

# Download Go modules
make deps
```

### 2. Configure Environment

The `.env` file is already created. Update the `COOKIE_SECRET` for production:

```env
DB_PATH=./tmp/readwillbe.db
COOKIE_SECRET=your-secure-random-secret-here
ALLOW_SIGNUP=true
PORT=8080
TZ=America/Chicago
```

### 3. Run the Application

**Option A: Docker + Hot Reload (Recommended)**
```bash
make dev
```

**Option B: Local Development**
```bash
make run
```

Visit http://localhost:8080

## Creating Your First Reading Plan

1. Sign up for an account at http://localhost:8080/auth/sign-up
2. Navigate to "Plans" in the navigation menu
3. Click "Create New Plan"
4. Upload a CSV file with your reading schedule

### CSV Format Example

Create a file named `reading-plan.csv`:

```csv
date,reading
2025-10-15,Read Genesis 1-3
2025-10-16,Read Genesis 4-6
2025-10-17,Read Genesis 7-9
January 2025,Complete Book Report
2025-W42,Weekly Review
```

Supported date formats:
- **Day**: `2025-10-15` or `10/15/2025`
- **Month**: `January 2025` or `Jan 2025` or `2025-01`
- **Week**: `2025-W42`

## Common Makefile Commands

```bash
make dev            # Run with Docker + hot reload
make run            # Run locally without Docker
make build          # Build production binary
make test           # Run all tests
make templ-generate # Generate Templ templates
make tailwind-dev   # Compile TailwindCSS
```

## Project Features

✅ User authentication (sign up/sign in)
✅ CSV upload for reading plans
✅ Dashboard with today's readings
✅ Overdue reading indicators
✅ Reading history
✅ Plan management (rename/delete)
✅ Account settings
✅ Responsive design with JetBrains Mono font

## Tech Stack

- **Backend**: Go with Echo framework
- **Frontend**: Templ templates + HTMX
- **Styling**: TailwindCSS
- **Database**: SQLite with GORM
- **Dev Tools**: Air (hot reload), Docker

## Troubleshooting

**Port already in use:**
```bash
# Change PORT in .env file
PORT=3000
```

**Docker issues:**
```bash
# Kill existing container
docker kill readwillbe
# Rebuild
make docker-build
```

**Templ generation errors:**
```bash
# Reinstall templ
go install github.com/a-h/templ/cmd/templ@latest
make templ-generate
```
