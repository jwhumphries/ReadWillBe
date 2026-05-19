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
- [x] **PR 6 — Dagger / CI tidy-ups** — [#168](https://github.com/jwhumphries/ReadWillBe/pull/168) (open, `just check` + `just build` green locally)
  - `dagger.json` named `readwillbe`; `engineVersion` bumped `v0.20.1` → `v0.20.8` by `dagger develop` (no breaking SDK changes in range)
  - `Build` decoupled from inline lint/test; `Release` now gates on `Check`, so quality runs once across the pipeline (not twice when Check + Build both run)
  - Expanded `Check` doc comment with the style-guides deviation rationale (parallel checks in one Dagger session vs. separate GH Actions jobs)
  - Code-review follow-up: tightened `Release` error wrap (was `"checks failed: check failed: ..."`, now `"release blocked: pre-release checks failed: ..."`)
  - One pre-existing Prettier whitespace fix in this progress doc rolled in so the new `Release → Check` gate doesn't block CI on an unrelated issue
- [x] **PR 7 — Documentation consolidation** — [#169](https://github.com/jwhumphries/ReadWillBe/pull/169) (open, `just check` green locally)
  - `AGENTS.md` rewritten as canonical (post-migration recipes, Go 1.26, current `internal/views/` and `internal/model/` paths, three documented deviations from style-guides)
  - `CLAUDE.md` reduced to a one-line pointer at `AGENTS.md`
  - `README.md` refreshed (stale `task` → `just`; current paths/commands; stays human-facing)
  - `.jules/palette/{palette,palette_one_shot}.md` — minimal-touch fixes of stale `task` references
  - Code-review follow-ups: clarified `just lint` vs `just check` distinction; noted `templ fmt` as the documented Dagger exception; softened the absolute style-guides path to "local copy ... on the primary developer machine"

🎉 **Migration complete.** All seven PRs landed; `just check` includes lint + typecheck + test + prettier-check + eslint-check in one Dagger session; `Release` gates on `Check`.

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
- Post-merge follow-up (CodeRabbit feedback): replaced `void savedCallback.current()` in `usePolling` with `Promise.resolve(...).catch(console.error)` extracted to a local `runCallback`. Per-tick polling failures would otherwise fire as unhandled rejections with no diagnostic trail. Not applied to one-shot `void` sites (`useApi`, `DashboardReadings`, `NotificationBell`, `ReadingList`) — the per-tick failure-mode argument doesn't apply.

### PR 6
- `dagger develop` bumped `engineVersion` from `v0.20.1` → `v0.20.8` to match the local Dagger CLI. `.dagger/go.mod` / `.dagger/go.sum` regen tracked the engine change (no breaking Go SDK API changes in this range — `v0.20.6` reorganised the generated client code but consumers are unaffected). `dagger.io/dagger` resolves to a pseudo-version (`v0.20.6-0.20260415192040-7058e9313c72`) — that's what `dagger develop` produced and it locks to a specific SHA, so reproducible.
- `Build` is now pure-build (TemplGenerate → BuildAssets → BuildBinary); `Release` gates on `Check` once instead of `Build` duplicating lint/test. Net: one quality pass per pipeline, not two.
- Followed the PR 5 precedent for the recurring "pre-existing Prettier blank-line failure in this progress doc" issue — fixed in-PR so the new `Release → Check` gate didn't trip on an unrelated change.

### PR 7
- Plan's draft AGENTS.md template was a starting point only — diverged where the plan was stale (Go 1.25 → 1.26 actual; `views/`/`types/` → `internal/views/`/`internal/model/`; plan's draft `prettier-fix`/`eslint-check` recipe names never existed — the real recipes are `format`/`format-check`/`lint-js`).
- Ergonomic friction worth a follow-up some day: `just format` exports `node_modules/` to the host as a side effect of `dagger ... export --path .`. Cleaned up here with `just clean` before the final `just check`, but a cleaner Dagger pattern (e.g., exporting only the changed files) would avoid the issue.
- Documented `templ fmt` as the lone Dagger exception in AGENTS.md's host-tool ban — `just templ-fmt` invokes the templ CLI directly because it's part of the dev environment, not the Dagger image.
