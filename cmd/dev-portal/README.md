# BloodHound Developer Portal

The BloodHound Developer Portal is a locally hosted website for browsing documentation built from this monorepo.

## Start, stop, and open

```bash
just dev-portal          # build the Dev Portal, then start the stack
just dev-portal up -d    # same, detached
just dev-portal down     # stop and remove the stack
```

Then open <http://dev-portal.localhost:8888>.

The stack (`docker-compose.dev-portal.yml`) runs a Traefik proxy on `127.0.0.1:8888` and an nginx file server. It uses
its own compose project and Traefik's file provider only, so it runs alongside `just bh-dev` without either stack
discovering or affecting the other.

## Development

```bash
yarn workspace dev-portal dev          # Vite dev server
yarn workspace dev-portal test         # Vitest
yarn workspace dev-portal lint
yarn workspace dev-portal check-types
yarn workspace dev-portal format       # Prettier (check-format to only report)
```
