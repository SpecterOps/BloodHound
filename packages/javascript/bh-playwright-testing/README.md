# BloodHound Playwright Testing Utils

Shared Playwright testing utilities for BloodHound UI workspaces — axe-core fixture and reporting, auth bootstrap, route stubs, and theme matrix helpers.

## Purpose

`bh-playwright-testing` centralizes the Playwright building blocks that are common across BloodHound UI suites and consumers (e.g. CE's `cmd/ui` and BHE's `cmd/ui`). It exists so each consumer can compose suites — accessibility today, end-to-end and visual regression later — out of the same fixtures, snapshot helpers, and route stubs without reimplementing them.

The package intentionally does **not** own:

-   Playwright configs (browsers, projects, reporters, web server).
-   The pages, routes, or DOM subtrees that get scanned or asserted against.
-   App-specific environment variables (e.g. `*_TEST_URL`, credentials).
-   Suite-specific orchestration (which selectors to wait on, which routes to scope).

Consumers compose those concerns on top of the modules below.

## Modules

The package is consumed via subpath imports so each consumer pulls only what it uses. Each subpath is a thin, focused module — no cross-module coupling beyond shared `Theme` types.

| Subpath                                      | Exports                                                                                                       | Purpose                                                                                                                                      |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `bh-playwright-testing`                      | `test`, `expect`, `WCAG_TAGS`, `attachAxeReport`, `expectNoAccessibilityViolations`, `AttachAxeReportOptions` | Default export — the axe-core fixture and reporting helpers. Convenience re-export of `./axe`.                                               |
| `bh-playwright-testing/axe`                  | same as above                                                                                                 | Same as the default export. Use this path when you want to be explicit.                                                                      |
| `bh-playwright-testing/themes`               | `Theme`, `THEMES`, `TestOptions`, `authStorageStateFor`                                                       | Type-only and constant helpers for the per-theme storage-state convention used by `loginAndSnapshotThemes` and Playwright configs.           |
| `bh-playwright-testing/auth`                 | `loginAndSnapshotThemes`, `LoginAndSnapshotThemesOptions`                                                     | One-time login helper that snapshots `storageState` for both light and dark themes from a single session.                                    |
| `bh-playwright-testing/stubs/graph-has-data` | `installGraphHasDataStub`                                                                                     | Stubs `POST /api/v2/graphs/cypher` so the `useGraphHasData` probe resolves to "true" and the "No Data Available" upload dialog stays closed. |

### `axe`

The axe fixture is a thin wrapper around `@axe-core/playwright`'s `AxeBuilder`:

1. `test` is Playwright's `test` extended with a `makeAxeBuilder()` fixture and a worker-scoped `theme` option (default `'light'`).
2. Calling `makeAxeBuilder()` returns a fresh `AxeBuilder` bound to the current `page` and constrained to `WCAG_TAGS` (`wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`).
3. The spec chains scoping or rule methods it needs (`.include(...)`, `.exclude(...)`, `.disableRules(...)`) and `await builder.analyze()`.
4. The `AxeResults` are handed to `expectNoAccessibilityViolations(testInfo, results, opts?)`, which attaches `axe-results.json` (always) and `a11y-violations.md` (only on failure) before asserting that `results.violations` is empty.

The fixture also records an `a11y-tags` annotation on `testInfo` so the active tag set shows up in Playwright and Allure reports.

#### Per-Node Screenshots

`expectNoAccessibilityViolations` (and `attachAxeReport`) accept an optional third argument:

```ts
type AttachAxeReportOptions = {
    page?: Page;
    maxNodesPerViolation?: number; // default 5
};
```

When `page` is provided and there are violations, the helper screenshots each affected element via `page.locator(node.target).first().screenshot()` and attaches it as `a11y-<violation.id>-<n>.png` so the Playwright HTML and Allure reports show a visual indicator for each violation right next to the textual one. `maxNodesPerViolation` caps how many element screenshots are taken per rule (e.g. a `color-contrast` violation spanning 30 elements only attaches the first 5 PNGs).

Targets that cross iframe or shadow-DOM boundaries (where axe returns a nested-array target) are skipped — Playwright requires a different API for those, and the textual report still describes the violation. Per-element screenshot failures (detached/animated-off elements) are swallowed so a missing screenshot never blocks the assertion.

Without `page`, behavior is unchanged (textual attachments only).

### `themes`

Type-only and constant helpers. `Theme = 'light' | 'dark'`, `THEMES` is the `readonly` tuple `['light', 'dark']`, and `TestOptions` is the worker-scoped option shape consumed at `defineConfig<TestOptions>` time.

`authStorageStateFor(theme)` returns the canonical path Playwright projects should pass to `use.storageState` — `./playwright/.auth/user-<theme>.json` relative to the consumer's Playwright project root. The `auth` module writes to the same paths via the caller-supplied `storageStatePathFor` callback, so the two modules agree on the layout without one depending on the other.

### `auth`

`loginAndSnapshotThemes` logs in once and snapshots `storageState` for both themes from the same session. Capturing both snapshots from one session avoids the parallel-login race where two setups as the same user would invalidate each other's session. The helper assumes the BloodHound shared UI shell (login form labels `Email Address` / `Password`, the `LOGIN` submit button, the `global_nav-dark-mode` toggle, and the `persistedState` localStorage key written by the global store), which is shared across CE and BHE.

An optional `dismissPostLogin(page)` hook lets the caller dismiss a post-login overlay (e.g. the "No Data Available" dialog) that could intercept the dark-mode toggle click.

### `stubs/graph-has-data`

`installGraphHasDataStub(page)` registers a `page.route` for `POST **/api/v2/graphs/cypher` that returns a populated payload. Install it before navigation. Tests that need a different cypher response can register a higher-priority handler for the same URL — Playwright runs `page.route` handlers in LIFO order, so a test-local handler wins for the cases it cares about. Non-`POST` traffic falls through.

## Usage

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
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs/graph-has-data';
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

A Playwright config that consumes the theme matrix:

```ts
import { defineConfig, devices } from '@playwright/test';
import { authStorageStateFor, THEMES, type TestOptions } from 'bh-playwright-testing/themes';

export default defineConfig<TestOptions>({
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

### Extending The Fixture

Consumers can wrap the shared `test` to layer suite-specific fixtures on top of `makeAxeBuilder` (see `cmd/ui/tests/fixtures.ts` for the BloodHound UI suite's wrapper that installs `installGraphHasDataStub` on every test's `page`):

```ts
import { test as playwrightTest } from 'bh-playwright-testing';
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs/graph-has-data';

export const test = playwrightTest.extend({
    page: async ({ page }, use) => {
        await installGraphHasDataStub(page);
        await use(page);
    },
});

export { expect, expectNoAccessibilityViolations } from 'bh-playwright-testing';
```

## Source-Only Distribution

The package ships TypeScript source via the `exports` map — there is no compiled `dist`. Consumers run it directly through their own Vite/Playwright TS pipelines. This avoids a build step that would only ever be consumed inside the monorepo and keeps the modules editable in place. `tsc --noEmit` (`yarn check-types`, also wired to `yarn lint`) is the only type-check.
