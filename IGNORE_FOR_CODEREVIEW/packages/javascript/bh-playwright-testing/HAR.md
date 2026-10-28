# Create a test with HAR mocks

- Run commands from your application repository root.
- Start a backend with data for your feature.
- Set `A11Y_TEST_USERNAME` and `A11Y_TEST_PASSWORD` in `cmd/ui/.env`.
- Record your test:

```sh
A11Y_TEST_SERVE=true yarn codegen my-feature --url '/ui/your-feature'
```

- Click through the feature and add assertions using the recorder's **Assert visibility**, **Assert text**, or **Assert value** tools.
- Close the browser. Codegen creates your test and mocks called HARs:

```text
cmd/ui/tests/e2e/my-feature.spec.ts
cmd/ui/tests/e2e/my-feature.har
```

- Codegen runs the test automatically. Review the assertions and captured data.
- Run it again:

```sh
A11Y_TEST_SERVE=true yarn test:e2e tests/e2e/my-feature.spec.ts
```

- Replay needs a UI server; it does not need a backend or login.
- To edit the generated test, keep its HAR reference:

```ts
import { test, expect } from 'bh-playwright-testing/mocked';

test.use({ replayHar: new URL('./my-feature.har', import.meta.url) });

test('my-feature', async ({ page }) => {
    await page.goto('/ui/your-feature');
    await expect(page.getByRole('heading', { name: 'My feature' })).toBeVisible();
});
```

## Update or remove

- Record again with the same name. Existing files are preserved; new files get a suffix such as `my-feature-1`.
- Review the new pair. Replace the original HAR and copy any changed actions or assertions into the original spec.
- Keep the original spec's `./my-feature.har` reference. Delete the temporary pair and rerun the original test.
- Delete or rename a spec and its HAR together; update the HAR reference when renaming.
- Preview or delete HARs without matching specs:

```sh
yarn har:prune
yarn har:prune --delete
```
