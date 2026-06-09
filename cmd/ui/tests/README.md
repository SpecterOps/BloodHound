# BloodHound UI Playwright Tests

This directory is the entry point for browser-driven Bloodhound UI Playwright test suites. It is organized into per-suite subdirectories. Each suite has its own Playwright config, `testMatch` pattern, and artifact subfolder under `cmd/ui/playwright/`. Vitest, Bloodhound UI's unit test framework, is configured to exclude `tests/**` from collection (`cmd/ui/vite.config.ts`), so `yarn test` will not pick up the `*.a11y.spec.ts` files; the Playwright suite only runs via `yarn test:a11y`.

## Layout

```text
tests/
├── global.setup.ts   # Playwright "setup" project: logs in once and snapshots auth storage state
├── fixtures.ts       # Shared wrapper around bh-playwright-testing's `test` that installs stubs
└── a11y/             # Accessibility regression suite (axe-core via @axe-core/playwright)
    └── *.a11y.spec.ts
```

Shared Playwright building blocks (fixtures, stubs, and auth helper) live in the `bh-playwright-testing` workspace package (`packages/javascript/bh-playwright-testing`) for reusablility. See that package's README for the full module map.

## Shared Scaffolding

### `global.setup.ts`

A Playwright **setup project** that runs once per Playwright invocation before any browser-theme project executes. It:

1. Installs any stubs
2. Delegates to `bh-playwright-testing` helpers to manage per-theme auth sessions

> About auth session management:\
> `loginAndSnapshotThemes` logs in via `/ui/login`, snapshots `storageState` to `playwright/.auth/user-light.json`, toggles dark mode (waiting for the throttled `persistedState` write to land), then snapshots `playwright/.auth/user-dark.json`. Capturing both theme snapshots from a single session avoids the parallel-login race where two setups as the same user would invalidate each other's session. Downstream projects load the snapshot that matches their theme via `authStorageStateFor(theme)`.

## Conventions

-   **Folders** - Each suite had its own folder: `cmd/ui/tests/<suite>` (e.g. `cmd/ui/tests/a11y`)
-   **Filenames** - Within each suite folder, test files use a `<name>.<suite>.spec.ts` pattern (e.g. `login.a11y.spec.ts`). Each suite's Playwright config picks them up with a `testMatch` regex.
-   **Env loading** - Suite configs may load `cmd/ui/.env` via `dotenv` at config-evaluation time. See `cmd/ui/.env.example` for example keys used by the `a11y` suite.
-   **Artifacts directory** - All Playwright output is written under `cmd/ui/playwright/`, which is gitignored. Each suite scopes its artifacts to its own subfolder (e.g. `playwright/a11y/...`).
-   **Auth bootstrapping** - Suites that need an authenticated session depend on the `setup` project and load the appropriate `storageState` rather than logging in per-test.

## Accessibility Tests (`a11y/`)

The accessibility suite runs `axe-core` scans against the live BloodHound UI through `@axe-core/playwright`. Shared scan defaults, reporting helpers, and the `makeAxeBuilder` fixture come from the `bh-playwright-testing` workspace package (`packages/javascript/bh-playwright-testing`). See that package's README for the fixture API.

### Scope

The a11y suite's goal is automated WCAG 2.x accessibility regression coverage for user-facing pages in the BloodHound UI. See [Adding A New A11y Spec](#adding-a-new-a11y-spec) for the per-spec recipe.

Each spec is a self-contained scan of one route, covering any of its in-page states. Specs are kept narrow rather than chained because axe violations are easier to triage when the failure points at a single, well-scoped DOM subtree.

### `fixtures.ts`

A wrapper around `bh-playwright-testing`'s `test` that adds a `page` fixture which installs a shared stub from `bh-playwright-testing` for managing graph data state (ensuring the "No Data" dialog is not present). `fixtures.ts` also re-exports `expect` and `expectNoAccessibilityViolations` so a spec can import everything it needs from one module.

### Running The Suite

From the BHCE root (`bhce/`) or from within `cmd/ui`:

```sh
yarn test:a11y       # clears cmd/ui/playwright folder and runs the suite
```

Running in interactive UI mode

```sh
yarn test:a11y --ui
```

The `bhce/` script delegates to the `bloodhound-ui` workspace, whose `test:a11y` script clears the playwright artifact directory (`cmd/ui/playwright`) and runs the a11y test suite as configured in `playwright.a11y.config.ts` — the clean step is baked in so every run starts with a fresh `playwright/` directory. CI-mode behavior (single worker, 1 retry, `forbidOnly`) is auto-enabled when `process.env.CI` is set, so there is no separate `:ci` script. The Playwright config (`cmd/ui/playwright.a11y.config.ts`) generates a project matrix of `chromium-light`, `chromium-dark`, `firefox-light`, `firefox-dark`, each depending on the `setup` project.

