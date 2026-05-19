# Style-Guides Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring ReadWillBe into conformance with `/Users/john/code/git/style-guides/` for Go linting, TypeScript strictness, JS formatting/linting, CI runner, and project documentation — without regressing behavior.

**Architecture:** Seven independent PRs, each landed green on `main`, each producing a usable repo state. Tooling-first: each PR introduces canonical config from the style-guides repo, then fixes the findings that result, then commits. The Dagger build pipeline remains the source of truth for CI; the only externally-visible change is `task` → `just` as the developer-facing wrapper.

**Tech Stack:** Go 1.26, Echo, Templ, SQLite (gorm), React 19 + TypeScript 6, esbuild, Tailwind CSS v4 + DaisyUI 5, Dagger v0.20.1, Bun, just, GitHub Actions.

**Reference:** Style guide repo is at `/Users/john/code/git/style-guides/`. Canonical files referenced throughout this plan:

- `/Users/john/code/git/style-guides/.golangci.yml`
- `/Users/john/code/git/style-guides/.prettierrc.json`
- `/Users/john/code/git/style-guides/eslint.config.js`
- `/Users/john/code/git/style-guides/tsconfig.base.json` (reference only — we copy fields inline)
- `/Users/john/code/git/style-guides/ci/README.md` (justfile + Dagger pattern)

---

## Decisions made up-front

These choices were settled in the planning conversation and are not revisited inside the plan:

1. **Task runner:** Switch to `just` (full conformance). Delete `Taskfile.yml` once `justfile` is verified.
2. **golangci.yml:** Adopt the style-guide file verbatim. Fix every new finding rather than re-introducing exclusions.
3. **tsconfig.json:** Don't `extends` the base. Copy the strict flags inline. Keep browser-bundle settings (`jsx: react-jsx`, `module: esnext`, `moduleResolution: bundler`, `lib: [dom, dom.iterable, esnext]`).
4. **Prettier:** Adopt config and reformat the whole tree.
5. **ESLint:** Adopt config, wire into Dagger and `just lint`, fix all findings.
6. **CI structure:** Keep the parallel `Check` Dagger function. Document the deviation from the style-guide example.
7. **Docs consolidation:** Deferred to PR 7.

---

## Phase / PR Map

| PR  | Title                          | Touches                                                                   | Risk               |
| --- | ------------------------------ | ------------------------------------------------------------------------- | ------------------ |
| 1   | Replace Task with just         | `justfile`, delete `Taskfile.yml`, dev script, GH workflow refs, README   | Low                |
| 2   | Adopt style-guide golangci.yml | `.golangci.yml`, `.dagger/.golangci-lint-ignore`, Dagger `Fmt`, Go source | Medium (lint debt) |
| 3   | TypeScript strict-family flags | `tsconfig.json`, any TS files that surface errors                         | Low                |
| 4   | Prettier                       | `.prettierrc.json`, `package.json`, Dagger `PrettierCheck`, reformat      | Low (mechanical)   |
| 5   | ESLint                         | `eslint.config.js`, `package.json`, Dagger `Eslint`, TS source fixes      | Medium (rule debt) |
| 6   | Dagger / CI tidy-ups           | `dagger.json`, `.dagger/main.go`, optional workflow tweaks                | Low                |
| 7   | Documentation consolidation    | `CLAUDE.md`, `AGENTS.md`, `README.md`, `docs/`                            | Low                |

Each PR ends with `dagger -m .dagger call check --source=.` green and CI green.

---

## Pre-flight (do once, before PR 1)

- [ ] **Step P1: Confirm tooling installed**

Run: `which just dagger bun docker`
Expected: paths for all four. If `just` is missing on macOS: `brew install just`.

- [ ] **Step P2: Confirm clean working tree on `main`**

Run: `git -C /Users/john/code/git/ReadWillBe status --short && git -C /Users/john/code/git/ReadWillBe log -1 --oneline`
Expected: empty status, recent main commit.

