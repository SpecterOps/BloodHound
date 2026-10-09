# BloodHound Playwright Testing Utils

Shared Playwright testing utilities for BloodHound UI workspaces.

## Purpose

`bh-playwright-testing` centralizes the Playwright building blocks that are common across BloodHound UI suites and consumers (e.g. both BHE and BHCE's `cmd/ui`). It exists so each consumer can consistently compose test suites without reimplementing common features:

-   axe-core accessibility, fixture, and reporting helpers
-   one-time auth bootstrap
-   `page.route` API stubs
-   theme matrix helpers

The package intentionally does **not** own:

-   Playwright configs (browsers, projects, reporters, web server).
-   The pages, routes, or DOM subtrees that get scanned or asserted against.
-   App-specific environment variables (e.g. `*_TEST_URL`, credentials).
-   Suite-specific orchestration (which selectors to wait on, which routes to scope).

Consumers compose those concerns on top of the modules below.

## Modules

The package is consumed via subpath imports so each consumer pulls only what it uses.

| Subpath                        | Purpose                                                                                                                                                  |
| ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bh-playwright-testing`        | Package root entry (re-export of `./axe`): the shared `test`/`expect`, the axe fixture, the `checkA11y` / `goAndWaitFor` helpers, and reporting helpers. |
| `bh-playwright-testing/axe`    | Same exports as the root entry. Use this path when you want to be explicit.                                                                              |
| `bh-playwright-testing/themes` | Theme types and constants (`Theme`, `THEMES`, `TestOptions`, `authStorageStateFor`) for the per-theme storage-state convention.                          |
| `bh-playwright-testing/auth`   | `loginAndSnapshotThemes` — logs in once and snapshots `storageState` for both light and dark themes.                                                     |
| `bh-playwright-testing/stubs`  | A barrel of `page.route` stubs so tests can control API responses without mutating real state.                                                           |

### Accessibility (`axe`)

The shared `test` extends Playwright's `test` with an axe-core fixture and a couple of helpers that collapse the boilerplate most specs repeat:

-   `makeAxeBuilder()` returns a fresh `AxeBuilder` bound to the current `page` and constrained to `WCAG_TAGS` (`wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`). Chain `.include(...)` / `.exclude(...)` / `.disableRules(...)` as needed and `await builder.analyze()`.
-   `expectNoAccessibilityViolations(testInfo, results, opts?)` attaches the axe report and asserts there are no violations. Pass `{ page }` to also attach a screenshot of each affected element.
-   `checkA11y(options?)` runs a full-page scan by default, can optionally be scoped, and asserts no violations in one call (builder + `expectNoAccessibilityViolations`).
-   `goAndWaitFor(path, target, options?)` navigates to `path`, collapses the global nav, then waits for `target` (a `Locator` or `(page) => Locator`) to become visible.
-   `hideBySelector(page, selector)` / `restoreHidden(handle)` temporarily hide background content (e.g. behind a dialog) that would otherwise produce noisy `incomplete` results.

`checkA11y` and `goAndWaitFor` read a few consumer-primed Playwright options so app-specific assumptions live in each consumer's config `use` block rather than in the package:

| Option                 | Default               | Purpose                                                                                                 |
| ---------------------- | --------------------- | ------------------------------------------------------------------------------------------------------- |
| `a11yDefaultInclude`   | `null` (full page)    | Default scan scope for `checkA11y` (e.g. `'#content-wrapper'`).                                         |
| `a11yDefaults`         | `{}`                  | Describe-scoped default `A11yScanOptions`; set once via `test.use({ a11yDefaults: { ... } })`.          |
| `navToggleName`        | `'Toggle Navigation'` | Accessible name of the button `goAndWaitFor` clicks to collapse the nav.                                |
| `installGraphDataStub` | `false`               | When `true`, the `page` fixture installs the cypher "has data" stub so the "No Data" dialog stays shut. |

Prime them in the config `use` block (typed via `A11yTestOptions`, which combines `TestOptions` with these four options).

### Themes (`themes`)

Theme TypeScript types and constants, plus `authStorageStateFor(theme)` — the canonical per-theme `storageState` path shared by `loginAndSnapshotThemes` and Playwright project configs.

### Auth (`auth`)

`loginAndSnapshotThemes(...)` logs in once and writes an authenticated `storageState` snapshot for both light and dark themes, avoiding the parallel-login race two setups as the same user would hit.

### Stubs (`stubs`)

A single barrel of `page.route` stubs, grouped by the API surface they cover — asset group tags, BloodHound users (MFA enrollment, password reset), feature flags, graph data (cypher), and API tokens. Each stub installs a route handler that falls through (`route.fallback()`) for requests it doesn't own, so stubs compose and test-local overrides win under Playwright's LIFO routing.

```ts
import { installMFAEnrollmentStub } from 'bh-playwright-testing/stubs';

test('MFA dialog', async ({ page }) => {
    await installMFAEnrollmentStub(page);
    // Click the MFA toggle and walk the dialog steps.
});
```

## Usage

### HAR recordings and mocks

The shared `test` fixture supports `harMode: 'record' | 'update' | 'mock' | 'off'` (default `off`). Set it at file or describe scope, or in a typed Playwright config. Existing `page` test bodies need no changes:

```ts
import { test } from 'bh-playwright-testing';

test.use({ harMode: 'record' }); // switch to 'update' or 'mock' for the same test
test('login', async ({ page }) => {
    await page.goto('/login');
});
```

```ts
import { defineConfig } from '@playwright/test';
import type { A11yTestOptions } from 'bh-playwright-testing';

export default defineConfig<A11yTestOptions>({
    use: { harMode: 'mock', harRootDir: './test-artifacts/har', harNotFound: 'abort' },
});
```

`harRootDir` defaults to `test-artifacts/har` relative to the Playwright config's test root. Each HAR-enabled test has a canonical directory named from its spec filename, suite, title, and project, followed by a stable hash of its full identity. Different files, parameterized titles, projects, and `repeatEach` instances get separate directories. The path is stable across runs and retries; `harArtifacts` exposes its `directory`, `recording`, `requests`, and `responses` paths to tests. A directory contains `recording.har`, `requests.json`, and `responses.json`. The JSON files summarize actual HAR entries (method, URL, status, MIME type, and size). The helper functions `readHar`, `readJsonArtifact`, `writeJsonArtifact`, and `makeHarArtifactPaths` are exported for tooling.

Run once in `record` mode against a live service, then switch to `mock` for replay. Use `update` against the live service to refresh a fixture. Record and update first write to Playwright's per-execution output path; a valid completed HAR replaces the canonical file after the browser context closes. Update keeps the previous recording if the test fails. Recordings are also finalized after a test failure when Playwright can close the context. Native HAR capture and update require the default Playwright context lifecycle; manually created contexts are outside this fixture. Service workers can bypass HAR routing, so use `serviceWorkers: 'block'` when the app registers them.

Mock mode aborts unmatched requests by default and reports failed requests. Set `harNotFound: 'fallback'` to allow unmatched requests to reach the live network. HAR files can contain credentials, cookies, and response bodies: inspect and redact them before committing. Keep temporary Playwright output out of version control.

Updating a test replaces its HAR and summaries in the same directory. Renaming or moving a test, changing its project, or changing a parameterized title creates a new canonical directory; the old one remains. In BHCE UI, `yarn har:prune` compares marked HAR directories with the complete a11y test list and previews orphaned directories. `yarn har:prune --delete` removes only the previewed, marked directories. Run the preview after test reorganizations; unmarked hand-maintained fixtures are left alone.

Add the package as a workspace `devDependency`:

```json
"bh-playwright-testing": "workspace:*"
```

A typical accessibility spec:

```ts
import { expect, expectNoAccessibilityViolations, test } from 'bh-playwright-testing';

test('login form has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
    await page.goto('/ui/login');
    await expect(page.getByRole('textbox', { name: 'Email Address' })).toBeVisible();

    const results = await makeAxeBuilder().analyze();
    // Pass `{ page }` so each violation's affected nodes are screenshotted and attached
    // to the test result. Omit it for text-only attachments.
    await expectNoAccessibilityViolations(testInfo, results, { page });
});
```

A `global.setup.ts` that bootstraps auth for both themes:

```ts
import path from 'path';
import { test as setup } from 'bh-playwright-testing';
import { loginAndSnapshotThemes } from 'bh-playwright-testing/auth';
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs';
import { authStorageStateFor, type Theme } from 'bh-playwright-testing/themes';

