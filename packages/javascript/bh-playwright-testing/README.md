# BloodHound Playwright Testing Utils

Shared Playwright testing utilities for BloodHound UI workspaces.

## Purpose

`bh-playwright-testing` centralizes the Playwright building blocks that are common across BloodHound UI suites and consumers (e.g. both BHE and BHCE's `cmd/ui`). It exists so each consumer can consistently compose test suites without reimplementing common features. This includes:

-   axe-core fixture
-   reporting
-   auth bootstrap
-   route stubs
-   theme matrix helpers

The package intentionally does **not** own:

-   Playwright configs (browsers, projects, reporters, web server).
-   The pages, routes, or DOM subtrees that get scanned or asserted against.
-   App-specific environment variables (e.g. `*_TEST_URL`, credentials).
-   Suite-specific orchestration (which selectors to wait on, which routes to scope).

Consumers compose those concerns on top of the modules below.

## Modules

The package is consumed via subpath imports so each consumer pulls only what it uses. Each subpath is a thin, focused module — no cross-module coupling beyond shared `Theme` types.

| Subpath                        | Exports                                                                                                                                                                                                                                                                                                       | Purpose                                                                                                                                                                          |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bh-playwright-testing`        | `test`, `expect`, `WCAG_TAGS`, `attachAxeReport`, `expectNoAccessibilityViolations`, `hideBySelector`, `restoreHidden`, `AttachAxeReportOptions`                                                                                                                                                              | Package root entry — the axe-core fixture and reporting helpers, exposed as named exports (no default export). Convenience re-export of `./axe`.                                 |
| `bh-playwright-testing/axe`    | same as above                                                                                                                                                                                                                                                                                                 | Same named exports as the package root entry. Use this path when you want to be explicit.                                                                                        |
| `bh-playwright-testing/themes` | `Theme`, `THEMES`, `TestOptions`, `authStorageStateFor`                                                                                                                                                                                                                                                       | Type-only and constant helpers for the per-theme storage-state convention used by `loginAndSnapshotThemes` and Playwright configs.                                               |
| `bh-playwright-testing/auth`   | `loginAndSnapshotThemes`, `LoginAndSnapshotThemesOptions`                                                                                                                                                                                                                                                     | One-time login helper that snapshots `storageState` for both light and dark themes from a single session.                                                                        |
| `bh-playwright-testing/stubs`  | `GRAPHS_CYPHER_QUERY`, `installGraphHasDataStub`, `installGraphHasNoDataStub`, `installMFAEnrollmentStub`, `MFAEnrollmentStubOptions`, `installResetPasswordStub`, `installUserTokensStub`, `UserTokensStubOptions`, `installCreateUserTokenStub`, `CreateUserTokenStubOptions`, `installDeleteUserTokenStub` | Barrel of `page.route` stubs so tests can control API responses without mutating real state: cypher graph-data probes, MFA enrollment, password reset, and API token management. |

### `axe`

The axe fixture is a thin wrapper around `@axe-core/playwright`'s `AxeBuilder`:

1. `test` is Playwright's `test` extended with a `makeAxeBuilder()` fixture, a worker-scoped `theme` option (default `'light'`), and a `context` fixture that injects a `window.__APP_TEST_RUNTIME__` marker so the app can detect a Playwright run (e.g. to disable CSS transition animations).
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
    attachmentNamePrefix?: string;
};
```

When `page` is provided and there are violations, the helper screenshots each affected element and attaches it as `a11y-<violation.id>-<n>.png` so that reports can show a visual indicator for each violation next to the textual one. `maxNodesPerViolation` caps how many element screenshots are taken per rule (e.g. a `color-contrast` violation spanning 30 elements only attaches the first 5 PNGs).

`attachmentNamePrefix` keeps attachments distinct when a single test performs multiple scans, for example `mfa-password-axe-results.json`.

Targets that cross iframe or shadow-DOM boundaries are skipped — Playwright requires a different API for those, though the textual report still describes the violation. Per-element screenshot failures (detached/animated-off elements) are swallowed so a missing screenshot never blocks the assertion.

Without `page`, behavior is unchanged (textual attachments only).

#### Hiding Background Content

`hideBySelector(page, selector)` injects a `<style>` tag that hides matching elements with `visibility: hidden`, which is useful before an axe scan when background content (e.g. behind a dialog) produces noisy `incomplete` results. It returns the created `ElementHandle`; pass that handle to `restoreHidden(styleTag)` to remove the injected style and restore the content.

### `themes`

Theme TypeScript types and constants.

### `auth`

Auth storageState session snapshot helpers.

### `stubs`

A single barrel of `page.route` stubs, grouped by the API surface they cover. Each installs a route handler that falls through (`route.fallback()`) for methods/requests it doesn't own, so stubs compose and test-local overrides win under Playwright's LIFO routing.

#### Graph data (cypher)

Stubs for tests that need controlled cypher response states.

`installGraphHasDataStub(page)` stubs all cypher `POST` traffic with a populated graph payload. It is useful as a broad suite fixture when the "No Data Available" dialog should stay closed.

`installGraphHasNoDataStub(page)` only overrides the `useGraphHasData` probe (matched via the exported `GRAPHS_CYPHER_QUERY`) with an empty graph payload, then falls through for other cypher requests. Install it inside individual tests after any broad fixture route so Playwright's LIFO routing gives the test-local override priority.

#### MFA enrollment

`installMFAEnrollmentStub(page, opts?)` stubs the happy path for enabling MFA (the `POST .../mfa` and `POST .../mfa-activation` endpoints) so tests can walk the enrollment dialog without mutating the real user's MFA state. `MFAEnrollmentStubOptions` lets you override the returned `qrCode` and `totpSecret`.

```ts
import { installMFAEnrollmentStub } from 'bh-playwright-testing/stubs';

test('MFA dialog', async ({ page }) => {
    await installMFAEnrollmentStub(page);
    // Click the MFA toggle and walk the dialog steps.
});
```

#### Password reset

`installResetPasswordStub(page)` stubs the `PUT .../secret` endpoint so the Reset Password dialog can complete without changing the real user's password.

#### API tokens

Stubs for the API Key Management dialog so tests can render and exercise token flows without touching real tokens:

-   `installUserTokensStub(page, opts?)` stubs the `GET /api/v2/tokens` list. `UserTokensStubOptions` lets you supply custom `tokens` (pass an empty array for the "No tokens available" empty state).
-   `installCreateUserTokenStub(page, opts?)` stubs `POST /api/v2/tokens`. `CreateUserTokenStubOptions` lets you override the returned `token`.
-   `installDeleteUserTokenStub(page)` stubs `DELETE /api/v2/tokens/:id` so the revoke flow completes.

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
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs';

export const test = playwrightTest.extend({
    page: async ({ page }, use) => {
        await installGraphHasDataStub(page);
        await use(page);
    },
});

export { expect, expectNoAccessibilityViolations } from 'bh-playwright-testing';
```

## Source-Only Distribution

The package ships TypeScript source via the `exports` map — there is no compiled `dist`. Consumers run it directly through their own Vite/Playwright TS pipelines. This avoids a build step that would only ever be consumed inside the monorepo and keeps the modules editable in place. `tsc --noEmit` (`yarn check-types`) is the only type-check.
