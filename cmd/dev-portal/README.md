# BloodHound Developer Portal

The BloodHound Developer Portal is a locally hosted website for browsing documentation built from this monorepo. It is a
contributor tool, not part of the BloodHound product, and it runs outside the product's dev stack.

Its landing page shows a set of **Link Cards**. Each one leads to documentation produced from a **Doc Source** in the repo.

## Start, stop, and open

```bash
just dev-portal          # build the Dev Portal, then start the stack
just dev-portal up -d    # same, detached
just dev-portal down     # stop and remove the stack
just dev-portal logs -f  # any other arguments are passed to docker compose
```

Then open <http://dev-portal.localhost:8888>.

The stack (`docker-compose.dev-portal.yml`) runs a Traefik proxy on `127.0.0.1:8888` and an nginx file server. It uses
its own compose project and Traefik's file provider only, so it runs alongside `just bh-dev` without either stack
discovering or affecting the other.

## Themes

Components never hold raw colors, fonts, or radii. They use `dp-` Tailwind utilities (`bg-dp-canvas`,
`text-dp-text-muted`, `font-dp-display`, `rounded-dp-sm`, …), each backed by a `--dp-*` CSS custom property.

-   `src/theme/tokens.ts` lists the token names every theme must define and maps them to Tailwind and CSS properties.
-   `src/theme/themes.ts` holds the themes and `applyTheme`, which writes a theme's values onto `<html>`. On startup,
    `followSystemTheme` applies `themeForScheme[light|dark]` to match the operating system and follows later changes.

The `hero-*` tokens style the landing page's graph canvas; `marker` is the "you are here" color (the graph's source
node). DoodleUI's focus ring follows the `focus` token.

To add a theme:

1. Add a `Theme` object to `src/theme/themes.ts` and register it in `themes`. TypeScript reports any missing token.
2. Set `scheme` to `light` or `dark`; it switches DoodleUI components and native controls to match.
3. To show it by default, point `themeForScheme` at it, or call `applyTheme('<name>')`.
4. If it names a new typeface, import its `@fontsource` weights in `src/main.tsx`.

To add a token, add its name to the list in `src/theme/tokens.ts`, give it a value in every theme, and use it as
`dp-<token>` in components.

## Development

```bash
yarn workspace dev-portal dev          # Vite dev server
yarn workspace dev-portal test         # Vitest
yarn workspace dev-portal lint
yarn workspace dev-portal check-types
yarn workspace dev-portal format       # Prettier (check-format to only report)
```

## Manual running-stack check

Hosting is not covered by automated tests. After changing the compose file, nginx or Traefik config, or the recipe:

1. Run `just dev-portal up -d`.
2. Open <http://dev-portal.localhost:8888>: the landing page loads.
3. Open a Dev Portal deep link such as <http://dev-portal.localhost:8888/nope> and refresh: the Dev Portal's not-found
   page stays in place, and its link returns to the landing page.
4. With `just bh-dev` running, confirm both the product and the Dev Portal respond.
5. Run `just dev-portal down`.

Without a browser, the same checks can be made with
`curl -H 'Host: dev-portal.localhost' http://127.0.0.1:8888/<path>`.
