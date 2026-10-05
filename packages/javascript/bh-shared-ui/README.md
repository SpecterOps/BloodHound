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
using the same themes addon as DoodleUI. The preview also installs the application's
MUI theme provider, palettes, typography, component overrides, z-index values, and
`CssBaseline`, so tables, legacy surfaces, and upload dialogs match the selected theme.
Each story gets its own query client; queries do not retry or refetch on window focus.

Set the initial route when a component needs router context:

```tsx
parameters: {
    router: {
        initialEntries: ['/example'];
    }
}
```

### Reusable component coverage

Stories prioritize controls and patterns that can be composed across screens, rather than
pages that are shared solely to keep CE and BHE in sync. Alongside the existing stories,
the reusable examples cover:

| Pattern               | Components                                                                                                                                                                                                     | Examples                                                                                                                         |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Data display          | `DataTable`, `DetailsAccordion`, `BaseColumnHeader`, `SortableHeader`                                                                                                                                          | Loading, empty data, rich cells, pagination, expansion, disabled items, sorting                                                  |
| Lists                 | `VirtualizedNodeList`, `InfiniteScrollingTable`, `InfiniteQueryFixedList`, `SearchResultItem`                                                                                                                  | Local pagination, scrolling, loading rows, read-only nodes, name fallback, search highlighting                                   |
| Selection and actions | `CardWithSwitch`, `DropdownSelector`, `DropdownTriggerContents`, `CreateMenu`, `GraphButton`, `GraphMenu`, `ManagedDatePicker`                                                                                 | Controlled selection, enabled/disabled settings, menu actions, date validation                                                   |
| File controls         | `FileDrop`, `FileStatusListItem`                                                                                                                                                                               | Single/multiple selection, disabled chooser, progress, completion, failure, remove/retry callbacks                               |
| Dialogs               | `DeleteConfirmationDialog`, `PasswordDialog`                                                                                                                                                                   | Confirmation challenge, loading/error, self-service/admin modes, required-field validation                                       |
| Text and feedback     | `HighlightedText`, `TextWithFallback`, `LabelWithCopy`, `MarkdownContent`, `StatusIndicator`, `ProcessingIndicator`, `LoadingOverlay`, `GraphProgress`, `NotificationSnackbar`, `GenericErrorBoundaryFallback` | Literal matching, missing values, copy affordances, formatted Markdown, statuses, loading, notification dismissal, caught errors |
| Layout and navigation | `PageWithTitle`, `SkipLink`                                                                                                                                                                                    | Actions/description, full width, keyboard skip navigation                                                                        |

Controlled examples update their displayed state and expose callbacks in Actions.
Interaction stories use `play` functions to exercise selection, pagination, sorting,
keyboard activation, validation, file selection, retry/removal, and notification dismissal.
Infinite lists use deterministic local data; file controls do not upload files; password
and deletion examples only log callbacks. Clipboard examples use the browser clipboard.

Feature pages, authentication/SSO workflows, static help text and icons within `AppIcon`
are not separate story candidates in this pass. Their reusable building blocks are covered
here or in DoodleUI, while connected feature coverage can use the HTTP mocking pattern below.

### Mocking connected components

`Components/FileIngest` mocks business logic at the HTTP boundary while keeping the
real component, React Query hooks, permissions, feature flags, notifications, filters,
and upload dialog. Stories cover the feature flag enabled/disabled paths and all three permission levels,
empty/loading/error states, file details and uploads. Filters and pagination can be explored in the main example.
Play functions exercise interactions and assert their results.

The preview starts MSW before rendering and resets handlers on every story load.
Unhandled `/api/` requests are errors. Use a **factory** rather than a shared array
to reset mutable state on navigation or interaction reruns:

```tsx
parameters: {
    msw: {
        handlers: () =>
            createFileIngestHandlers({
                feature: 'enabled',
                permission: 'read',
                history: 'populated',
            });
    }
}
```

`FileIngest.stories.mocks.ts` implements API response envelopes and filter query
operators, paginates 24 deterministic jobs, and keeps uploaded files in memory.
Only the selected story's handlers are installed. The `file-error` scenario fails
each file once so its real retry action can recover. Loading scenarios keep requests
pending until navigation.

The named `FileIngestExample` adds the notification provider and upload context/dialog.
`StoryNotifications` presents actual notifications with `NotificationSnackbar`, keeping
displayed-key bookkeeping per mount because `AppNotifications` uses module-level state.
Its authenticated wrapper loads `/self` before mounting `FileIngest`, matching the
application route and avoiding a false mount-time permission warning. Redux is
unnecessary for this page. Other connected stories should provide their actual contexts.

For connected components with different API scenarios, render only the primary story
inline in Docs with a `docs.page` template using `Title`, `Description`, and
`Primary` from `@storybook/blocks`. All scenarios remain available in the sidebar. Inline rendering inherits the toolbar theme, and keeping
one live example prevents scenarios from overwriting each other's MSW handlers.

The worker in `.storybook/public/` is MSW's unmodified generated asset. After an MSW
upgrade, regenerate it from this package directory with the standard command:

```sh
yarn exec msw init .storybook/public
```

Mocks simulate uploads and UI results; they do not parse collector data or ingest a
graph. The stories display current component behavior, including five-second history
polling and existing upload failure behavior.
