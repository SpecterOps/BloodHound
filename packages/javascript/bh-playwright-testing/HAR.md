# Create a spec with HAR mocks

-   Run commands from the repository root. For first-time setup, install dependencies (`yarn install`) and Playwright browsers (`yarn playwright:install`).
-   Start a backend with data for your feature.
-   Set `A11Y_TEST_URL`, `A11Y_TEST_USERNAME`, and `A11Y_TEST_PASSWORD` in `cmd/ui/.env` (use `.env.example` as a starting point).
-   Record your spec (replace the name and route). Codegen logs in automatically; a separate `test:a11y` run is unnecessary:

```sh
A11Y_TEST_SERVE=true yarn codegen my-feature --url '/ui/your-feature'
```

-   Click through the feature and add assertions using the recorder's **Assert visibility**, **Assert text**, or **Assert value** tools.
-   Close the browser. Codegen creates your spec and mocks called HARs:

```text
cmd/ui/tests/e2e/my-feature.spec.ts
cmd/ui/tests/e2e/my-feature.har
```

-   Codegen runs the spec automatically. Review the assertions and captured data.
-   Run it again:

```sh
yarn test:e2e tests/e2e/my-feature.spec.ts
```

-   These examples use a local UI. `A11Y_TEST_SERVE=true` starts its server if needed; the `test:e2e` scripts already enable it.
-   HAR replay needs the UI and captured API responses; it does not need a backend or login.
-   To edit the generated test, keep its HAR reference:

```ts
import { test, expect } from 'bh-playwright-testing/har';

test.use({ replayHar: new URL('./my-feature.har', import.meta.url) });

test('my-feature', async ({ page }) => {
    await page.goto('/ui/your-feature');
    await expect(page.getByRole('heading', { name: 'My feature' })).toBeVisible();
});
```

## Update HARs

-   Start the backend and update one spec's HAR:

```sh
A11Y_TEST_SERVE=true yarn test:e2e:update-mocks my-feature
```

-   To run the suite and update HARs only for failing specs:

```sh
yarn test:e2e:update-failing-mocks
```

-   Each updater logs in and replaces a spec's HAR only after that spec passes against the live backend. Neither updater runs codegen or changes specs.
-   `update-failing-mocks` first replays the suite, updates only failing specs, then replays those specs again. The single-spec updater captures live responses without a separate replay afterward; use `yarn test:e2e tests/e2e/my-feature.spec.ts` to verify replay.
-   Review captured data before committing. Delete a spec and its HAR together when removing a flow.
