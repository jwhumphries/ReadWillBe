# AGENTS.md

## Build & Development Commands

```bash
# Development (Docker + hot reload) - recommended
task dev-start          # Start dev environment on port 7331 (proxied)
task dev-stop           # Stop dev environment

# Testing & Linting
task test               # Run tests in Docker
task lint               # Run golangci-lint in Docker
task fmt                # Auto-format Go files
go test ./...           # Run tests locally
go test -run TestName ./cmd/readwillbe  # Run single test

# Code Generation
templ generate          # Generate Go code from .templ files
npx tailwindcss -i ./input.css -o ./static/css/main.css  # Build CSS

# Production Build
task build              # Build production Docker image
task clean              # Remove generated files and Docker images

# Dagger CI Pipeline (runs from repo root)
dagger -m .dagger call lint --source=.           # Run golangci-lint
dagger -m .dagger call test --source=.           # Run tests
dagger -m .dagger call build --source=.          # Full pipeline (templ→lint→test→css→binary)
dagger -m .dagger call release --source=.        # Build production container
dagger -m .dagger call templ-generate --source=. # Generate templ files only
dagger -m .dagger call build-css --source=.      # Build CSS only
```

## Architecture Overview

**GOTH Stack**: Go, HTMX, Templ, TailwindCSS/DaisyUI

### Directory Structure
- `cmd/readwillbe/` - Main application (handlers, server, CLI)
- `views/` - Templ templates (.templ files)
- `types/` - Data models (User, Plan, Reading, Config)
- `static/` - CSS, icons, static assets
- `.dagger/` - Dagger CI pipeline (Go SDK, uses golangci-lint module)

### Key Patterns

**Handler Pattern**:
```go
func handler(ctx echo.Context) error {
    user := ctx.Get("user").(*types.User)
    return render(ctx, http.StatusOK, views.Template(user))
}
```

**HTMX Redirects**: Use `HX-Redirect` header for navigation after form submissions.

**User Context**: `UserMiddleware` extracts user from session, caches in `UserCache` (TTL-based).

**Async CSV Import**: Plan creation runs in goroutine; status transitions: `processing` → `active`/`failed`.

### Database (SQLite + GORM)
- `User` → has many `Plan`, `PushSubscription`
- `Plan` → has many `Reading`
- `Reading` → date (day/week/month), reading description, status

## Configuration

Environment variables (prefix: `READWILLBE_`):
- `COOKIE_SECRET` (required) - Session encryption
- `PORT` - Server port (default: 8080)
- `DB_PATH` - SQLite path (default: ./tmp/readwillbe.db)
- `ALLOW_SIGNUP` - Enable registration (default: true)
- `SEED_DB` - Seed sample data (default: false)
- `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY` - Push notifications
- `HOSTNAME` - For notification URLs

Config file: `readwillbe.yaml`

## Tech Stack Reference
- Use @/Users/john/Documents/Code/lib/fanks as an example for patterns
- Use fonts/colors from @/Users/john/Documents/Code/repos/jwh (styling only, not code)

## Code Style
- Don't write comments unless absolutely necessary
- Write tests for your code
- Use `pkg/errors` for error wrapping

## DaisyUI Component Patterns

### General Principles
- Prefer DaisyUI components over custom elements
- Use semantic color names (primary, secondary, accent, neutral, base-100)
- Extract reusable icons as separate templ components

### Form Controls
```html
<div class="form-control">
  <label class="label">
    <span class="label-text">Field Label</span>
  </label>
  <input type="text" class="input input-bordered" />
</div>
```
Do NOT use `<label class="form-control">` as wrapper.

### Cards
```html
<div class="card bg-base-200 shadow-xl">
  <div class="card-body">
    <h2 class="card-title">Title</h2>
    <p>Content</p>
    <div class="card-actions">
      <button class="btn">Action</button>
    </div>
  </div>
</div>
```

### Modals
Use native HTML dialog with DaisyUI styling:
```html
<button onclick={templ.ComponentScript{Call: "modal_id.showModal()"}}>Open</button>
<dialog id="modal_id" class="modal">
  <div class="modal-box">
    <h3 class="font-bold text-lg">Title</h3>
    <div class="modal-action">
      <form method="dialog">
        <button class="btn">Cancel</button>
      </form>
      <button class="btn btn-primary">Confirm</button>
    </div>
  </div>
  <form method="dialog" class="modal-backdrop">
    <button>close</button>
  </form>
</dialog>
```

### Other Components
- **Alerts**: `alert alert-info/success/warning/error` with icon
- **Badges**: `badge badge-success/outline/soft/ghost`, sizes: xs/sm/md/lg/xl
- **Toggles**: `<input type="checkbox" class="toggle toggle-primary" />`
- **File Inputs**: Hide native input, trigger via button onclick
- **Date Inputs**: Use native `<input type="date">`

### Icon Components
Extract SVG icons as templ components in views/icons.templ:
```go
templ IconName() {
  <svg class="h-5 w-5 opacity-75" viewBox="0 0 24 24" fill="none">
    <path class="stroke-current" d="..." stroke-width="1.5"></path>
  </svg>
}
```
