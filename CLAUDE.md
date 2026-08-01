# CLAUDE.md

This is the **RACFHound fork** of SpecterOps BloodHound — a RACF-aware analysis layer and UI on top
of BloodHound CE. This file is a router to the project's docs; read the relevant one for your task
rather than duplicating it here.

## Where to look

- **Code / PR / test standards → [`AGENTS.md`](AGENTS.md)** (upstream, authoritative). Go style,
  `just prepare-for-codereview` before PRs, PR title format, integration-test build tags, etc.
  Follow it for all code you write.
- **Fork maintenance → [`docs/racfhound/MAINTAINING.md`](docs/racfhound/MAINTAINING.md).** How to sync
  with SpecterOps upstream, conflict hotspots, the object-panel injection pattern, the dev environment,
  validation commands, and Windows/CI gotchas. **Read this before any upstream merge or dev-env work.**
- **RACF project intent & architecture → [`LLM_INSTRUCTIONS_RACFHOUND.md`](LLM_INSTRUCTIONS_RACFHOUND.md).**
  Goals, repo/branch strategy, RACF data model and analysis approach.
- **RACF panel behavior → [`docs/racfhound/`](docs/racfhound/)** (per-panel docs).
- **Design decisions (the "why") → [`docs/racfhound/DECISIONS.md`](docs/racfhound/DECISIONS.md).**
  Append a short entry here when you make a cross-cutting design/integration/maintenance decision.

## Always-on fork rules

- **Branch model:** `main` mirrors SpecterOps; `racf-main` is the RACF integration branch. RACF work
  lands on `racf-main` via focused `feature/*` branches. **Never merge upstream directly into
  `racf-main`** — always merge on a branch and PR it in.
- **After any merge, run `yarn install`** — a stale `node_modules` throws confusing "Cannot find
  module" errors in *upstream* files, not yours.
- **Rebuild the `bh-ui` container when JS deps change** (its `node_modules` is baked into the image,
  not bind-mounted): `docker compose --profile dev -f docker-compose.dev.yml up -d --build bh-ui`.
- **RACF object-panel sections must be injected as `priorityTables`, not `additionalTables`** — custom
  kinds bypass the `additionalTables` path. See MAINTAINING.md for why, and the regression tests that
  guard it.
- Windows-native test failures (path separators, timezone formatting, dir `fsync`, temp-file locking)
  are environment artifacts, not bugs — the real suites run in Linux containers.
- **Toolchain is pinned** in `.tool-versions` (mise/asdf: Node 22, Go 1.26.4) and `.nvmrc`; yarn is
  managed by Corepack (`packageManager` in `package.json`). Use these versions on every machine.