- [ ] **Step P3: Baseline current Dagger Check passes**

Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call check --source=.`
Expected: "All checks passed". If it fails, stop and fix `main` first — every PR in this plan assumes a green baseline.

---

## PR 1 — Replace Task with just

**Branch:** `chore/just-runner`

**Files:**

- Create: `/Users/john/code/git/ReadWillBe/justfile`
- Delete: `/Users/john/code/git/ReadWillBe/Taskfile.yml`
- Modify: `/Users/john/code/git/ReadWillBe/README.md` (any `task` references)
- Modify: `/Users/john/code/git/ReadWillBe/.github/workflows/*.yml` (only if any reference `task` — most call dagger directly)
- Inspect: `/Users/john/code/git/ReadWillBe/scripts/` and `/Users/john/code/git/ReadWillBe/tools/` for references to `task`

`CLAUDE.md` and `AGENTS.md` are intentionally left for PR 7.

### Task 1.1: Create justfile mirroring the current Taskfile.yml

- [ ] **Step 1.1.1: Audit current Taskfile recipes that must survive**

Run: `grep -E '^\s+[a-z][a-z-]*:' /Users/john/code/git/ReadWillBe/Taskfile.yml`
Expected: list of recipes. Verify it matches: `default`, `clean`, `build-dev` (internal), `dev`, `build`, `fmt`, `templ-fmt`, `lint`, `test`, `build-assets`, `typecheck`.

The `docker:*` and `clean:bun` recipes are `internal: true` — internal-only helpers, not user-facing. They don't need a 1:1 mapping in `justfile`; fold them inline where used.

- [ ] **Step 1.1.2: Write justfile**

Create `/Users/john/code/git/ReadWillBe/justfile` with these contents (note: use TAB indentation inside recipes per just convention, but `just` also accepts 4-space — use 4-space here for editor compatibility):

```just
# ReadWillBe justfile - thin wrapper around Dagger
# Every build/test/lint runs inside the Dagger module at .dagger/

set shell := ["bash", "-uc"]

APP_NAME := "readwillbe"
DEV_IMAGE := APP_NAME + ":dev"
GIT_COMMIT := `git rev-parse --short HEAD`

# List available recipes
default:
    @just --list

# Build the dev Docker image (used by `just dev`)
_build-dev:
    docker build \
        --target dev \
        -t {{DEV_IMAGE}} \
        --build-arg APP_NAME={{APP_NAME}} \
        --build-arg VERSION={{GIT_COMMIT}} \
        .

# Start dev environment with hot reload at http://localhost:7331
dev: _build-dev
    exec docker run --rm -it \
        --name {{APP_NAME}}-dev \
        -p 8080:8080 -p 7331:7331 \
        -v $(pwd):/app \
        -v go-mod-cache:/go/pkg/mod \
        -v go-build-cache:/go-build-cache \
        -e READWILLBE_PORT=:8080 \
        -e READWILLBE_DB_PATH=/app/data/readwillbe.db \
        -e READWILLBE_COOKIE_SECRET=dev-only-local-secret-min-32-chars \
        -e READWILLBE_SEED_DB=true \
        -e READWILLBE_ALLOW_SIGNUP=true \
        -e READWILLBE_LOG_LEVEL=debug \
        -e TEMPL_EXPERIMENT=rawgo \
        -e READWILLBE_HOSTNAME=localhost:7331 \
        {{DEV_IMAGE}} \
        sh -c "bun install && /develop.sh"

# Run linter (Go via golangci-lint)
lint:
    dagger -m .dagger call lint --source=.

# Run Go tests
test:
    dagger -m .dagger call test --source=.

# Run TypeScript type checking
typecheck:
    dagger -m .dagger call typecheck --source=.

# Run lint + typecheck + test in parallel
check:
    dagger -m .dagger call check --source=.

# Compile CSS (Tailwind) and React/TypeScript
build-assets:
    dagger -m .dagger call build-assets --source=. export --path=./static

# Build production Docker image
build:
    dagger -m .dagger call release --source=. --version dev-release export --path ./readwillbe-dev.tar
    id=$(docker load -i ./readwillbe-dev.tar | sed -n 's/^Loaded image.*: //p') && docker tag $id {{APP_NAME}}:latest
    rm ./readwillbe-dev.tar

# Format Go files (gofmt; switches to goimports in PR 2)
fmt:
    for file in $(gofmt -s -l . | grep -v "^vendor/" | grep -v "^.dagger/"); do go fmt "$file"; done

# Format templ files
templ-fmt:
    for file in $(find ./internal/views -type f -name '*.templ'); do templ fmt "$file"; done

# Remove build artifacts and node_modules
clean:
    rm -rf ./tmp ./bin ./node_modules
    rm -f bun.lock
```

- [ ] **Step 1.1.3: Verify justfile parses**

Run: `cd /Users/john/code/git/ReadWillBe && just --list`
Expected: list of recipes (`build`, `build-assets`, `check`, `clean`, `default`, `dev`, `fmt`, `lint`, `templ-fmt`, `test`, `typecheck`). Recipes starting with `_` are hidden.

- [ ] **Step 1.1.4: Verify lint still works through just**

Run: `cd /Users/john/code/git/ReadWillBe && just lint`
Expected: same output as previously running `task lint`. Should complete with no findings.

- [ ] **Step 1.1.5: Verify test still works through just**

Run: `cd /Users/john/code/git/ReadWillBe && just test`
Expected: tests pass.

- [ ] **Step 1.1.6: Verify check works through just**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

### Task 1.2: Audit non-Taskfile references to `task`

- [ ] **Step 1.2.1: Grep the repo for `task` references**

Run: `cd /Users/john/code/git/ReadWillBe && grep -rEn '\btask\b' --include='*.md' --include='*.yml' --include='*.yaml' --include='*.sh' --include='Dockerfile' --include='*.json' . | grep -v node_modules | grep -v static | grep -v .git/ | grep -v .dagger/internal`
Expected: a list of hits. Triage:

- Files in `CLAUDE.md` / `AGENTS.md` — leave for PR 7 (note in commit message).
- `README.md` — fix in this PR.
- `.github/workflows/*.yml` — should already use `dagger` directly, but verify.
- `scripts/*` — fix if any.
- `Dockerfile` — likely no references; skip if not present.

- [ ] **Step 1.2.2: Update README.md**

Read `/Users/john/code/git/ReadWillBe/README.md` and replace `task <recipe>` references with `just <recipe>`. For any reference that doesn't have a corresponding `just` recipe, update to use `dagger -m .dagger call <fn> --source=.` directly.

- [ ] **Step 1.2.3: Verify no workflow files break**

Run: `cd /Users/john/code/git/ReadWillBe && grep -nE '\btask\b' .github/workflows/*.yml || echo "no task references in workflows"`
Expected: "no task references in workflows" (workflows call `dagger` directly).

### Task 1.3: Remove Taskfile.yml and verify

- [ ] **Step 1.3.1: Delete Taskfile.yml**

Run: `rm /Users/john/code/git/ReadWillBe/Taskfile.yml`

- [ ] **Step 1.3.2: Re-run check to confirm nothing depended on Taskfile.yml**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

- [ ] **Step 1.3.3: Commit**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout -b chore/just-runner
git add justfile README.md
git rm Taskfile.yml
git commit -m "chore: replace Task with just as developer wrapper

Aligns with the style-guides ci/ pattern (justfile + Dagger).
Removed Taskfile.yml; all recipes preserved in justfile.
CLAUDE.md/AGENTS.md updates deferred to a later PR."
```

- [ ] **Step 1.3.4: Push and open PR**

```bash
git push -u origin chore/just-runner
gh pr create --title "Replace Task with just as developer wrapper" --body "$(cat <<'EOF'
## Summary
- Switch from Task (Taskfile.yml) to just (justfile) per style-guides ci/ pattern
- All recipes preserved with identical Dagger behavior
- CLAUDE.md/AGENTS.md updates deferred to later PR

## Test plan
- [ ] just lint passes
- [ ] just test passes
- [ ] just check passes
- [ ] just dev brings up the dev environment at localhost:7331
- [ ] CI green on PR
EOF
)"
```

---

## PR 2 — Adopt style-guide golangci.yml

**Branch:** `chore/golangci-style-guide`

**Files:**

- Modify: `/Users/john/code/git/ReadWillBe/.golangci.yml` (verbatim replace)
- Delete: `/Users/john/code/git/ReadWillBe/.dagger/.golangci-lint-ignore` (no longer needed)
- Modify: `/Users/john/code/git/ReadWillBe/.dagger/main.go` — change `Fmt` to use `goimports`
- Modify: `/Users/john/code/git/ReadWillBe/justfile` — `fmt` recipe uses `goimports`
- Modify: Go source files in `cmd/` and `internal/` (lint findings)

### Task 2.1: Replace .golangci.yml

- [ ] **Step 2.1.1: Capture style-guide file content**

Read `/Users/john/code/git/style-guides/.golangci.yml` (already known from audit). Confirm the canonical content is:

```yaml
version: '2'

run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  default: standard
  enable:
    - revive

formatters:
  enable:
    - goimports

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

- [ ] **Step 2.1.2: Write the new .golangci.yml**

Use the Write tool to replace `/Users/john/code/git/ReadWillBe/.golangci.yml` with the above content verbatim. The repo currently has version 2, hand-picked linters, an issue cap, and many exclusions — they all go away.

- [ ] **Step 2.1.3: Verify yaml parses**

Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call lint --source=. 2>&1 | head -20`
Expected: linter runs (probably with findings); does NOT fail with a yaml parse error.

### Task 2.2: Remove .golangci-lint-ignore

- [ ] **Step 2.2.1: Confirm the file is unused**

Run: `cd /Users/john/code/git/ReadWillBe && grep -rn 'golangci-lint-ignore' --include='*.go' --include='*.yml' --include='*.yaml' --include='*.md' .`
Expected: no matches (or only matches inside the file itself).

- [ ] **Step 2.2.2: Delete it**

Run: `rm /Users/john/code/git/ReadWillBe/.dagger/.golangci-lint-ignore`

### Task 2.3: Run lint and capture findings

- [ ] **Step 2.3.1: Run lint and save findings**

Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call lint --source=. 2>&1 | tee /tmp/golangci-findings.txt`
Expected: a list of findings. Common categories to expect:

- `revive`: exported var/func/type without comment, package comment missing, var naming (Id → ID, Url → URL, etc.)
- `goimports`: import ordering and grouping
- `errcheck`: previously-excluded paths (e.g. `_test.go` errcheck) now flag

Save the output. This is the work backlog for the rest of PR 2.

- [ ] **Step 2.3.2: Triage the findings**

Open `/tmp/golangci-findings.txt`. Group by file. For each file, the engineer should be able to address all findings in one edit pass.

### Task 2.4: Fix `goimports` formatting

`goimports` is a formatter, so its findings are auto-fixable.

- [ ] **Step 2.4.1: Run goimports via Dagger Fmt (after updating it — see Task 2.5)**

This step depends on Task 2.5; do Task 2.5 first if not done.

- [ ] **Step 2.4.2: Run lint again, confirm goimports findings gone**

Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call lint --source=. 2>&1 | grep -i goimports || echo "no goimports findings"`
Expected: "no goimports findings".

### Task 2.5: Update Fmt to use goimports

- [ ] **Step 2.5.1: Modify .dagger/main.go Fmt function**

In `/Users/john/code/git/ReadWillBe/.dagger/main.go`, replace the `Fmt` function:

```go
func (m *Readwillbe) Fmt(source *dagger.Directory) *dagger.Directory {
    return dag.Container().
        From("golang:1.26-alpine").
        WithEnvVariable("GOCACHE", "/go-build-cache").
        WithEnvVariable("GOMODCACHE", "/go-mod-cache").
        WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
        WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
        WithExec([]string{"go", "install", "golang.org/x/tools/cmd/goimports@latest"}).
        WithDirectory("/app", source).
        WithWorkdir("/app").
        WithExec([]string{"sh", "-c", "goimports -w $(find . -name '*.go' -not -path './.dagger/internal/*' -not -name '*_templ.go')"}).
        Directory("/app")
}
```

- [ ] **Step 2.5.2: Update justfile fmt recipe**

In `/Users/john/code/git/ReadWillBe/justfile`, replace the `fmt` recipe body:

```just
# Format Go files (goimports)
fmt:
    dagger -m .dagger call fmt --source=. export --path .
```

- [ ] **Step 2.5.3: Run fmt and confirm files updated**

Run: `cd /Users/john/code/git/ReadWillBe && just fmt && git status --short`
Expected: a list of `.go` files modified by goimports.

- [ ] **Step 2.5.4: Inspect the diff**

Run: `cd /Users/john/code/git/ReadWillBe && git diff --stat -- '*.go' | head -20`
Expected: mostly import grouping changes. Spot-check one or two files to confirm changes are sensible (no logic edits).

- [ ] **Step 2.5.5: Commit the formatter changes**

```bash
cd /Users/john/code/git/ReadWillBe
git add -A
git commit -m "chore: apply goimports formatting"
```

### Task 2.6: Address `revive` findings, file by file

`revive` will surface findings like:

- `package-comments`: missing package comment
- `exported`: exported types/funcs/vars need comments starting with the name
- `var-naming`: `Id` → `ID`, `Url` → `URL`, `Http` → `HTTP`
- `if-return`: redundant if/return
- `unused-parameter`: parameters that should be `_`

For each finding, the fix is mechanical. Do them in batches by file.

- [ ] **Step 2.6.1: Re-run lint and re-save findings**

Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call lint --source=. 2>&1 | tee /tmp/golangci-findings.txt; echo "exit=$?"`
Expected: a list of revive findings (goimports findings should be gone). Note the exit code — non-zero means findings remain.

- [ ] **Step 2.6.2: Sort findings by file**

Run: `sort -u /tmp/golangci-findings.txt | awk -F: '{print $1}' | sort -u`
Expected: list of files with findings.

- [ ] **Step 2.6.3: Fix findings file-by-file**

For each file in the list:

1. Open the file with the Read tool.
2. Apply the suggested change verbatim per the linter output. Common patterns:

   **Missing package comment:**

   ```go
   // Package <name> <one-sentence description>.
   package <name>
   ```

   **Exported without comment:**

   ```go
   // FuncName <verb starting with the function name describing behavior>.
   func FuncName(...) ... {
   ```

   **Var naming Id → ID:**

   ```go
   // Before
   func GetUserById(id string) (*User, error)

   // After
   func GetUserByID(id string) (*User, error)
   ```

   Then ripple-rename callers — use `grep -rn` and update each.

3. After editing all findings in a file, run lint scoped to that package (faster feedback):

   Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call lint --source=. 2>&1 | grep '<file-path>:'`
   Expected: no matches for that file. (We have to run the whole module — golangci-lint inside Dagger doesn't accept a path arg cleanly with the current function signature.)

- [ ] **Step 2.6.4: Run full lint after each ~10 files fixed**

Run: `cd /Users/john/code/git/ReadWillBe && dagger -m .dagger call lint --source=. 2>&1 | tail -20`
Expected: shrinking finding count. When zero, move on.

- [ ] **Step 2.6.5: Commit fixes in logical batches**

After fixing each cohesive group (e.g., "all handlers in cmd/readwillbe", "internal/repository"), commit:

```bash
cd /Users/john/code/git/ReadWillBe
git add -A
git commit -m "fix(lint): address revive findings in <area>"
```

Aim for 3–6 commits across the file. Don't squash — readable history helps if a fix breaks something subtle.

### Task 2.7: Run tests after all lint fixes

Renames (Id → ID, etc.) may break method calls in code paths we didn't check.

- [ ] **Step 2.7.1: Run tests**

Run: `cd /Users/john/code/git/ReadWillBe && just test`
Expected: all tests pass. If any fail with `undefined: SomeFunc`, complete the rename in the caller.

- [ ] **Step 2.7.2: Run check**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

### Task 2.8: Open PR 2

- [ ] **Step 2.8.1: Push and open PR**

```bash
cd /Users/john/code/git/ReadWillBe
git push -u origin chore/golangci-style-guide
gh pr create --title "Adopt style-guides .golangci.yml verbatim" --body "$(cat <<'EOF'
## Summary
- Replace .golangci.yml with style-guides verbatim (default: standard, +revive, +goimports formatter, no issue caps)
- Remove .dagger/.golangci-lint-ignore (no longer needed)
- Switch Dagger Fmt and just fmt to use goimports
- Fix all surfaced revive findings (missing comments, var naming, etc.)
- Fix import grouping via goimports

## Test plan
- [ ] just lint passes with no findings
- [ ] just test passes
- [ ] just check passes
- [ ] CI green
EOF
)"
```

---

## PR 3 — TypeScript strict-family flags

**Branch:** `chore/tsconfig-strict`

**Files:**

- Modify: `/Users/john/code/git/ReadWillBe/tsconfig.json`
- Possibly modify: `assets/js/**/*.tsx?` (fix any errors surfaced)

### Task 3.1: Update tsconfig.json

- [ ] **Step 3.1.1: Replace tsconfig.json**

Replace `/Users/john/code/git/ReadWillBe/tsconfig.json` with:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "react-jsx",
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "allowUnreachableCode": false,
    "allowUnusedLabels": false
  },
  "include": ["assets/js/**/*"]
}
```

