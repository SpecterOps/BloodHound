# BloodHound Playwright Testing Utils

-   Shared fixtures for accessibility, authentication, themes, API stubs, and HAR recordings.
-   Each UI workspace owns its Playwright config, credentials, routes, selectors, and reporters.

## Install

-   Add this workspace dependency to the consuming UI's `package.json`.

```json
{
    "devDependencies": {
        "bh-playwright-testing": "workspace:*"
    }
}
```

-   Run from the consuming UI directory.

```sh
yarn install
yarn playwright install chromium
```

## Imports

-   `bh-playwright-testing`: shared `test`, `expect`, accessibility helpers, and HAR utilities.
-   `bh-playwright-testing/axe`: accessibility fixtures and helpers.
-   `bh-playwright-testing/auth`: `loginAndSnapshotThemes`.
-   `bh-playwright-testing/themes`: `THEMES`, `Theme`, `TestOptions`, and `authStorageStateFor`.
-   `bh-playwright-testing/stubs`: API route stubs.

## Accessibility

-   Import the shared `test` to access `checkA11y` and `goAndWaitFor`.
-   `goAndWaitFor` navigates, collapses navigation, and waits for a locator.
-   `checkA11y` scans the configured scope and asserts no violations.

```ts
import { test } from 'bh-playwright-testing';

test('login form is accessible', async ({ page, goAndWaitFor, checkA11y }) => {
    await goAndWaitFor('/ui/login', page.getByRole('textbox', { name: 'Email Address' }));
    await checkA11y({ include: null });
});
```

-   `include: null` scans the full page; a selector limits the scan.
-   Set shared scan options at file or `describe` scope.

```ts
test.use({ a11yDefaults: { include: '#content-wrapper', failOnIncomplete: true } });
```

-   `makeAxeBuilder()` supports custom scans using WCAG 2.0 and 2.1 A/AA tags.
-   `expectNoAccessibilityViolations` attaches reports; `{ page }` also attaches screenshots of affected elements.

```ts
import { expectNoAccessibilityViolations, test } from 'bh-playwright-testing';

test('scan content', async ({ page, makeAxeBuilder }, testInfo) => {
    await page.goto('/ui/login');
    const results = await makeAxeBuilder().include('main').analyze();
    await expectNoAccessibilityViolations(testInfo, results, { page });
});
```

-   `hideBySelector(page, selector)` temporarily hides content; `restoreHidden(handle)` restores it.

## Authentication and themes

-   Log in once; save authenticated light and dark snapshots.
-   Store setup in `tests/global.setup.ts`; run commands from the UI directory.
-   Supply credentials through the consuming workspace's environment.

```ts
import path from 'node:path';
import { test as setup } from 'bh-playwright-testing';
import { loginAndSnapshotThemes } from 'bh-playwright-testing/auth';
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs';
import { authStorageStateFor } from 'bh-playwright-testing/themes';

setup('save authentication', async ({ page }) => {
    await installGraphHasDataStub(page);
    await loginAndSnapshotThemes({
        page,
        username: process.env.A11Y_TEST_USERNAME!,
        password: process.env.A11Y_TEST_PASSWORD!,
        storageStatePathFor: (theme) => path.resolve(authStorageStateFor(theme)),
    });
});
```

## Configure the suite

-   Use `A11yTestOptions` for typed accessibility, theme, and HAR settings.
-   `a11yDefaultInclude` defaults to full-page scanning; `a11yDefaults` defaults to `{}`.
-   `navToggleName` defaults to `Toggle Navigation`; `installGraphDataStub` defaults to `false`.
-   Save this example as `playwright.a11y.config.ts`; keep specs under `tests/a11y/`.

```ts
import { defineConfig, devices } from '@playwright/test';
import type { A11yTestOptions } from 'bh-playwright-testing';
import { authStorageStateFor, THEMES } from 'bh-playwright-testing/themes';
import dotenv from 'dotenv';

dotenv.config();

export default defineConfig<A11yTestOptions>({
    testDir: './tests',
    use: {
        baseURL: process.env.A11Y_TEST_URL,
        serviceWorkers: 'block',
        installGraphDataStub: true,
        a11yDefaultInclude: '#content-wrapper',
        navToggleName: 'Toggle Navigation',
    },
    projects: [
        { name: 'setup', testMatch: /global\.setup\.ts$/ },
        ...THEMES.map((theme) => ({
            name: `chromium-${theme}`,
            testMatch: '**/*.a11y.spec.ts',
            use: { ...devices['Desktop Chrome'], storageState: authStorageStateFor(theme), theme },
            dependencies: ['setup'],
        })),
    ],
});
```

-   Start the UI and API, then run the suite from the UI directory.

