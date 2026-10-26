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
missing or older than any file in Doodle UI's `src/`, `.storybook/`, or `package.json`. nginx serves the build at
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

## Development

```bash
yarn workspace dev-portal dev          # Vite dev server (serves the existing Storybook build as-is; no rebuild)
yarn workspace dev-portal test         # Vitest
yarn workspace dev-portal lint
yarn workspace dev-portal check-types
yarn workspace dev-portal format       # Prettier (check-format to only report)
```