Notes on what's NOT brought in from `tsconfig.base.json`:

- `lib: ["ES2023"]` — replaced by DOM-flavored libs (we're targeting the browser).
- `module: "commonjs"` — kept as `esnext` (esbuild bundles ES modules).
- `noEmitOnError`, `pretty`, `sourceMap` — esbuild handles emit/sourcemap; `tsc` only typechecks.

- [ ] **Step 3.1.2: Run typecheck**

Run: `cd /Users/john/code/git/ReadWillBe && just typecheck`
Expected: probably some errors. Common findings:

- `noImplicitReturns`: functions that have a code path that doesn't return.
- `noFallthroughCasesInSwitch`: switch cases missing `break` / `return`.
- `allowUnreachableCode: false`: code after a `return` or `throw`.
- `allowUnusedLabels: false`: rare; probably none.

Save the output for fixing.

### Task 3.2: Fix typecheck findings

- [ ] **Step 3.2.1: Address each finding by editing the source**

For each error, open the file with Read, apply the fix, save. There's no general code block to put here because the fixes are file-specific.

If a switch case is intentional fallthrough, add `// fall through` comment AND mark it with `// @ts-expect-error` is wrong — use `// eslint-disable-next-line no-fallthrough` if ESLint is configured (it isn't yet — for now use an explicit `break;` or restructure).

- [ ] **Step 3.2.2: Re-run typecheck**

Run: `cd /Users/john/code/git/ReadWillBe && just typecheck`
Expected: zero errors.

- [ ] **Step 3.2.3: Run full check**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

### Task 3.3: Commit and PR

- [ ] **Step 3.3.1: Commit**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout -b chore/tsconfig-strict
git add tsconfig.json assets/js/
git commit -m "chore: adopt tsconfig strict-family flags

Adds noImplicitReturns, noFallthroughCasesInSwitch,
allowUnreachableCode: false, allowUnusedLabels: false.
Bumps target to ES2022. Browser-bundle settings (module,
moduleResolution, jsx, lib) intentionally diverge from
tsconfig.base.json."
```

- [ ] **Step 3.3.2: Push and open PR**

```bash
git push -u origin chore/tsconfig-strict
gh pr create --title "Adopt TypeScript strict-family flags" --body "$(cat <<'EOF'
## Summary
- Add noImplicitReturns, noFallthroughCasesInSwitch, allowUnreachableCode: false, allowUnusedLabels: false
- Bump target ES2020 -> ES2022
- Fix any errors surfaced by stricter flags

## Test plan
- [ ] just typecheck passes
- [ ] just check passes
- [ ] CI green
EOF
)"
```

---

## PR 4 — Prettier

**Branch:** `chore/prettier`

**Files:**

- Create: `/Users/john/code/git/ReadWillBe/.prettierrc.json`
- Create: `/Users/john/code/git/ReadWillBe/.prettierignore`
- Modify: `/Users/john/code/git/ReadWillBe/package.json`
- Modify: `/Users/john/code/git/ReadWillBe/.dagger/main.go` (add `PrettierCheck`)
- Modify: `/Users/john/code/git/ReadWillBe/justfile` (`format`, `format-check`)
- Modify: every `.tsx`, `.ts`, `.json`, `.css` file under `assets/`

### Task 4.1: Add Prettier config

- [ ] **Step 4.1.1: Copy .prettierrc.json**

Create `/Users/john/code/git/ReadWillBe/.prettierrc.json` with the verbatim style-guide contents:

```json
{
  "bracketSpacing": false,
  "singleQuote": true,
  "trailingComma": "all",
  "arrowParens": "avoid"
}
```

- [ ] **Step 4.1.2: Add .prettierignore**

Create `/Users/john/code/git/ReadWillBe/.prettierignore` with:

```
node_modules
static
data
.dagger/internal
*.gen.go
*_templ.go
charts
```

This keeps Prettier away from generated, vendored, or non-JS content.

### Task 4.2: Add Prettier to package.json

- [ ] **Step 4.2.1: Modify package.json**

In `/Users/john/code/git/ReadWillBe/package.json`:

1. Add `"prettier": "3.4.2"` (or latest 3.x; Renovate will keep current) to `devDependencies`.
2. Add scripts:

```json
"format": "prettier --write .",
"format:check": "prettier --check ."
```

The full scripts block becomes:

```json
"scripts": {
    "init": "tailwindcss -i ./input.css -o ./static/css/main.css",
    "dev": "tailwindcss -i ./input.css -o ./static/css/main.css --watch",
    "build": "tailwindcss -i ./input.css -o ./static/css/main.css --minify",
    "build:js": "node tools/build.js",
    "watch:js": "node tools/build.js --watch",
    "watch:css": "node tools/watch_css.js",
    "typecheck": "tsc --noEmit",
    "format": "prettier --write .",
    "format:check": "prettier --check ."
}
```

- [ ] **Step 4.2.2: Install (locally for verification only; production install runs in Dagger)**

Run: `cd /Users/john/code/git/ReadWillBe && bun install`
Expected: prettier 3.x added.

### Task 4.3: Add Dagger PrettierCheck

- [ ] **Step 4.3.1: Add PrettierCheck function to main.go**

In `/Users/john/code/git/ReadWillBe/.dagger/main.go`, add after `Typecheck`:

```go
func (m *Readwillbe) PrettierCheck(ctx context.Context, source *dagger.Directory) (string, error) {
    return dag.Container().
        From("ghcr.io/jwhumphries/frontend:latest").
        WithMountedCache("/root/.bun/install/cache", dag.CacheVolume("bun-cache")).
        WithDirectory("/app", source).
        WithWorkdir("/app").
        WithExec([]string{"bun", "install"}).
        WithExec([]string{"bun", "run", "format:check"}).
        Stdout(ctx)
}
```

- [ ] **Step 4.3.2: Wire PrettierCheck into the parallel Check function**

In `/Users/john/code/git/ReadWillBe/.dagger/main.go`, modify the `Check` function — add a new errgroup goroutine for prettier:

```go
func (m *Readwillbe) Check(ctx context.Context, source *dagger.Directory) (string, error) {
    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        _, err := m.Lint(ctx, source)
        return err
    })
    g.Go(func() error {
        _, err := m.Typecheck(ctx, source)
        return err
    })
    g.Go(func() error {
        _, err := m.Test(ctx, source)
        return err
    })
    g.Go(func() error {
        _, err := m.PrettierCheck(ctx, source)
        return err
    })

    if err := g.Wait(); err != nil {
        return "", fmt.Errorf("check failed: %w", err)
    }
    return "All checks passed", nil
}
```

### Task 4.4: Wire format into justfile

- [ ] **Step 4.4.1: Add format recipes**

In `/Users/john/code/git/ReadWillBe/justfile`, add after `templ-fmt`:

```just
# Format JS/TS/JSON/CSS with Prettier
format:
    bun run format

