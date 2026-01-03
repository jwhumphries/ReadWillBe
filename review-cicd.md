# CI/CD Review - ReadWillBe

## Summary

| Severity | Count |
|----------|-------|
| Critical | 5 |
| High | 5 |
| Medium | 5 |
| Low | Multiple |

---

## 1. DAGGER PIPELINE ISSUES

### 1.1 CRITICAL: Redundant Templ Generation

**File:** `.dagger/main.go`

**Lines:** 35-36, 45-46, 68-69

**Issue:** Templ templates are generated 3 times per pipeline execution:
1. In `Lint()` (line 36)
2. In `Test()` (line 45)
3. In `Build()` main function (line 27)

**Impact:** Pipeline takes 3x longer than necessary (~15-30 seconds wasted).

**Fix:** Generate templ once and pass the result through the chain.

---

### 1.2 CRITICAL: Disabled Function Caching

**File:** `.dagger/dagger.json` (Line 14)

**Issue:** `"disableDefaultFunctionCaching": true` forces full re-execution even when inputs haven't changed.

**Impact:** Significantly slows down iterative development and CI runs.

**Fix:** Remove this line or set to `false` after ensuring proper cache key handling.

---

### 1.3 HIGH: Cache Volume Inconsistency

**File:** `.dagger/main.go` (Lines 37-51)

**Issue:**
- `Lint()` uses both `WithModuleCache()` and `WithMountedCache()` for go-mod-cache
- `Test()` only uses `WithMountedCache()`

**Impact:** Cache misses between different pipeline stages.

**Fix:** Standardize cache mounting strategy across all functions.

---

### 1.4 MEDIUM: Missing Error Context in Build Chain

**File:** `.dagger/main.go` (Lines 27-30)

**Issue:** `BuildCss()` and `TemplGenerate()` don't return errors but could fail silently.

**Impact:** Difficult debugging of build failures.

**Fix:** Add proper error returns and handling.

---

### 1.5 MEDIUM: Templ Version Not Pinned

**File:** `.dagger/main.go` (Line 76)

**Issue:** Installs `github.com/a-h/templ/cmd/templ@latest`

**Impact:** Non-deterministic builds - different templ version each run.

**Fix:** Pin to specific version (e.g., `@v0.2.778`).

---

## 2. DOCKERFILE ISSUES

### 2.1 CRITICAL: Duplicate Tool Installation

**File:** `Dockerfile` (Lines 67-73)

**Issue:** `air` and `templ` are installed twice in succession in "develop" stage:
```dockerfile
RUN go install github.com/air-verse/air@latest && \
    go install github.com/a-h/templ/cmd/templ@latest
# ... then again:
RUN go install github.com/air-verse/air@latest && \
    go install github.com/a-h/templ/cmd/templ@latest
```

**Impact:** Wasted Docker build layer and execution time.

**Fix:** Remove duplicate installation.

---

### 2.2 HIGH: Missing go mod download in dev Stage

**File:** `Dockerfile` (Lines 5-18)

**Issue:** "dev" stage doesn't call `go mod download` unlike "init" stage (line 32-33).

**Impact:** First container startup will be slow due to downloading modules.

**Fix:** Add `RUN go mod download` to dev stage.

---

### 2.3 HIGH: Missing ca-certificates in develop Stage

**File:** `Dockerfile` (Line 66)

**Issue:** Only adds `git` via apk, but not `ca-certificates`. Compare to "dev" stage (line 14) which includes both.

**Impact:** Potential TLS/HTTPS issues if dependencies require certificate verification.

**Fix:** Add `ca-certificates` to apk installation.

---

### 2.4 MEDIUM: Inconsistent Cache Mount Paths

**File:** `Dockerfile` (Lines 17, 32, 41, 46, 54, 89)

**Issue:**
- Some stages use `/${GOBUILDCACHE}` and `/${GOMODCACHE}` (with variable)
- Some use `/go-build-cache` and `/go-mod-cache` (hardcoded)

**Impact:** Cache inconsistency between stages.

**Fix:** Use consistent path references throughout.

---

### 2.5 LOW: No Health Check Configuration

**File:** `Dockerfile`

**Issue:** None of the Docker stages include HEALTHCHECK directive.

**Impact:** Production deployments without health monitoring.

**Fix:** Add `HEALTHCHECK CMD curl --fail http://localhost:8080/health || exit 1`

---

## 3. TASKFILE ISSUES

### 3.1 CRITICAL: Task Names Don't Match Documentation

