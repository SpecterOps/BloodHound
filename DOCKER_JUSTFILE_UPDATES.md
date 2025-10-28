# Docker and Justfile Updates for pnpm Migration

This document summarizes the changes made to Docker files and the justfile to support pnpm instead of yarn.

## Files Updated

### 1. justfile

**Changed commands:**
- `yarn-local` → `pnpm-local` - Run pnpm commands in workspace root
- `yarn` → `pnpm` - Run pnpm commands and rebuild containers
- `build-js-client` - Now uses `pnpm build`
- `build-shared-ui` - Now uses `pnpm build`

**Usage examples:**
```bash
just pnpm install              # Install dependencies
just pnpm-local add lodash     # Add a dependency
just build-js-client           # Build client library
just build-shared-ui           # Build shared UI
```

### 2. dockerfiles/bloodhound.Dockerfile

**Changes in UI Build stage:**
- Added pnpm installation via corepack
- Replaced yarn.lock with pnpm-lock.yaml
- Replaced .yarnrc.yml and .yarn with .npmrc
- Added pnpm-workspace.yaml and nx.json to build context
- Changed `yarn install` to `pnpm install --frozen-lockfile`
- Changed `yarn build` to `pnpm build`

**Before:**
```dockerfile
COPY --parents constraints.pro package.json **/package.json yarn* .yarn*  ./
RUN yarn install
COPY --parents cmd/ui packages/javascript ./
RUN yarn build
```

**After:**
```dockerfile
RUN corepack enable && corepack prepare pnpm@9.0.0 --activate
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml* .npmrc ./
COPY --parents **/package.json ./
RUN pnpm install --frozen-lockfile
COPY --parents cmd/ui packages/javascript nx.json ./
RUN pnpm build
```

### 3. tools/docker-compose/ui.Dockerfile

**Changes:**
- Replaced yarn cache directory setup with pnpm directories
- Changed corepack to prepare pnpm instead of yarn
- Updated workspace file copies (pnpm-workspace.yaml, .npmrc, nx.json)
- Added project.json files for NX configuration
- Changed install location to workspace root
- Changed `yarn` to `pnpm install`

**Key differences:**
```dockerfile
# Old
RUN mkdir /.yarn && chmod -R go+w /.yarn
RUN corepack prepare yarn@stable --activate
COPY yarn.lock ./
COPY .yarnrc.yml ./
COPY .yarn ./.yarn
RUN yarn

# New
RUN mkdir /.local && chmod -R go+w /.local
RUN corepack prepare pnpm@9.0.0 --activate
COPY pnpm-workspace.yaml ./
COPY pnpm-lock.yaml* ./
COPY .npmrc ./
COPY nx.json ./
RUN pnpm install
```

### 4. docker-compose.dev.yml

**Changes:**
- Updated bh-ui service command from `yarn dev` to `pnpm dev`

**Before:**
```yaml
command: sh -c "yarn dev"
```

**After:**
```yaml
command: sh -c "pnpm dev"
```

### 5. docker-compose.watch.yml

**Changes:**
- Replaced `.yarnrc.yml` watch with `.npmrc`
- Added `pnpm-workspace.yaml` to watch list
- Added `nx.json` to watch list

This ensures Docker Compose watch mode rebuilds containers when pnpm or NX configuration changes.

### 6. .dockerignore

**Added entries:**
```
# Yarn (legacy)
.yarn/cache
.yarn/install-state.gz
.yarnrc.yml
yarn.lock

# pnpm
.pnpm-store
.pnpm-debug.log

# NX
.nx/cache
.nx/workspace-data
```

This prevents unnecessary files from being copied into Docker build context.

### 7. Documentation Updates

**DEVREADME.md:**
- Updated prerequisite from "Yarn v3.6" to "pnpm v9.0"
- Updated installation instructions

**cmd/ui/README.md:**
- Changed `yarn start` to `pnpm start` or `pnpm dev`
- Changed `yarn test` to `pnpm test`
- Changed `yarn build` to `pnpm build`
- Added NX usage section

**debian/README.md:**
- Replaced `yarnpkg` with `nodejs npm`
- Added `npm install -g pnpm@9.0.0` step

## Testing the Changes

### Local Development
```bash
# Clean up old dependencies
just reset-node-modules

# Install with pnpm
pnpm install

# Start dev environment
just bh-dev up
```

### Docker Build
```bash
# Clean build
just bh-clean-docker-build

# Build production container
just build-bhce-container linux/amd64 edge v5.0.0
```

### Watch Mode
```bash
# Start with watch mode for auto-rebuild
just bh-watch
```

## Important Notes

1. **yarn-workspaces.json** - This file is still present and used by the stbernard build tool. It has not been removed to maintain compatibility with existing Go tooling.

2. **Lockfile** - Make sure to commit `pnpm-lock.yaml` to the repository. Docker builds use `--frozen-lockfile` to ensure reproducible builds.

3. **Node Version** - All Dockerfiles use `node:22-alpine` which includes corepack for managing pnpm.

4. **Build Context** - The production Dockerfile now includes `nx.json` in the build context to support NX-based builds.

5. **Permissions** - The dev Dockerfile creates `/.local` directory with write permissions for pnpm's global store.

## Rollback Instructions

If you need to rollback to yarn:

1. Restore the original Dockerfiles from git
2. Restore the original justfile from git
3. Remove pnpm files: `rm -rf pnpm-lock.yaml .npmrc pnpm-workspace.yaml`
4. Restore yarn files from git
5. Run `yarn install`
6. Rebuild containers: `just bh-clean-docker-build`