# Check Prettier formatting (used by CI)
format-check:
    dagger -m .dagger call prettier-check --source=.
```

### Task 4.5: Reformat the tree

- [ ] **Step 4.5.1: Run Prettier across the repo**

Run: `cd /Users/john/code/git/ReadWillBe && bun run format`
Expected: many files reformatted. Spot-check `assets/js/index.tsx` — should be 2-space, no bracket spacing, single quotes, no arrow parens for single params.

- [ ] **Step 4.5.2: Verify format-check now passes**

Run: `cd /Users/john/code/git/ReadWillBe && bun run format:check`
Expected: "All matched files use Prettier code style!"

- [ ] **Step 4.5.3: Run typecheck and tests to confirm no regressions**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

### Task 4.6: Commit and PR

- [ ] **Step 4.6.1: Commit the config separately from the reformat**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout -b chore/prettier
git add .prettierrc.json .prettierignore package.json bun.lock .dagger/main.go justfile
git commit -m "chore: add Prettier config and wire into Dagger Check"
```

- [ ] **Step 4.6.2: Commit the reformat as a separate commit**

```bash
cd /Users/john/code/git/ReadWillBe
git add -A
git commit -m "style: apply Prettier formatting to JS/TS/JSON/CSS"
```

Two commits make `git blame` more useful — the config change is small and reviewable; the reformat is mechanical and large.