setup('Generate and cache auth state', async ({ page }) => {
    await installGraphHasDataStub(page);
    await loginAndSnapshotThemes({
        page,
        username: process.env.TEST_USERNAME!,
        password: process.env.TEST_PASSWORD!,
        storageStatePathFor: (theme: Theme) => path.resolve(__dirname, '..', authStorageStateFor(theme)),
    });
});
```

A Playwright config that consumes the theme matrix and primes the a11y helper options:

```ts
import { defineConfig, devices } from '@playwright/test';
import type { A11yTestOptions } from 'bh-playwright-testing';
import { authStorageStateFor, THEMES } from 'bh-playwright-testing/themes';

export default defineConfig<A11yTestOptions>({
    use: {
        // App-specific priming of the shared a11y helper options.
        installGraphDataStub: true,
        a11yDefaultInclude: '#content-wrapper',
        navToggleName: 'Toggle Navigation',
    },
    projects: [
        { name: 'setup', testMatch: /global\.setup\.ts$/ },
        ...THEMES.flatMap((theme) => [
            {
                name: `chromium-${theme}`,
                use: { ...devices['Desktop Chrome'], storageState: authStorageStateFor(theme), theme },
                dependencies: ['setup'],
            },
        ]),
    ],
});
```

Use `TestOptions` (from `bh-playwright-testing/themes`) instead of `A11yTestOptions` if a suite only needs the `theme` option and not the `checkA11y` / `goAndWaitFor` helpers.

### Extending The Fixture

Most suites need nothing beyond priming the options above — the `page` stub, `checkA11y`, and `goAndWaitFor` all ship in the shared `test`. For a bespoke fixture the options don't cover, wrap `test` with `test.extend` in the consuming suite:

```ts
import { test as base } from 'bh-playwright-testing';

export const test = base.extend({
    // ...suite-specific fixtures...
});
```

## Source-Only Distribution

The package ships TypeScript source via the `exports` map — there is no compiled `dist`. Consumers run it directly through their own Vite/Playwright TS pipelines. This avoids a build step that would only ever be consumed inside the monorepo and keeps the modules editable in place. `tsc --noEmit` (`yarn check-types`) is the only type-check.
