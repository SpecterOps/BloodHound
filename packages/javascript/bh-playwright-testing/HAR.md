# Create a spec with HAR mocks

-   Run commands from the repository root.
-   Start a backend with data for your feature.
-   Set `A11Y_TEST_USERNAME` and `A11Y_TEST_PASSWORD` in `cmd/ui/.env`.
-   Record your spec:

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
A11Y_TEST_SERVE=true yarn test:e2e tests/e2e/my-feature.spec.ts
```

-   Replay needs a UI server; it does not need a backend or login.
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

-   The command logs in, runs existing specs against the live backend, and replaces their HARs after passing specs. It never runs codegen or changes specs.
-   Review captured data before committing. Delete a spec and its HAR together when removing a flow.
