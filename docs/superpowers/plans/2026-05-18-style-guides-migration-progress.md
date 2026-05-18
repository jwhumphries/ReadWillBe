# Style-Guides Migration — Progress Tracker

Companion to `2026-05-18-style-guides-migration.md`. Update as PRs land.

## PRs

- [x] **PR 1 — Replace Task with just** — [#163](https://github.com/jwhumphries/ReadWillBe/pull/163) (open, CI green on first commit, re-running on fix commit)
  - `justfile` mirrors all Taskfile recipes; `Taskfile.yml` deleted; README updated
  - Code review fixes applied: `build` recipe uses `trap` for tarball cleanup; README typo
- [ ] **PR 2 — Adopt style-guides `.golangci.yml`** — *biggest unknown; revive may surface many findings*
- [ ] **PR 3 — TypeScript strict-family flags**
- [ ] **PR 4 — Prettier (config + tree reformat)**
- [ ] **PR 5 — ESLint (config + fix `any`, `no-floating-promises`, `array-type`)**
- [ ] **PR 6 — Dagger / CI tidy-ups** (name module, decouple `Build`, document parallel `Check` deviation)
- [ ] **PR 7 — Documentation consolidation** (`AGENTS.md` canonical; `CLAUDE.md` pointer; **include `.jules/palette/*.md` here**)

## Deferred to PR 7 (do not forget)

- `.jules/palette/palette.md` and `.jules/palette/palette_one_shot.md` contain stale `task` references — both reviewers agreed these are agent persona/scratchpad files, parallel to CLAUDE.md/AGENTS.md.
- `AGENTS.md` and `CLAUDE.md` themselves — still stale after PR 1.

## Issues parked for follow-up (not in any PR)

*(None yet.)*

## Notes per PR

### PR 1
- All work targets `dev` (active integration branch, 59 ahead of `main`).
- Docker daemon was down on the host, so `just check` was not verified locally — relied on CI.
- The `fmt` recipe still uses `gofmt`; intentionally deferred to PR 2 where it switches to `goimports`.
