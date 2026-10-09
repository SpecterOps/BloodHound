# HAR recording and replay

-   Run UI commands from the consuming UI's `cmd/ui` directory.
-   Configure authentication and projects using the [package README](README.md).
-   Enable `serviceWorkers: 'block'` in your Playwright config.

-   Import `test` from `bh-playwright-testing`; the base Playwright fixture does not recognize `harMode`.
-   `off`: default; disables HAR capture and replay.
-   `record`: captures live traffic.
-   `update`: refreshes an existing recording using live traffic.
-   `mock`: replays recorded responses; unmatched requests fail by default.

## Record

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

## Replay

-   Change only the mode; keep the filename, titles, and project unchanged.
-   Reuse saved authentication with `--no-deps`.

```ts
test.use({ harMode: 'mock' });
```

```sh
A11Y_TEST_SERVE=false yarn playwright test -c playwright.a11y.config.ts \
  tests/a11y/har-example.a11y.spec.ts --project=chromium-light --no-deps
```

## Update

-   Run the same command against live services, then restore `mock` and verify replay.
-   Failed updates preserve the previous recording.

```ts
test.use({ harMode: 'update' });
```

```sh
yarn playwright test -c playwright.a11y.config.ts \
  tests/a11y/har-example.a11y.spec.ts --project=chromium-light --no-deps
```

## Options and artifacts

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

## Remove obsolete recordings

-   Renaming tests leaves old directories; updates only replace recordings for unchanged identities.
-   From either UI workspace, preview orphaned fixtures before deleting.
-   Pruning checks the complete suite and preserves unmarked directories.

```sh
yarn har:prune
yarn har:prune --delete
yarn har:prune --root test-artifacts/custom-har
```

## Test the HAR utilities

-   Run from this package directory.
-   Tests exercise record, update, and mock modes against local endpoints.

```sh
yarn test
yarn check-types
yarn lint
node --test ../../../cmd/ui/scripts/prune-har.check.mjs
```