By default, the full 2x2 matrix (browsers x themes) is run, but projects may be individually specified:

```sh
# Ex. Running light and dark Chromium browser projects
yarn test:a11y --project='chromium-light' --project='chromium-dark'
```

### Required Environment Variables

Populate `cmd/ui/.env` (see `cmd/ui/.env.example`):

-   `A11Y_TEST_URL` — base URL for the UI (e.g. `http://127.0.0.1:3000`, `http://bloodhound.localhost`).
-   `A11Y_TEST_USERNAME` / `A11Y_TEST_PASSWORD` — app login credentials used by `global.setup.ts`.
-   `A11Y_TEST_SERVE` — when `true`, Playwright starts the Vite dev server itself via `yarn dev --host <host> --port <port>` derived from `A11Y_TEST_URL`. When unset or `false`, Playwright expects an already-running target at `A11Y_TEST_URL`. Used to target other environments such as `test`.

### Artifacts And Reports

Each run writes to `cmd/ui/playwright/a11y/`:

-   `results/` — Playwright `outputDir` (traces, screenshots, raw attachments).
-   `html-report/` — Playwright HTML report (browsable as-is).
-   `allure-results/` — Allure raw results (`*-result.json`). See [Viewing The Reports](#viewing-the-reports) below for how to render it.

Every assertion via `expectNoAccessibilityViolations` attaches `axe-results.json` (always) and `a11y-violations.md` (only when violations exist) to the Playwright test result, which surfaces in both the HTML and Allure reports. The specs in this suite pass `{ page }` as the third argument, which adds per-node element screenshots (`a11y-<rule>-<n>.png`, up to 5 nodes per violation; additional nodes are skipped to reduce redundancy) so each violation has a visual indicator next to its textual description.

The artifacts described above are produced and consumed locally; CI integration is a separate follow-up.

### Viewing The Reports

Local workflow: run the suite, then point a report viewer at the output. Yarn scripts are mirrored at both the BHCE root (`bhce/`) and the workspace (`cmd/ui/`); pick whichever cwd is convenient. Examples below default to `bhce/`.

When BHCE is consumed as a submodule of BHE, the BHE workspace can define its own `report:a11y` (e.g. one that aggregates BHCE + BHE allure results into a combined report). The scripts here intentionally cover only BHCE's own suite.

#### Allure Report

The Allure reporter only writes raw `*-result.json` files — viewing them requires the `allure` CLI to generate HTML.

**One-time install:**

```sh
brew install allure              # macOS, recommended (brings the JRE)
# or:
npm i -g allure-commandline      # cross-platform, requires Java on PATH
```

**Per-run workflow (from `bhce/`):**

```sh
# 1. If this step is not run first, user will see the error "does not exist"
#    as the Allure results directory will be empty.
yarn test:a11y

# 2a. Ad-hoc: build HTML to a temp dir, serve it, open browser. Ctrl+C cleans up.
yarn report:a11y

# 2b. Or, persistent HTML build at cmd/ui/playwright/a11y/allure-report (shareable, zip-friendly).
yarn report:a11y:build

# 2c. Or view simpler, default Playwright HTML reports
yarn workspace bloodhound-ui playwright show-report playwright/a11y/html-report
```

The same script names exist at the `cmd/ui` level (`yarn test:a11y`, `yarn report:a11y`, `yarn report:a11y:build`) and resolve paths relative to `cmd/ui` so they work when invoked from inside the workspace too.

The Allure report aggregates across all four browser-theme projects (`chromium-light`, `chromium-dark`, `firefox-light`, `firefox-dark`), groups violations by suite, and surfaces the same `a11y-<rule>-<n>.png` attachments as inline images on each failed test.

### Adding A New A11y Spec

1. Create `tests/a11y/<feature>.a11y.spec.ts`.
2. Import `test` from `../fixtures` if you need the cypher stub, or from `bh-playwright-testing` if you don't.
3. Navigate to the route, wait on a stable selector that proves the relevant UI has rendered, then call `await makeAxeBuilder()...analyze()`.
4. Prefer scoping with `.include('#content-wrapper')` or `.include('[role="dialog"]')` over full-page scans — scoped scans ignore global components, such as the nav, which are separately tested, are easier to debug, and less likely to flake on unrelated regressions.
5. Pass the result to `expectNoAccessibilityViolations(testInfo, results, { page })`. The `{ page }` opt is optional but recommended — it attaches a screenshot of each violating element to the report.
