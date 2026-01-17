# ReadWillBe

A GOTH stack application for tracking progress through daily reading plans.
Heavily inspired by [fanks](https://github.com/oliverisaac/fanks).

## Features

- User registration and authentication
- Upload reading plans via CSV files
- Dashboard view showing today's readings and overdue items
- History view of completed readings
- Browser push notifications & Email reminders
- Support for day, week, and month-based reading schedules

### CSV Format Example

```csv
date,reading
2025-10-15,Read Genesis 1-3
January 2025,Read Mary Had A Little Lamb
2025-W42,Read Oliver Isaac's Blog
```

## Setup & Configuration

ReadWillBe is designed to be run via Docker or Kubernetes.

- **[Docker Configuration](docs/docker.md)**
- **[Helm Chart Configuration](docs/helm.md)**

### Required Configuration

The only strictly required configuration is the **Cookie Secret**.

| Variable | Description |
|----------|-------------|
| `READWILLBE_COOKIE_SECRET` | A 32+ character random string. |

Generate one using:
```bash
openssl rand -base64 32
```

### Optional: Push Notifications

To enable browser push notifications, generate VAPID keys:

```bash
go run github.com/SherClockHolmes/webpush-go/cmd/vapid-keygen@latest
```

Set `READWILLBE_VAPID_PUBLIC_KEY`, `READWILLBE_VAPID_PRIVATE_KEY` and `READWILLBE_HOSTNAME`.

## Development

This project uses [Task](https://taskfile.dev/) for development workflows.

### Quick Start

```bash
# Start development environment (Docker Compose + Hot Reload)
task dev-start

# Stop development environment
task dev-stop
```

### Available Tasks

| Command | Description |
|---------|-------------|
| `task dev-start` | Starts the dev environment with hot-reload at http://localhost:8080 |
| `task dev-stop` | Stops the dev environment and cleans up |
| `task test` | Runs tests using Dagger |
| `task lint` | Runs linters using Dagger |
| `task build` | Builds the production Docker image |
| `task clean` | Removes generated files and images |
| `task fmt` | Formats Go files |
| `task templ-fmt`| Formats Templ files |