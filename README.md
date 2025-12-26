# ReadWillBe

A GOTH stack application for tracking progress through daily reading plans.
Heavily inspired by [fanks](https://github.com/oliverisaac/fanks).

## Features

- User registration and authentication
- Upload reading plans via CSV files
- Dashboard view showing today's readings and overdue items
- History view of completed readings
- Manage multiple reading plans
- Configure notification settings (time-based alerts)
- Browser push notifications for daily reading reminders
- In-app notification bell with real-time updates
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

## Setup

### 1. Generate VAPID Keys for Push Notifications

Browser push notifications require VAPID (Voluntary Application Server Identification) keys. Generate them using:

```bash
go run github.com/SherClockHolmes/webpush-go/cmd/vapid-keygen@latest
```

This will output:
```
Public Key: <your-public-key>
Private Key: <your-private-key>
```

Add these keys to your `.env` file or environment variables:
```bash
READWILLBE_VAPID_PUBLIC_KEY=<your-public-key>
READWILLBE_VAPID_PRIVATE_KEY=<your-private-key>
READWILLBE_HOSTNAME=yourdomain.com  # or localhost:7331 for development
```

**Note**:
- VAPID keys should be generated once and kept consistent across deployments
- Push notifications require HTTPS in production (localhost works for development)
- Without VAPID keys, the app will run but browser push notifications will be disabled

### 2. Create Notification Icon Assets

Create the following icon files in the `static/` directory:
- `static/icon-192.png` - 192x192px notification icon
- `static/badge-128.png` - 128x128px notification badge

Or use placeholder images during development.

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

### Required
- `READWILLBE_COOKIE_SECRET` - Secret for session cookies (required)

### Optional
- `READWILLBE_DB_PATH` - SQLite database path (default: `./tmp/readwillbe.db`)
- `READWILLBE_ALLOW_SIGNUP` - Enable user registration (default: `true`)
- `READWILLBE_PORT` - Server port (default: `8080`)
- `READWILLBE_SEED_DB` - Seed database with sample data (default: `false`)

### Push Notifications (Optional)
- `READWILLBE_VAPID_PUBLIC_KEY` - VAPID public key for push notifications
- `READWILLBE_VAPID_PRIVATE_KEY` - VAPID private key for push notifications
- `READWILLBE_HOSTNAME` - Hostname for notification URLs (e.g., `yourdomain.com`)

If VAPID keys are not provided, the app will run normally but browser push notifications will be unavailable. Users can still use in-app notifications and set notification preferences.

