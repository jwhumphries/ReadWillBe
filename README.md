# ReadWillBe

📖 An application for tracking progress through daily reading plans.
🙏 Heavily inspired by [fanks](https://github.com/oliverisaac/fanks).

## Features

- 🔐 User registration and authentication
- ⬆️ Upload reading plans via CSV files
- 📊 Dashboard view showing today's readings and overdue items
- 🕰️ History view of completed readings
- 🔔 Browser push notifications & Email reminders
- 📆 Support for day, week, and month-based reading schedules

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

The only strictly required configuration is the **Cookie Secret** 🍪.

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

This project uses [just](https://just.systems/) for development workflows. The pipeline is managed by [Dagger](https://dagger.io/) 🗡️.

### Quick Start

```bash
# Start development environment (Docker + Hot Reload) at http://localhost:7331
just dev
```

To stop the dev environment, press `Ctrl+C` in the terminal running `just dev`.

### Available Recipes

| Command | Description |
|---------|-------------|
| `just dev` | Starts the dev environment with hot-reload at http://localhost:7331 |
| `just test` | Runs tests using Dagger |
| `just lint` | Runs linters using Dagger |
| `just typecheck` | Runs TypeScript type checking using Dagger |
| `just check` | Runs lint + typecheck + test in parallel using Dagger |
| `just build` | Builds the production Docker image |
| `just build-assets` | Compiles CSS (Tailwind) and React/TypeScript |
| `just clean` | Removes generated files and `node_modules` |
| `just fmt` | Formats Go files |
| `just templ-fmt`| Formats Templ files |

### Verifying the Build

As shown above, the development environment features hot-reloading. To build a copy of the release image locally (minimized for production with no reloading), run `just build`. 

Set a cookie secret, then run:

```bash
docker run -p 8080:8080 \
-e READWILLBE_COOKIE_SECRET="$READWILLBE_COOKIE_SECRET" \
-e READWILLBE_DB_PATH="/tmp/readwillbe.db" \
-e READWILLBE_LOG_LEVEL=debug \
readwillbe:latest
```
