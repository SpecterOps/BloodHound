# BloodHound UI Playwright Tests

This directory is the entry point for browser-driven Bloodhound UI Playwright test suites. It is organized into per-suite subdirectories. Each suite has its own Playwright config, `testMatch` pattern, and artifact subfolder under `cmd/ui/playwright/`.

Prefer Playwright when a test's outcome depends on real-browser behavior (layout, computed styles, the accessibility tree, focus/keyboard, cross-browser rendering, or full-app navigation) that jsdom can't faithfully reproduce; use Vitest for fast, isolated unit and component logic.

Because these suites drive a real browser against a running app, the live state of that environment — backend data, auth session, feature flags, timing — can influence results. The suites install mocks and stubs to pin the states tests depend on, but full isolation from the environment isn't always achievable.

## Quick Start

First-time setup and a full run of the `a11y` suite. All commands are run from the root (`/`); the same script names also work from within `cmd/ui`.

1. **Install dependencies** — from the root, install workspace packages and the Playwright browsers the suite drives:

    ```sh
    yarn install
    yarn workspace bloodhound-ui playwright install chromium firefox
    ```

2. **Configure environment** — copy the example env and fill in real login credentials (see [Required Environment Variables](#required-environment-variables) for what each key does):

    ```sh
    cp cmd/ui/.env.example cmd/ui/.env
    # then edit cmd/ui/.env and set A11Y_TEST_USERNAME / A11Y_TEST_PASSWORD
    ```

    With `A11Y_TEST_SERVE="true"` (the default in `.env.example`), Playwright starts the Vite dev server itself, so no separately-running UI is required.

3. **Run the suite for the first time** — this clears `cmd/ui/playwright/` and runs the full browser × theme matrix:

    ```sh
    # Run the full test matrix
    yarn test:a11y

    # Specify browser and themes
    yarn test:a11y --project='chromium-light' --project='firefox-dark'

    # Run a single spec file, or a specific test within it by title
    yarn test:a11y tests/a11y/login.a11y.spec.ts
    yarn test:a11y tests/a11y/login.a11y.spec.ts -g 'no detectable WCAG'
    ```

    See [Running The Suite](#running-the-suite) for filtering to specific projects or interactive UI mode.

4. **View the Allure report** — the run writes raw Allure results but not HTML, so rendering needs the `allure` CLI once:

    ```sh
    brew install allure    # one-time; see Allure Report section for non-macOS
    yarn report:a11y       # builds and serves report HTML and opens a browser
    ```

    See [Viewing The Reports](#viewing-the-reports) for a persistent build and the plain Playwright HTML report.

## Layout

```text
tests/
├── global.setup.ts   # Logs in once and snapshots auth storage state
└── a11y/             # Accessibility regression suite
    └── *.a11y.spec.ts
```

Shared Playwright building blocks (fixtures, stubs, and auth helper) live in the `bh-playwright-testing` workspace package (`packages/javascript/bh-playwright-testing`) for reusability. See the [full module map](../../../packages/javascript/bh-playwright-testing/README.md#modules) in that package's README.

## Shared Scaffolding

### `global.setup.ts`

A Playwright **setup project** that runs once per Playwright invocation before any browser-theme project executes. It:

1. Installs any stubs
2. Delegates to `bh-playwright-testing` helpers to manage per-theme auth sessions

## Accessibility Tests (`a11y/`)

The accessibility suite runs `axe-core` scans against the live BloodHound UI through `@axe-core/playwright`. Shared scan defaults, reporting helpers, and the `makeAxeBuilder` fixture come from the `bh-playwright-testing` workspace package (`packages/javascript/bh-playwright-testing`). See the [`axe` fixture API](../../../packages/javascript/bh-playwright-testing/README.md#axe) in that package's README.

### Scope

The a11y suite's goal is automated WCAG 2.x accessibility regression coverage for user-facing pages in the BloodHound UI. See [Adding A New A11y Spec](#adding-a-new-a11y-spec) for the per-spec recipe.

Each spec is a self-contained scan of one route, covering any of its in-page states. Specs are kept narrow rather than chained because axe violations are easier to triage when the failure points at a single, well-scoped DOM subtree.

### Shared Fixtures And Options

Specs import `test` directly from `bh-playwright-testing`, which provides the `checkA11y` and `goAndWaitFor` helpers alongside `makeAxeBuilder`. App-specific behavior is primed once in `playwright.a11y.config.ts` via the `use` block:

-   `installGraphDataStub: true` installs the cypher "has data" stub on every test's `page` so the "No Data Available" dialog stays shut. A spec that needs it off (e.g. `login.a11y.spec.ts`) opts out with `test.use({ installGraphDataStub: false })`.
-   `a11yDefaultInclude: '#content-wrapper'` is the default scan scope used by `checkA11y()`.
-   `navToggleName: 'Toggle Navigation'` is the button `goAndWaitFor` clicks to collapse the nav.

See the [`axe` fixture API](../../../packages/javascript/bh-playwright-testing/README.md#axe) in that package's README for the full helper and option reference.

### Running The Suite

From the root or from within `cmd/ui`:

```sh
yarn test:a11y       # clears cmd/ui/playwright folder and runs the suite
```

Running in interactive UI mode

```sh
yarn test:a11y --ui
```

The root script delegates to the `bloodhound-ui` workspace, whose `test:a11y` script clears the playwright artifact directory (`cmd/ui/playwright`) and runs the a11y test suite as configured in `playwright.a11y.config.ts` — the clean step is baked in so every run starts with a fresh `playwright/` directory. CI-mode behavior (single worker, 1 retry, `forbidOnly`) is auto-enabled when `process.env.CI` is set, so there is no separate `:ci` script. The Playwright config (`cmd/ui/playwright.a11y.config.ts`) generates a project matrix of `chromium-light`, `chromium-dark`, `firefox-light`, `firefox-dark`, each depending on the `setup` project.

By default, the full 2x2 matrix (browsers x themes) is run, but projects may be individually specified:

```sh
# Ex. Running light and dark Chromium browser projects
yarn test:a11y --project='chromium-light' --project='chromium-dark'
```

To run a single test, pass a spec file (optionally with `-g` to match a test title):

```sh
# A single spec file, or a specific test within it by title
yarn test:a11y tests/a11y/login.a11y.spec.ts
yarn test:a11y tests/a11y/login.a11y.spec.ts -g 'no detectable WCAG'
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

Local workflow: run the suite, then point a report viewer at the output. Yarn scripts are mirrored at both the root and the workspace (`cmd/ui/`); pick whichever cwd is convenient. Examples below default to the repo root.

#### Allure Report

The Allure reporter only writes raw `*-result.json` files — viewing them requires the `allure` CLI to generate HTML.

**One-time install:**

```sh
brew install allure              # macOS, recommended (brings the JRE)
# or:
npm i -g allure-commandline      # cross-platform, requires Java on PATH
```

**Per-run workflow (from the root):**

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

1. Create `tests/a11y/<Section>/<feature>.a11y.spec.ts`.
2. Import `test` from `bh-playwright-testing`.
3. Navigate with `await goAndWaitFor('<route>', <stableLocator>)`, which visits the route, collapses the nav, and waits on a selector that proves the relevant UI has rendered. Install any route stubs before calling it.
4. Call `await checkA11y()` to scan the default scope (`#content-wrapper`). For a narrower scan pass `checkA11y({ include: '[role="dialog"]' })` — scoped scans ignore global components, such as the nav, which are separately tested, are easier to debug, and less likely to flake on unrelated regressions.
5. For multiple scans in one test, give each a distinct `attachmentNamePrefix` (e.g. `checkA11y({ include: '...', attachmentNamePrefix: 'step-1' })`) so their report attachments don't collide.
