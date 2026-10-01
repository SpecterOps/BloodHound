# bh-shared-ui

Shared React components and utilities for BloodHound Community Edition and BloodHound Enterprise.

## Storybook

From the BHCE workspace root (`bhce/`):

```sh
yarn install --immutable
yarn workspace bh-shared-ui storybook
```

Open http://localhost:6007. DoodleUI's Storybook uses port 6006, so both can run together.
Storybook resolves DoodleUI and the JS client from their sibling source directories; no prerequisite package build is needed.

```sh
yarn workspace bh-shared-ui build:storybook
yarn workspace bh-shared-ui check-types:storybook
```

Static output is written to `packages/javascript/bh-shared-ui/storybook-static/`.

### Adding stories

Place `*.stories.tsx` beside the component and import the component directly. 
Add `tags: ['autodocs']` for a generated Docs page.
Stories are type-checked but excluded from the published Rollup output.

The preview provides the application fonts, Tailwind/DoodleUI styles, React Query,
an in-memory router, Helmet, app name, and announcements.
Light/dark mode uses DoodleUI CSS variables and the document root class, so portal content
follows the selected theme. The toolbar button toggles directly between light and dark,
using the same themes addon as DoodleUI. The preview does not install an MUI theme provider or baseline.
Each story gets its own query client; queries do not retry or refetch on window focus.

Set the initial route when a component needs router context:

```tsx
parameters: {
    router: {
        initialEntries: ['/example'];
    }
}
```

API mocking, Redux state, and notification providers are not configured yet. For connected components, add story-specific
fixtures and providers before rendering them; the starter stories do not require a backend.
