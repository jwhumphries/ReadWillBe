# Docker Configuration

This guide explains how to configure and run the ReadWillBe container using Docker.

## Prerequisites

- Docker installed
- A generated cookie secret (minimum 32 characters)

## Configuration Methods

You can configure the application using environment variables or a configuration file.

### Method 1: Environment Variables (Recommended)

Environment variables must be prefixed with `READWILLBE_`.

| Variable | Description | Default | Required |
|----------|-------------|---------|:--------:|
| `READWILLBE_COOKIE_SECRET` | Secret for session cookies (min 32 chars) | - | Yes |
| `READWILLBE_PORT` | Port to listen on | `3000` | No |
| `READWILLBE_DB_PATH` | Path to SQLite database | `./tmp/readwillbe.db` | No |
| `READWILLBE_LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | `info` | No |
| `READWILLBE_ALLOW_SIGNUP` | Allow new user registration | `false` | No |
| `READWILLBE_SEED_DB` | Seed database with initial data | `false` | No |
| `READWILLBE_HOSTNAME` | Public hostname (e.g. `https://read.example.com`) | - | No |
| `TZ` | Timezone (e.g., `America/New_York`) | - | No |

#### Email Configuration (Optional)

Set `READWILLBE_EMAIL_PROVIDER` to `smtp` or `resend`.

**SMTP:**
- `READWILLBE_SMTP_HOST`
- `READWILLBE_SMTP_PORT` (Default: 587)
- `READWILLBE_SMTP_USERNAME`
- `READWILLBE_SMTP_PASSWORD`
- `READWILLBE_SMTP_FROM`
- `READWILLBE_SMTP_TLS` (`starttls`, `tls`, `none`)

**Resend:**
- `READWILLBE_RESEND_API_KEY`
- `READWILLBE_RESEND_FROM`

#### Example `docker run`

```bash
# Generate a secret
SECRET=$(openssl rand -base64 32)

docker run -d \
  -p 8080:8080 \
  -e READWILLBE_PORT=":8080" \
  -e READWILLBE_COOKIE_SECRET="$SECRET" \
  -e READWILLBE_ALLOW_SIGNUP="true" \
  -v $(pwd)/data:/data \
  -e READWILLBE_DB_PATH="/data/readwillbe.db" \
  readwillbe:latest
```

### Method 2: Configuration File

You can mount a `readwillbe.yaml` file into the container.

1.  **Create `readwillbe.yaml`**:
    ```yaml
    port: ":8080"
    log_level: "info"
    db_path: "/data/readwillbe.db"
    cookie_secret: "YOUR_GENERATED_SECRET_HERE"
    allow_signup: true
    ```

2.  **Run with volume mount**:
    ```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/readwillbe.yaml:/app/readwillbe.yaml \
  -v $(pwd)/data:/data \
  readwillbe:latest
```
