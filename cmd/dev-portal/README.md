# BloodHound Developer Portal

The BloodHound Developer Portal is a locally hosted website for browsing documentation built from this monorepo.

## Start, stop, and open

```bash
just dev-portal          # build the Dev Portal, then start the stack
just dev-portal up -d    # same, detached
just dev-portal down     # stop and remove the stack
```

Then open <http://dev-portal.localhost:8888>.

On `up`, the recipe also builds the Doodle UI Storybook (`packages/javascript/doodle-ui/storybook-static`) when it is
missing or older than any file in the Doodle UI package (excluding `node_modules/`, `dist/`, `coverage/`, and
`storybook-static/`). nginx serves the build at
`/doodle-ui-storybook/`, and the portal's Storybook page (`#/storybook`) embeds it.

The stack (`docker-compose.dev-portal.yml`) runs a Traefik proxy on `127.0.0.1:8888` and an nginx file server. It uses
its own compose project and Traefik's file provider only, so it runs alongside `just bh-dev` without either stack
discovering or affecting the other.

## Linking to a Storybook story

The Storybook page keeps the selected story in the portal URL, so a reload or a shared link reopens the same story:

```text
http://dev-portal.localhost:8888/#/storybook?path=/story/button--primary
```

-   `path` is Storybook's own story parameter, the same value Storybook shows after `?path=` when run standalone.
-   On load, the page passes `path` to the embedded Storybook. Without it, Storybook opens its default story.
-   While you browse, the page reads the embedded Storybook's address every 500 ms and copies any new `path` into the
    portal URL. It replaces the current history entry, so the browser's Back button leaves the Storybook page rather
    than stepping through stories.
-   Editing `path` in the address bar while already on the page has no effect until you reload.

## Configuration (`conf.json`)

`conf.json` controls the landing page cards and which markdown files the Documentation page shows. It is read at build
time (and on every change under `yarn workspace dev-portal dev`); an invalid config fails the build.

-   `cards`: the landing page cards, in display order. Each has `accent` (`blue` or `orange`), `title`, `description`,
    `href`, and `icon` (`storybook` or `documentation`). The card number comes from its position.
-   `docs.include` / `docs.exclude`: glob patterns, relative to the repo root, selecting the docs. Folders whose name
    starts with a dot are skipped unless a pattern names them, as `.github/**/*.md` does. An include pattern that
    matches nothing prints a warning.
-   `docs.defaultPage`: the doc opened by `#/docs`. It must be one of the included docs.
-   `docs.repositoryUrl`: where links to repo files that are not included docs point, for example source files.

## Dependencies

Packages used only by the Dev Portal, not by `cmd/ui` or Doodle UI:

| Package      | Type | Purpose                                                                        |
| ------------ | ---- | ------------------------------------------------------------------------------ |
| `tinyglobby` | dev  | Resolves the `conf.json` doc patterns at build time (`portalContentPlugin.ts`) |

## Development

```bash
yarn workspace dev-portal dev          # Vite dev server (serves the existing Storybook build as-is; no rebuild)
yarn workspace dev-portal test         # Vitest
yarn workspace dev-portal lint
yarn workspace dev-portal check-types
yarn workspace dev-portal format       # Prettier (check-format to only report)
```