**Files:** `AGENTS.md` vs `Taskfile.yml`

**Issue:**
- Documentation says: `task dev:start` and `task dev:stop`
- Taskfile defines: `dev-start` and `dev-stop` (hyphens, not colons)

**Impact:** Users following documentation will use wrong task names.

**Fix:** Update either documentation or Taskfile to match.

---

### 3.2 CRITICAL: Wrong Dependency Name

**File:** `Taskfile.yml` (Line 72)

**Issue:** `dev-start` depends on `build:dev`, but task is named `build-dev` (line 61).

**Impact:** Task will fail because dependency doesn't exist.

**Fix:** Change `deps: [build:dev]` to `deps: [build-dev]`.

---

### 3.3 MEDIUM: Incomplete Clean Task

**File:** `Taskfile.yml` (Lines 54-59)

**Issue:** `clean:css` removes `./static/css/style.css` but CSS output is `./static/css/main.css`.

**Impact:** CSS file won't be cleaned.

**Fix:** Update to correct filename `main.css`.

---

### 3.4 LOW: Missing Task for Dagger Build

**File:** `Taskfile.yml`

**Issue:** `dagger -m .dagger call build` isn't wrapped in a task. Only `lint` and `test` are wrapped.

**Impact:** Inconsistent developer experience.

**Fix:** Add `build` task that wraps Dagger build command.

---

## 4. DOCKER-COMPOSE ISSUES

### 4.1 CRITICAL: CSS-reload Read-Only Mount

**File:** `docker-compose.dev.yml` (Line 37)

**Issue:** `css-reload` mounts `./static/css:/app/static/css:ro` (read-only), but `tailwind` generates files into `./static/css`.

**Impact:** CSS changes in tailwind won't trigger app reloads.

**Fix:** Remove `:ro` flag or restructure volume mounts.

---

### 4.2 HIGH: Race Condition in Service Dependencies

**File:** `docker-compose.dev.yml` (Lines 16-20, 48-52)

**Issue:** Services depend on `service_started` but don't wait for actual readiness.

**Impact:** Initial container startup might fail due to missing CSS files.

**Fix:** Add health checks and use `service_healthy` condition.

---

### 4.3 MEDIUM: No Health Checks

**File:** `docker-compose.dev.yml` (All Services)

**Issue:** None of the services have health checks defined.

**Impact:** Flaky development environment startup.

**Fix:** Add healthcheck for each service:
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 5s
  timeout: 3s
  retries: 3