- [ ] **Step 4.6.3: Push and open PR**

```bash
git push -u origin chore/prettier
gh pr create --title "Add Prettier and reformat tree" --body "$(cat <<'EOF'
## Summary
- Add .prettierrc.json from style-guides verbatim
- Add .prettierignore for generated/vendored content
- Wire Prettier check into Dagger Check (parallel)
- Reformat all JS/TS/JSON/CSS

Two commits: config introduction, then the mechanical reformat.

## Test plan
- [ ] just check passes (includes prettier-check)
- [ ] bun run format:check passes
- [ ] No regression in dev or typecheck
EOF
)"
```

---

## PR 5 — ESLint

**Branch:** `chore/eslint`

**Files:**

- Create: `/Users/john/code/git/ReadWillBe/eslint.config.js`
- Modify: `/Users/john/code/git/ReadWillBe/package.json` (devDeps + script)
- Modify: `/Users/john/code/git/ReadWillBe/.dagger/main.go` (`EslintCheck` + Check wiring)
- Modify: `/Users/john/code/git/ReadWillBe/justfile` (`lint-js` recipe)
- Modify: TS source files (rule fixes)

### Task 5.1: Copy eslint config and install dependencies

- [ ] **Step 5.1.1: Copy eslint.config.js verbatim**

Copy `/Users/john/code/git/style-guides/eslint.config.js` to `/Users/john/code/git/ReadWillBe/eslint.config.js` byte-for-byte.

- [ ] **Step 5.1.2: Add devDependencies to package.json**

In `/Users/john/code/git/ReadWillBe/package.json`, add to `devDependencies`:

```json
"eslint": "^9.18.0",
"@eslint/js": "^9.18.0",
"typescript-eslint": "^8.20.0",
"eslint-config-prettier": "^10.0.0",
"eslint-plugin-prettier": "^5.2.0"
```

Bump versions to current on `npm view <pkg> version` if these have moved on by the date of execution.

Add script:

```json
"lint:js": "eslint ."
```

- [ ] **Step 5.1.3: Install**

Run: `cd /Users/john/code/git/ReadWillBe && bun install`
Expected: ESLint and friends installed.

### Task 5.2: Add Dagger EslintCheck

- [ ] **Step 5.2.1: Add EslintCheck function**

In `/Users/john/code/git/ReadWillBe/.dagger/main.go`, after `PrettierCheck`:

```go
func (m *Readwillbe) EslintCheck(ctx context.Context, source *dagger.Directory) (string, error) {
    return dag.Container().
        From("ghcr.io/jwhumphries/frontend:latest").
        WithMountedCache("/root/.bun/install/cache", dag.CacheVolume("bun-cache")).
        WithDirectory("/app", source).
        WithWorkdir("/app").
        WithExec([]string{"bun", "install"}).
        WithExec([]string{"bun", "run", "lint:js"}).
        Stdout(ctx)
}
```

