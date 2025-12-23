# AGENTS.md

## Project Overview

ReadWillBe is a web application for managing reading plans. It allows users to upload CSVs of readings, track their progress, and view their history.

### Tech Stack

*   **Language:** Go (1.23+)
*   **Web Framework:** Echo v4
*   **Database:** SQLite (with `gorm` and `ncruces/go-sqlite3`)
*   **Templating:** templ (`github.com/a-h/templ`)
*   **Frontend:** HTMX, TailwindCSS (DaisyUI)
*   **Session Management:** `gorilla/sessions`
*   **Testing:** Go standard library `testing` + `testify`

### Key Patterns

1.  **HTMX & Server-Side Rendering:**
    *   The application uses HTMX for dynamic interactions (SPA-like feel) while relying on server-side rendering with `templ`.
    *   Handlers often return partial HTML (components) or use `HX-Redirect`.

2.  **Database Management:**
    *   GORM is used for ORM.
    *   **Connection Pooling:** Explicitly configured in `main.go` to handle SQLite concurrency (WAL mode is enabled by the driver usually, but connection limits are set to avoid locks).
    *   **User Middleware:** `UserMiddleware` handles session validation and user retrieval. It uses a **User Cache** (`UserCache` in `cmd/readwillbe/user_cache.go`) to minimize database hits for the logged-in user.

3.  **Background Processing:**
    *   **Plan Creation:** Creating a plan from a CSV is handled asynchronously.
        *   The `createPlan` handler creates a `Plan` record immediately with status `processing`.
        *   A goroutine parses the CSV and inserts `Reading` records.
        *   The frontend polls (or the user refreshes) to see the updated status (`active` or `failed`).
        *   `types.Plan` has `Status` and `ErrorMessage` fields to support this state machine.

4.  **Templating:**
    *   `.templ` files in `views/` define the UI.
    *   Run `templ generate` after modifying any `.templ` file.

### Development & Maintenance

*   **Running:** `go run ./cmd/readwillbe`
*   **Testing:** `go test ./...`
*   **Configuration:** `.env` file (see `.env.example` or code for keys like `DB_PATH`, `COOKIE_SECRET`).

### Robustness & Performance

*   **User Cache:** A simple in-memory cache with TTL prevents thundering herds on the users table.
*   **Async Import:** Prevents request timeouts on large CSV uploads.
*   **Error Handling:** All errors should be wrapped (using `pkg/errors`) for context before logging/returning.
*   **Logging:** `logrus` is used for structured logging.

### Future Improvements

*   Implement a proper job queue (e.g., Redis-backed) if background tasks become more complex.
*   Add more comprehensive integration tests.
*   Add metrics/monitoring.