```

---

### 4.4 LOW: Hardcoded Image Tags

**File:** `docker-compose.dev.yml` (Lines 3, 23, 34, 55)

**Issue:** Uses hardcoded `readwillbe:dev` image tag with no override option.

**Impact:** Can't easily switch between development stages.

**Fix:** Use environment variable for image tag.

---

## 5. GITHUB ACTIONS ISSUES

### 5.1 HIGH: Missing Build Job

**File:** `.github/workflows/ci.yml`

**Issue:** Workflow only runs lint and test, never builds the production Docker image.

**Impact:** Production build failures won't be caught until deployment.

**Fix:** Add build job that runs `dagger call release --source=.`

---

### 5.2 MEDIUM: No Dependency Between Jobs

**File:** `.github/workflows/ci.yml` (Lines 10-42)

**Issue:** `lint` and `test` jobs run in parallel with no dependency.

**Impact:** If lint fails, tests still run unnecessarily.

**Fix:** Add `needs: lint` to test job.

---

### 5.3 LOW: Missing Permissions Declaration

**File:** `.github/workflows/ci.yml`

**Issue:** No explicit GITHUB_TOKEN permissions defined.

**Impact:** Uses default permissions which might be overly permissive.

**Fix:** Add `permissions` field with minimal required scopes.

---

### 5.4 LOW: No Artifact Upload

**File:** `.github/workflows/ci.yml`

**Issue:** Test results aren't uploaded as artifacts.

**Impact:** Harder to debug CI failures.

**Fix:** Add `actions/upload-artifact` step for test results.

---

## 6. BUILD PROCESS ISSUES

### 6.1 HIGH: Unversioned External Image Dependency

**Files:** `Dockerfile` (Line 3), `docker-compose.dev.yml` (Line 23)

**Issue:** All pipelines depend on `ghcr.io/jwhumphries/frontend:latest` with no version pinning.

**Impact:**
- `latest` tag could change unexpectedly
- If image becomes unavailable, entire pipeline breaks

**Fix:** Pin to specific digest or version tag.

---

### 6.2 MEDIUM: Go Version Hardcoded in Multiple Places

**Files:**
- `Dockerfile`: `golang:1.25-alpine`
- `go.mod`: `go 1.25`
- `.dagger/go.mod`: `go 1.25`

**Issue:** Updating Go version requires changes in 3 places.

**Fix:** Use build argument in Dockerfile and standardize version management.

---

### 6.3 LOW: Missing Lock File for Frontend Dependencies

**File:** Project root

**Issue:** `package.json` exists but no `bun.lock` is committed.

**Impact:** Non-deterministic CSS output.

**Fix:** Commit bun.lock file.

---

## 7. PIPELINE ORDER ISSUES

### 7.1 MEDIUM: Lint Before Generate is Backwards

**File:** `.dagger/main.go` (Lines 19-25)

**Issue:** Lint runs on original source before templ is generated.

**Impact:** Linting might miss errors in generated code.

**Fix:** Generate templ first, then lint the generated code.

---

### 7.2 LOW: BuildCss Happens After Tests

**File:** `.dagger/main.go` (Line 28)

**Issue:** CSS is built after tests run.

**Impact:** If CSS build fails, tests still passed (false positive).

**Fix:** Build CSS earlier in pipeline or add CSS validation step.

---

## 8. SPEED & EFFICIENCY

### 8.1 HIGH: Sequential Stages Could Be Parallel

**Issue:** Lint and Test both depend on templ generation but run sequentially.

**Impact:** Pipeline serializes work that could be parallelized.

**Fix:** Run Lint and Test in parallel after templ generation.

---

### 8.2 MEDIUM: Frontend Image latest Tag Caching

**Issue:** Using `latest` tag means Docker can't cache layers properly.

**Impact:** Slow first build on new CI agents.

**Fix:** Pin to specific version/digest.

---

### 8.3 MEDIUM: Go Module Cache Not Pre-populated

**Issue:** docker-compose app service runs `go mod download` at startup.

**Impact:** Slow initial dev environment startup.

**Fix:** Pre-download modules in Docker build stage.

---

## 9. PORTABILITY & EASE OF USE

### 9.1 HIGH: Complex Multi-Tool Setup

**Issue:** Requires: Docker, Docker Compose, Dagger, Task, Go, Node/Bun.

**Impact:** Steep learning curve for new contributors.

**Fix:** Create setup script and document prerequisites clearly.

---

### 9.2 MEDIUM: Dagger Requires Manual Module Knowledge

**Issue:** Users must understand `dagger -m .dagger call` syntax.

**Impact:** Inconsistent interface for developers.

**Fix:** Wrap all Dagger commands in Taskfile tasks.

---

### 9.3 LOW: Missing Prerequisites Check

**Issue:** No validation that required tools are installed.

**Impact:** Setup friction for new developers.

**Fix:** Add `check-deps` task that validates tool installation.

---

## 10. ERROR HANDLING

### 10.1 MEDIUM: Silent Failures in BuildCss

**File:** `.dagger/main.go` (Lines 58-66)

**Issue:** Returns `*dagger.Directory` but could fail silently.

**Impact:** Build continues with broken CSS.

**Fix:** Add proper error handling and validation.

---

### 10.2 LOW: Missing .dockerignore Content

**File:** `.dockerignore` (Lines 1-10)

**Issue:** Minimal patterns, doesn't exclude `.git`.

**Impact:** Docker build context includes unnecessary files.

**Fix:** Add `.git`, `node_modules`, `tmp`, etc.

---

## Priority Fix Order

**Immediate (Critical):**
1. Fix redundant templ generation in Dagger (3x execution)
2. Fix duplicate tool install in Dockerfile
3. Fix wrong task dependency name in Taskfile
4. Fix task names to match documentation
5. Fix CSS-reload read-only mount in docker-compose

**High Priority:**
1. Enable Dagger function caching (remove disableDefaultFunctionCaching)
2. Add go mod download to dev stage
3. Add ca-certificates to develop stage
4. Add build job to GitHub Actions
5. Pin external image versions

**Medium Priority:**
1. Standardize cache mount paths
2. Add health checks to docker-compose services
3. Fix clean task to remove correct CSS file
4. Add error handling to Dagger functions
5. Run Lint and Test in parallel

**Low Priority:**
1. Add Dagger build task to Taskfile
2. Add prerequisites check task
3. Pin templ version in Dagger
4. Add artifact upload to GitHub Actions
5. Update .dockerignore