- [ ] **Step 5.2.2: Wire into Check**

Modify `Check` to add a fifth errgroup:

```go
g.Go(func() error {
    _, err := m.EslintCheck(ctx, source)
    return err
})
```

### Task 5.3: justfile entry

- [ ] **Step 5.3.1: Add lint-js recipe**

In `/Users/john/code/git/ReadWillBe/justfile`, replace the `lint` recipe with a Go-only `lint-go`, add `lint-js`, and update `lint` to call both:

```just
# Run Go linter
lint-go:
    dagger -m .dagger call lint --source=.

# Run JS/TS linter
lint-js:
    dagger -m .dagger call eslint-check --source=.

# Run all linters (Go + JS/TS)
lint: lint-go lint-js
```

### Task 5.4: Run ESLint and fix findings

- [ ] **Step 5.4.1: Run ESLint**

Run: `cd /Users/john/code/git/ReadWillBe && bun run lint:js 2>&1 | tee /tmp/eslint-findings.txt`
Expected: a list of findings.

Predicted finding sites (from the audit):

- `assets/js/index.tsx:19` — `Record<string, React.ComponentType<any>>` → `@typescript-eslint/no-explicit-any`
- `assets/js/index.tsx:53` — `JSON.parse(propsJson)` returns `any` → flows into props with `no-explicit-any` or `no-unsafe-*`
- `assets/js/components/DashboardReadings.tsx:67` — `document.body.dispatchEvent(new Event(...))` — fine
- Promise.then chains without `.catch` or `await` → `no-floating-promises`
- `Array<T>` vs `T[]` mismatches → `array-type` rule (`array-simple`)

- [ ] **Step 5.4.2: Fix index.tsx component registry typing**

In `/Users/john/code/git/ReadWillBe/assets/js/index.tsx`, change:

```ts
// Before
const components: Record<string, React.ComponentType<any>> = {

// After
type ReactProps = Record<string, unknown>;
const components: Record<string, React.ComponentType<ReactProps>> = {
```

And:

```ts
// Before
const props = propsJson ? JSON.parse(propsJson) : {};

// After
const props: ReactProps = propsJson
  ? (JSON.parse(propsJson) as ReactProps)
  : {};
```

This is the minimum-cost fix. A deeper refactor (per-component prop types) is out of scope.

- [ ] **Step 5.4.3: Fix any remaining no-explicit-any findings**

For each remaining `any`, prefer `unknown` (then narrow at use site) or a defined type. Address one at a time.

- [ ] **Step 5.4.4: Fix no-floating-promises findings**

For each unawaited promise, either:

- `await` it (if inside an async function),
- `void` it explicitly: `void doAsync()` — accepted by the rule and signals intent, or
- `.catch(err => ...)` it.

- [ ] **Step 5.4.5: Fix array-type findings**

The rule `array-simple` accepts `T[]` for primitives and `Array<T>` for complex types like `Array<Foo | Bar>`. ESLint will tell you which to use; apply each suggestion.

- [ ] **Step 5.4.6: Re-run ESLint until clean**

Run: `cd /Users/john/code/git/ReadWillBe && bun run lint:js`
Expected: no errors. Warnings are acceptable (`ban-ts-comment` is configured as warn).

### Task 5.5: Verify full check passes

- [ ] **Step 5.5.1: Run check**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

- [ ] **Step 5.5.2: Verify the dev server still works**

Run: `cd /Users/john/code/git/ReadWillBe && just dev`
Manually exercise: open http://localhost:7331, sign in (seeded user), confirm:

- Dashboard renders, DashboardReadings React island mounts.
- ConfirmModal works on a plan deletion or similar destructive action.
- DatePicker works on plan editor.

Ctrl-C the dev container.

### Task 5.6: Commit and PR

- [ ] **Step 5.6.1: Commit infrastructure + fixes together**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout -b chore/eslint
git add eslint.config.js package.json bun.lock .dagger/main.go justfile assets/js/
git commit -m "chore: add ESLint with style-guides flat config

Wires eslint-check into Dagger Check (parallel). Fixes
findings: replace 'any' with typed unknown in the React
registry, fix no-floating-promises, and address array-type
findings."
```

- [ ] **Step 5.6.2: Push and open PR**

```bash
git push -u origin chore/eslint
gh pr create --title "Add ESLint" --body "$(cat <<'EOF'
## Summary
- Add eslint.config.js from style-guides verbatim
- Add ESLint deps and lint:js script
- Wire ESLint into Dagger Check (parallel)
- Fix all surfaced findings (no-explicit-any in React registry, no-floating-promises, array-type)

