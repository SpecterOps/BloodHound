# API replay with one HAR per flow

Codegen creates `<name>.spec.ts` and one adjacent `<name>.har` with embedded API responses. Run `yarn codegen <name> --url /ui/feature` from a consuming repository. If either file exists, codegen warns and selects an available integer suffix for both files and the test name.

```ts
import { test, expect } from 'bh-playwright-testing/mocked';

test.use({ replayHar: new URL('./my-feature.har', import.meta.url) });

test('my-feature', async ({ page }) => {
    await page.goto('/ui/feature');
    await expect(page.getByRole('heading', { name: 'My feature' })).toBeVisible();
});
```

The fixture rebases API request URLs to the configured `baseURL`, blocks service workers, and starts with an isolated fake session. It replays `/api/` requests and aborts requests missing from the HAR. Frontend assets load from the UI server. No backend login is needed during replay.

The mocked fixture also exposes the shared accessibility helpers, including `checkA11y`. Regular accessibility tests can continue importing `bh-playwright-testing` without loading HAR setup.

## Refresh and cleanup

Record the flow again, review the new pair, and replace the old HAR with the new capture. Update the spec only when the recorded actions or assertions changed. Delete the temporary suffixed pair after copying the changes. Rename or delete the spec and HAR together; update `replayHar` when renaming.

`yarn har:prune` previews HAR files under `cmd/ui/tests/e2e` without a matching adjacent spec. Add `--delete` to remove them. There are no per-project copies, test identity hashes, markers, or JSON summary files.

Review captured application data before committing. Codegen sanitizes authorization and cookie headers and the login session token. Dynamic request bodies, external services, WebSockets, and changing responses to identical requests may need explicit route mocks.

## Validation

Run `yarn workspace bh-playwright-testing test` for the self-contained replay check. It serves a small local page, replays GET and POST responses on a different host, verifies the fake session, and confirms unrecorded API calls cannot reach the server. No application backend is needed.