```sh
yarn playwright test -c playwright.a11y.config.ts
```

-   Use `TestOptions` from `bh-playwright-testing/themes` when only typing the theme option.

## API stubs

-   Install stubs before navigation or actions that trigger their requests.
-   Unhandled requests fall through; later route handlers take precedence.

```ts
import { test } from 'bh-playwright-testing';
import { installMFAEnrollmentStub } from 'bh-playwright-testing/stubs';

test.beforeEach(async ({ page }) => {
    await installMFAEnrollmentStub(page);
});
```

## HAR recording and replay

-   Import `test` from `bh-playwright-testing`; the base Playwright fixture does not recognize `harMode`.
-   `off`: default; disables HAR capture and replay.
-   `record`: captures live traffic.
-   `update`: refreshes an existing recording using live traffic.
-   `mock`: replays recorded responses; unmatched requests fail by default.

### Record

-   Save as `tests/a11y/har-example.a11y.spec.ts`; adapt actions and assertions to your scenario.
-   Run against the live UI and API.

```ts
import { expect, test } from 'bh-playwright-testing';

test.use({ harMode: 'record' });

test('graph page', async ({ page }) => {
    await page.goto('/ui/graphview');
    await expect(page.getByRole('main')).toBeVisible();
});
```

```sh
yarn playwright test -c playwright.a11y.config.ts --project=setup
yarn playwright test -c playwright.a11y.config.ts \
  tests/a11y/har-example.a11y.spec.ts --project=chromium-light --no-deps
```

### Replay

-   Change only the mode; keep the filename, titles, and project unchanged.
-   Reuse saved authentication with `--no-deps`.

```ts
test.use({ harMode: 'mock' });
```

```sh
A11Y_TEST_SERVE=false yarn playwright test -c playwright.a11y.config.ts \
  tests/a11y/har-example.a11y.spec.ts --project=chromium-light --no-deps
```

### Update

-   Run the same command against live services, then restore `mock` and verify replay.
-   Failed updates preserve the previous recording.

```ts
test.use({ harMode: 'update' });
```

```sh
yarn playwright test -c playwright.a11y.config.ts \
  tests/a11y/har-example.a11y.spec.ts --project=chromium-light --no-deps
```

### Options and artifacts

-   Set HAR options at file scope, inside `describe`, or in typed config `use`.
-   `harRootDir` defaults to `test-artifacts/har`, relative to the configured test root.
-   `harNotFound: 'fallback'` allows unmatched requests to reach live services.

```ts
test.use({
    harMode: 'mock',
    harRootDir: './test-artifacts/har',
    harNotFound: 'abort',
});
```

-   Each test directory contains these files.

```text
test-artifacts/har/<test-label>-<identity-hash>/
  recording.har
  requests.json
  responses.json
  .har-fixture.json
```

-   JSON summaries contain captured methods, URLs, statuses, MIME types, and sizes.
-   `harArtifacts` exposes `directory`, `recording`, `requests`, and `responses` paths.
-   Files, titles, projects, and repeated tests determine directory identity; retries reuse the same directory.
-   Record separately for every project that will replay the test.
-   HARs finalize after context closure, including failed tests when closure succeeds.
-   Use the fixture's default context; manually created contexts are unmanaged.
-   Block service workers; avoid route stubs for endpoints you want HARs to serve.
-   Inspect recordings for credentials, cookies, and sensitive response data before committing.
-   Keep temporary Playwright output outside version control.

```ts
import {
    makeHarArtifactPaths,
    readHar,
    readJsonArtifact,
    writeJsonArtifact,
    summarizeHar,
} from 'bh-playwright-testing';
```

### Remove obsolete recordings

-   Renaming tests leaves old directories; updates only replace recordings for unchanged identities.
-   From either UI workspace, preview orphaned fixtures before deleting.
-   Pruning checks the complete suite and preserves unmarked directories.

```sh
yarn har:prune
yarn har:prune --delete
yarn har:prune --root test-artifacts/custom-har
```

## Extend the fixture

-   Add suite-specific fixtures with `test.extend`.

```ts
import { test as base } from 'bh-playwright-testing';

export const test = base.extend<{ featurePath: string }>({
    featurePath: '/ui/graphview',
});
```

## Package development

-   The package exports TypeScript source directly; consumers need no package build.
-   Run from `bhce/packages/javascript/bh-playwright-testing` in BHE, or `packages/javascript/bh-playwright-testing` in BHCE.

```sh
yarn check-types
yarn lint
yarn test
node --test ../../../cmd/ui/scripts/prune-har.check.mjs
```

-   `yarn test` runs record, update, and mock tests against local endpoints.
