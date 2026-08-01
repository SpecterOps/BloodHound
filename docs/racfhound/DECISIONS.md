# RACFHound Design Decisions

An append-only log of cross-cutting design/engineering decisions for this fork. Record a decision
here when it affects how the fork is structured, how it integrates with upstream, or how it's
maintained — anything a future contributor (or AI session) would otherwise have to re-derive.

Keep entries short. Newest first. One decision per entry:

> **YYYY-MM-DD — Title**
> **Decision:** what we chose.
> **Why:** the reason / problem it solves.
> **Alternatives:** what we rejected and why (optional).
> **Refs:** files, PRs, or docs.

Per-feature UI/behavior docs live beside this file in `docs/racfhound/`; the maintenance workflow
lives in [`MAINTAINING.md`](MAINTAINING.md). This log is for the "why", not the "how".

---

> **2026-08-01 — Pin the toolchain in fork-only files, not `package.json` engines**
> **Decision:** Pin Node (`.nvmrc`, `.tool-versions`) and Go (`.tool-versions`, matches `go.mod`);
> let Corepack manage yarn via the existing `packageManager` field. Do **not** add an `engines` block
> to the root `package.json`.
> **Why:** Consistent toolchain across machines without editing an upstream-owned file. The root
> `package.json` comes from SpecterOps; adding `engines` there creates merge surface on every sync.
> **Alternatives:** `engines` guardrail in `package.json` (active enforcement) — rejected to avoid
> upstream conflicts, consistent with the fork's "touch zero shared files" preference.
> **Refs:** `.nvmrc`, `.tool-versions`, `docs/racfhound/MAINTAINING.md`.

> **2026-07-31 — Object-info panel: inject RACF sections as `priorityTables`**
> **Decision:** The app-local `GraphItemInformationPanel` wrapper injects RACF relationship sections
> via `priorityTables`, not `additionalTables`.
> **Why:** `EntityInfoContent` only renders `additionalTables` for built-in kinds; custom/OpenGraph
> kinds like RACF are routed to `KindInfoItems`, which drops them — the sections silently vanish.
> `priorityTables` render unconditionally.
> **Alternatives:** Editing shared `EntityInfoContent` to render `additionalTables` for custom kinds —
> rejected as it touches an actively-refactored upstream file. Durable end-state: upstream a proper
> relationship-table slot for custom kinds.
> **Refs:** `cmd/ui/src/views/Explore/GraphItemInformationPanel.tsx`, `racfAdditionalTables.tsx`, and
> the two guarding tests; PR #6.

> **2026-07-27 — App-local wrapper for the object-information panel**
> **Decision:** Keep an app-local `GraphItemInformationPanel` wrapper that handles RACF nodes and
> delegates everything else to the shared upstream panel; `GraphView.tsx` imports the local wrapper.
> **Why:** Upstream moved the panel into `bh-shared-ui`, which cannot import app-local RACF
> components. The wrapper keeps all RACF UI in the app layer and touches zero shared-library files.
> **Refs:** PR #6; `docs/racfhound/MAINTAINING.md` (panel pattern).

> **2026-07-27 — Always merge upstream on a branch, never directly into `racf-main`**
> **Decision:** Upstream syncs happen on a `merge/*` branch, validated, then PR'd into `racf-main`.
> `main` stays a clean mirror of SpecterOps.
> **Why:** Keeps `racf-main` reviewable and revertible; isolates conflict resolution.
> **Refs:** PRs #6 and #8; `docs/racfhound/MAINTAINING.md` (syncing with upstream).

> **2026-07-27 — Guard the CLA Assistant workflow to the upstream org**
> **Decision:** The `cla.yml` CLA Assistant workflow should be guarded with
> `if: github.repository_owner == 'SpecterOps'` (or disabled on the fork).
> **Why:** On the fork the owner is a user, not an org, so `GET /orgs/<owner>/members` 404s and the
> `jq` step dies (`Cannot index string with string "login"`). The CLA secrets don't exist on the fork
> either. The check has no meaning on a personal fork.
> **Refs:** `.github/workflows/cla.yml`; `docs/racfhound/MAINTAINING.md` (CI section).