## Test plan
- [ ] just check passes (now includes eslint)
- [ ] just dev runs; dashboard / plans / account UI still works
- [ ] CI green
EOF
)"
```

---

## PR 6 — Dagger / CI tidy-ups

**Branch:** `chore/dagger-tidy`

**Files:**

- Modify: `/Users/john/code/git/ReadWillBe/dagger.json`
- Modify: `/Users/john/code/git/ReadWillBe/.dagger/main.go`

### Task 6.1: Name the dagger module

- [ ] **Step 6.1.1: Edit dagger.json**

Replace contents of `/Users/john/code/git/ReadWillBe/dagger.json`:

```json
{
  "name": "readwillbe",
  "engineVersion": "v0.20.1",
  "sdk": {
    "source": "go"
  },
  "source": ".dagger"
}
```

- [ ] **Step 6.1.2: Regenerate dagger client (only if dagger CLI requires it)**

Run: `cd /Users/john/code/git/ReadWillBe && dagger develop`
Expected: no-op or regenerated `.dagger/dagger.gen.go`. If `.dagger/dagger.gen.go` changes, that's fine — commit it.

### Task 6.2: Decouple Build from Lint and Test

The current `Build` function calls `lintSource` and `testSource` inline. This duplicates work when run alongside `Check`. Better: `Build` only builds; `Release` chains them.

- [ ] **Step 6.2.1: Edit Build to remove inline lint/test**

In `/Users/john/code/git/ReadWillBe/.dagger/main.go`, replace `Build`:

```go
func (m *Readwillbe) Build(
    ctx context.Context,
    source *dagger.Directory,
    // +optional
    // +defaultPath="/.git"
    git *dagger.Directory,
    // +optional
    version string,
) (*dagger.Container, error) {
    if version == "" {
        v, err := m.gitVersion(ctx, git)
        if err != nil {
            return nil, fmt.Errorf("version detection failed: %w", err)
        }
        version = v
    }

    templSource := m.TemplGenerate(source)
    assetsDir := m.BuildAssets(source)
    buildSource := templSource.WithDirectory("static", assetsDir)

    return m.BuildBinary(buildSource, version), nil
}
```

- [ ] **Step 6.2.2: Remove now-unused lintSource and testSource helpers**

In the same file, change `Lint` and `Test` to inline the container creation, then delete `lintSource` and `testSource`. Or simpler: keep `lintSource`/`testSource` as private helpers — they're still called from `Lint`/`Test`. The only removal is the calls inside `Build`.

Choose: **keep the helpers** (less churn). Action: just remove the two `if _, err := m.lintSource...` / `m.testSource...` blocks from `Build`.

- [ ] **Step 6.2.3: Verify Release still gates on lint/test via Check**

`Release` should fail closed if lint/test/typecheck/eslint/prettier fail. Update `Release` to call `Check` first:

```go
func (m *Readwillbe) Release(
    ctx context.Context,
    source *dagger.Directory,
    // +optional
    // +defaultPath="/.git"
    git *dagger.Directory,
    // +optional
    version string,
) (*dagger.Container, error) {
    if _, err := m.Check(ctx, source); err != nil {
        return nil, fmt.Errorf("checks failed: %w", err)
    }

    binaryContainer, err := m.Build(ctx, source, git, version)
    if err != nil {
        return nil, err
    }
    binary := binaryContainer.File("/readwillbe")

    return dag.Container().
        From("alpine:3.23").
        WithExec([]string{"apk", "add", "--no-cache", "tzdata", "ca-certificates"}).
        WithFile("/usr/local/bin/readwillbe", binary).
        WithExec([]string{"sh", "-c", "echo 'nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin' >> /etc/passwd"}).
        WithEnvVariable("TZ", "America/New_York").
        WithEnvVariable("PORT", ":8080").
        WithExposedPort(8080).
        WithUser("10001").
        WithEntrypoint([]string{"/usr/local/bin/readwillbe"}), nil
}
```

### Task 6.3: Document the parallel Check deviation

- [ ] **Step 6.3.1: Add a comment on the Check function**

In `/Users/john/code/git/ReadWillBe/.dagger/main.go`, add a doc comment above `Check`:

```go
// Check runs lint, typecheck, test, prettier-check, and eslint-check in
// parallel within a single Dagger session.
//
// This deviates from the style-guides ci/ pattern (which uses separate
// GitHub Actions jobs per task). The parallel approach is faster locally
// and in CI because it shares one Dagger engine init and module cache.
// The trade-off is coarser-grained status reporting in the GitHub UI.
func (m *Readwillbe) Check(ctx context.Context, source *dagger.Directory) (string, error) {
```

### Task 6.4: Verify and commit

- [ ] **Step 6.4.1: Run full check**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

- [ ] **Step 6.4.2: Verify a release build works**

Run: `cd /Users/john/code/git/ReadWillBe && just build`
Expected: a Docker image is built and tagged `readwillbe:latest`. (This is slow — only run if you have time.)

- [ ] **Step 6.4.3: Commit and PR**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout -b chore/dagger-tidy
git add dagger.json .dagger/main.go .dagger/dagger.gen.go
git commit -m "chore: tidy Dagger module

- Name the module 'readwillbe' in dagger.json
- Remove duplicate lint/test from Build (Release gates via Check)
- Document the parallel Check deviation from style-guides"
git push -u origin chore/dagger-tidy
gh pr create --title "Dagger module tidy-ups" --body "$(cat <<'EOF'
## Summary
- Name the Dagger module 'readwillbe' (was empty)
- Decouple Build from lint/test (Release still gates via Check)
- Document why we deviate from style-guides ci/ structure with parallel Check

## Test plan
- [ ] just check passes
- [ ] just build produces a Docker image
- [ ] CI green
EOF
)"
```

---

## PR 7 — Documentation consolidation

**Branch:** `docs/consolidate`

**Files:**

- Modify: `/Users/john/code/git/ReadWillBe/CLAUDE.md`
- Modify: `/Users/john/code/git/ReadWillBe/AGENTS.md`
- Modify: `/Users/john/code/git/ReadWillBe/README.md`

### Task 7.1: Decide consolidation approach

The user said this decision was deferred. Two options:

1. **Single source of truth in `AGENTS.md`, point `CLAUDE.md` to it** (recommended — `AGENTS.md` is the newer cross-vendor convention).
2. **Keep both, mirror content** (more maintenance forever).

Default to option 1. If the user disagrees during PR review, revisit.

### Task 7.2: Rewrite AGENTS.md to reflect the new reality

- [ ] **Step 7.2.1: Rewrite AGENTS.md**

Replace `/Users/john/code/git/ReadWillBe/AGENTS.md` with content that accurately documents:

- Build commands use `just` (list the recipes from `just --list`).
- Architecture overview matches the current layout (`internal/views/`, `internal/model/`, `internal/repository/`, `cmd/readwillbe/`, etc.).
- Dagger parallel `Check` is the documented deviation from style-guides.
- React islands pattern (registry in `assets/js/index.tsx`, mount via `@React("Name", props)` in `.templ`).
- DaisyUI components in `internal/views/components/` are `gsi`-generated.

A draft template (fill in the gaps from the current accurate state):

````markdown
# AGENTS.md

This file provides guidance to AI Agents when working with code in this repository.

## Build & Development Commands

All builds, tests, and lints run through Dagger. `just` is the developer wrapper.

```bash
# Development
just dev              # Start dev environment with hot-reload at http://localhost:7331

# CI/Build (via Dagger)
just check            # Run lint, typecheck, test, prettier-check, eslint-check (parallel)
just lint             # Run Go and JS/TS linters
just lint-go          # Go-only (golangci-lint)
just lint-js          # JS/TS-only (ESLint)
just test             # Run Go tests
just typecheck        # TypeScript type checking
just build-assets     # Compile CSS (Tailwind) and React/TypeScript
just build            # Build production Docker image

# Formatting
just fmt              # Format Go files with goimports
just templ-fmt        # Format Templ files
just format           # Format JS/TS/JSON/CSS with Prettier
just format-check     # Verify Prettier formatting
```
````

## Architecture Overview

### Tech Stack

- Backend: Go 1.26, Echo, SQLite (go-sqlite3/gormlite)
- Frontend: Templ + React 19 islands + Tailwind CSS v4 + DaisyUI 5
- Build: Dagger CI/CD, Docker, Bun
- Runner: just
- Deployment: Kubernetes / Helm in `charts/readwillbe/`

### Request Flow

1. Echo router (`cmd/readwillbe/server.go`) handles HTTP requests
2. Handlers render Templ templates (`internal/views/*.templ`) producing HTML
3. React components mount as "islands" via `@React("Name", props)`
4. DaisyUI components in `internal/views/components/` provide reusable UI primitives (gsi-generated)

### React Islands Pattern

React components are embedded in Templ via a registry pattern. To add one:

1. Create component in `assets/js/components/`
2. Register it in `assets/js/index.tsx`
3. Mount it in a `.templ` file: `@React("ComponentName", map[string]interface{}{"prop": value})`

### Data Models

`internal/model/`:

- `User`, `Plan`, `Reading`, `PushSubscription`

### Configuration

`READWILLBE_*` environment variables via Viper. Required: `READWILLBE_COOKIE_SECRET` (32+ chars).

### Key Directories

```
cmd/readwillbe/                # Main application, HTTP handlers
cmd/dev/                       # Dev tooling
internal/views/                # Templ page templates
internal/views/components/     # DaisyUI components (gsi-generated)
internal/model/                # Domain types
internal/repository/           # Persistence layer
internal/middleware/           # Echo middleware
assets/js/                     # React + TypeScript
.dagger/                       # Dagger module (CI/build)
charts/                        # Helm chart
static/                        # Generated assets (don't edit by hand)
```

### Style guides

This repo follows `/Users/john/code/git/style-guides/`. Deviations:

- CI uses a parallel `Check` Dagger function instead of separate jobs (faster).
- `tsconfig.json` does not `extends` `tsconfig.base.json` (browser bundle target).

````

- [ ] **Step 7.2.2: Reduce CLAUDE.md to a pointer**

Replace `/Users/john/code/git/ReadWillBe/CLAUDE.md` with a one-liner pointer:

```markdown
# CLAUDE.md

See [AGENTS.md](AGENTS.md) for guidance to AI agents working in this repo.
````

- [ ] **Step 7.2.3: Update README.md command references**

Read `/Users/john/code/git/ReadWillBe/README.md`. Replace any stale references to `task` / Taskfile / paths / commands. Keep it short — a human-facing overview, not a duplicate of AGENTS.md.

### Task 7.3: Final verification

- [ ] **Step 7.3.1: Confirm all referenced commands exist**

For every `just <name>` command mentioned in AGENTS.md and README.md, verify it's listed in `just --list`.

Run: `cd /Users/john/code/git/ReadWillBe && just --list`
Expected: recipes match the docs.

- [ ] **Step 7.3.2: Confirm all referenced paths exist**

For every directory mentioned in AGENTS.md, verify with `ls`:

Run: `cd /Users/john/code/git/ReadWillBe && for d in cmd/readwillbe cmd/dev internal/views internal/views/components internal/model internal/repository internal/middleware assets/js .dagger charts static; do test -d "$d" || echo "missing: $d"; done`
Expected: no "missing" output.

- [ ] **Step 7.3.3: Final check**

Run: `cd /Users/john/code/git/ReadWillBe && just check`
Expected: "All checks passed".

### Task 7.4: Commit and PR

- [ ] **Step 7.4.1: Commit and PR**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout -b docs/consolidate
git add AGENTS.md CLAUDE.md README.md
git commit -m "docs: consolidate AGENTS.md as single source, reflect new toolchain"
git push -u origin docs/consolidate
gh pr create --title "Consolidate agent docs and refresh README" --body "$(cat <<'EOF'
## Summary
- AGENTS.md is the canonical doc; CLAUDE.md is now a one-line pointer
- All command references updated for just + Prettier + ESLint
- All path references updated to reality (internal/views, internal/model, etc.)
- Documents the parallel Check deviation from style-guides

## Test plan
- [ ] Every command referenced in AGENTS.md exists in `just --list`
- [ ] Every directory referenced in AGENTS.md exists
- [ ] just check passes
EOF
)"
```

---

## Final verification (after all PRs land)

- [ ] **Step F1: Pull main and run full check**

```bash
cd /Users/john/code/git/ReadWillBe
git checkout main
git pull
just check
```

Expected: "All checks passed".

- [ ] **Step F2: Diff config files against style-guides**

Run: `diff /Users/john/code/git/style-guides/.golangci.yml /Users/john/code/git/ReadWillBe/.golangci.yml`
Expected: no diff.

Run: `diff /Users/john/code/git/style-guides/.prettierrc.json /Users/john/code/git/ReadWillBe/.prettierrc.json`
Expected: no diff.

Run: `diff /Users/john/code/git/style-guides/eslint.config.js /Users/john/code/git/ReadWillBe/eslint.config.js`
Expected: no diff.

- [ ] **Step F3: Open new shell, dev environment cold start**

Run: `cd /Users/john/code/git/ReadWillBe && just dev`
Expected: dev container builds, server starts on :7331, app is usable. Sign in, navigate, complete a reading, edit a plan.

---

## Out of scope (explicitly)

These were considered and intentionally left for later:

- **Refactor away `React.FC`** — minor convention drift; `React.FC` works fine and the style guide is silent on it.
- **Drop the unnecessary `import React from 'react'`** — noise with `jsx: react-jsx`, but harmless.
- **Replace `tools/build.js`** with a Dagger-resident build pipeline — current dev hot-reload depends on local node tooling; not worth disturbing.
- **Convert templ files to use the gsi component library more aggressively** — the gsi components are partially adopted; full adoption is a separate effort.
- **Add Renovate config tweaks** to track ESLint/Prettier — Renovate already handles JS deps generically; only add if a specific issue appears.
- **Pre-commit hook to run `just check` or `bun run format:check`** — opt-in per developer; can be a follow-up if drift appears.

---

## Risk register

- **Lint debt in PR 2** is the largest unknown. If `revive` surfaces 200+ findings, consider splitting PR 2 into "config swap + goimports" and a second PR for revive fixes by package.
- **ESLint findings in PR 5** could surface `no-floating-promises` in places that look subtle. If a fix requires a real refactor (more than wrapping in `void` or `await`), open a follow-up issue rather than blocking PR 5.
- **Dev image** (`ghcr.io/jwhumphries/frontend:latest`) needs to have `bun` and node available — both `PrettierCheck` and `EslintCheck` rely on it. If the image lacks something, expect to add deps via `bun install` (already done in those functions) — but if a CLI is missing, the image itself needs updating.
