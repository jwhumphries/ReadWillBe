# Style-Guides Migration — Progress Tracker

Companion to `2026-05-18-style-guides-migration.md`. Update as PRs land.

## PRs

- [x] **PR 1 — Replace Task with just** — [#163](https://github.com/jwhumphries/ReadWillBe/pull/163) (open, CI green on first commit, re-running on fix commit)
  - `justfile` mirrors all Taskfile recipes; `Taskfile.yml` deleted; README updated
  - Code review fixes applied: `build` recipe uses `trap` for tarball cleanup; README typo
- [x] **PR 2 — Adopt style-guides `.golangci.yml`** — [#164](https://github.com/jwhumphries/ReadWillBe/pull/164) (open, `just check` green locally)
  - `.golangci.yml` replaced verbatim with style-guides version; `.dagger/.golangci-lint-ignore` removed; `Fmt` and `just fmt` now run `goimports`
  - Revive findings fixed in-place (no `//nolint` directives) across `cmd/readwillbe`, `internal/{cache,middleware,model,repository,service,views}`, `static`
  - Two package renames driven by `var-naming` / package-name rules (see notes below)
- [x] **PR 3 — TypeScript strict-family flags** — [#165](https://github.com/jwhumphries/ReadWillBe/pull/165) (open)
- [x] **PR 4 — Prettier (config + tree reformat)** — [#166](https://github.com/jwhumphries/ReadWillBe/pull/166) (open, `just check` green locally)
  - `.prettierrc.json` + `.prettierignore` copied verbatim from style-guides; `prettier@3.4.2` added; `format` / `format:check` scripts; `PrettierCheck` wired into the parallel Dagger `Check`; `just format` + `just format-check` recipes
  - Tree-wide reformat across 38 files (JS/TS/CSS/JSON/MD/YAML); pre-existing markdown fence in `docs/docker.md` fixed so Prettier output is idempotent
  - Code-review follow-up: corrected the stale `Check` doc comment that previously claimed Check ran "build"
- [x] **PR 5 — ESLint (config + fix `any`, `no-floating-promises`, `array-type`)** — [#167](https://github.com/jwhumphries/ReadWillBe/pull/167) (open, `just check` green locally)
  - `eslint.config.js` copied verbatim from style-guides; ESLint deps pinned with carets; `lint:js` script; `EslintCheck` wired into the parallel Dagger `Check` (now 5 goroutines: lint + typecheck + test + prettier-check + eslint-check); justfile split into `lint-go` / `lint-js` with `lint` running both
  - Findings fixed in-place: `no-floating-promises` (explicit `void` on `queryClient.invalidateQueries`, `refetch`, `savedCallback.current()`), `no-case-declarations` (block-wrap `case 'week'` in `DatePicker`), `no-unused-vars` (drop unused `id` param + unused `Plan` import)
  - Code-review follow-up: corrected stale doc comment on `EslintCheck`
- [ ] **PR 6 — Dagger / CI tidy-ups** (name module, decouple `Build`, document parallel `Check` deviation)
- [ ] **PR 7 — Documentation consolidation** (`AGENTS.md` canonical; `CLAUDE.md` pointer; **include `.jules/palette/*.md` here**)

## Deferred to PR 7 (do not forget)

- `.jules/palette/palette.md` and `.jules/palette/palette_one_shot.md` contain stale `task` references — both reviewers agreed these are agent persona/scratchpad files, parallel to CLAUDE.md/AGENTS.md.
- `AGENTS.md` and `CLAUDE.md` themselves — still stale after PR 1.

## Issues parked for follow-up (not in any PR)

_(None yet.)_

## Notes per PR

### PR 1

- All work targets `dev` (active integration branch, 59 ahead of `main`).
- Docker daemon was down on the host, so `just check` was not verified locally — relied on CI.
- The `fmt` recipe still uses `gofmt`; intentionally deferred to PR 2 where it switches to `goimports`.

### PR 2

- Two non-obvious revive-driven package renames:
  - `version/` → `versioninfo/` (`package versioninfo`). The original `version` name conflicted with the stdlib `go/version` package (Go 1.22+), and the first attempted replacement (`buildinfo`) collided with stdlib `debug/buildinfo`. `versioninfo` has no stdlib conflict. `-ldflags -X readwillbe/versioninfo.Tag=...` updated in `.dagger/main.go`.
  - `internal/service/csv/` keeps its directory but the package is now `csvservice` (callers import as `csvservice "readwillbe/internal/service/csv"`). The previous `package csv` shadowed the stdlib `encoding/csv` import used inside `parser.go`, which revive's `var-naming` flagged.
- Revive surfaced ~40 findings (mostly `package-comments`, `exported`, one `unused-parameter`). All fixed in-place — no exclusions, no `//nolint` directives.
- `just check` passes locally after the renames; CI status to be confirmed on the open PR.

### PR 3

- Zero source fixes needed: `just typecheck` was already green under the four new strict-family flags and `target: ES2022`. The `assets/js/` tree is small (a handful of islands) and was already clean of implicit returns, switch fallthroughs, unreachable code, and unused labels.
- `tsconfig.json` intentionally does NOT `extends` `style-guides/tsconfig.base.json`: the repo's browser-bundle settings (`module: esnext`, `moduleResolution: bundler`, `jsx: react-jsx`, `lib: [dom, dom.iterable, esnext]`) diverge from the base. The strict flags were copied inline instead.
- `just check` green locally on the first try.

### PR 4
- Recipe names are `format` / `format-check` (per the plan), NOT `prettier-fix` / `prettier-check` as the current CLAUDE.md claims — CLAUDE.md fix is deferred to PR 7 along with AGENTS.md.
- Prettier reformatted further than `assets/`: also `.github/`, `.golangci.yml`, `.jules/**`, `AGENTS.md`, `README.md`, `docs/**`, `input.css`, `readwillbe.yaml`, `tools/**`, and even the migration plan docs themselves. Plan permitted this ("Prettier's default scope is correct").
- `docs/docker.md` had a malformed code fence inside a list item that made Prettier non-idempotent (rewrite → re-flag). Fixed structurally (re-indented the fence under its list item; no prose changes) and rolled into the `style:` reformat commit.
- Two `Minor` review findings deferred: adding `bun.lock` to `.prettierignore` (Prettier silently skips it today — no-op), and `bun install` running on every `PrettierCheck` (matches the existing `Typecheck` / `BuildAssets` pattern; intentional).
- Mid-PR additions during code review: added `PrettierFix` Dagger function and rerouted `just format` through Dagger (was `bun run format` direct), per the "Dagger-only" project rule. Plan's PR 4 spec only required `PrettierCheck`; expanding the surface here keeps the host clean of bun invocations.

### PR 5
- `lint:js` is `eslint assets/js` (NOT `eslint .`). The style-guides flat config only ignores `node_modules/`, `dist/`, `build/`, so `eslint .` would lint the committed `static/js/bundle.js` and produce thousands of vendor-code findings. Side effect: `tools/build.js` and `tools/watch_css.js` are no longer linted (tiny esbuild/Tailwind driver scripts — acceptable).
- React islands registry in `assets/js/index.tsx` keeps `Record<string, React.ComponentType<any>>` with a narrow `eslint-disable-next-line @typescript-eslint/no-explicit-any` and explanatory comment. Five registered components have required props, so the plan's prescribed `Record<string, unknown>` typing fails TS contravariance. Refactoring registered-component prop types is out of PR 5's scope.
- One pre-existing Prettier failure on `dev` (missing blank line after `### PR 4` heading in this progress doc) was rolled into the `fix(lint):` commit so `just check` would stay green. Whitespace-only.
- Refreshing `bun.lock` was done via an ephemeral `BunInstall` helper in `.dagger/main.go`, used to `export --path .`, then deleted before committing — no host-side `bun install`.
