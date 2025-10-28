# PNPM and NX Migration Guide

This document describes the migration from Yarn workspaces to pnpm + NX monorepo for the BloodHound CE JavaScript packages.

## What Changed

### Root Configuration

1. **package.json**
   - Changed `packageManager` from `yarn@3.5.1` to `pnpm@9.0.0`
   - Replaced `workspaces` configuration (now in `pnpm-workspace.yaml`)
   - Converted `resolutions` to `pnpm.overrides`
   - Updated all scripts to use NX commands (`nx run-many -t <target>`)
   - Added NX dependencies: `nx`, `@nx/vite`, `@nx/js`

2. **New Files Created**
   - `.npmrc` - pnpm configuration (hoisting, peer dependencies)
   - `pnpm-workspace.yaml` - workspace package definitions
   - `nx.json` - NX workspace configuration with caching and task orchestration
   - `cmd/ui/project.json` - NX project configuration for the UI app
   - `packages/javascript/bh-shared-ui/project.json` - NX project configuration for shared UI library
   - `packages/javascript/js-client-library/project.json` - NX project configuration for client library

### Package Updates

All three packages (`cmd/ui`, `packages/javascript/bh-shared-ui`, `packages/javascript/js-client-library`) were updated:

1. Removed `packageManager` field (managed at root)
2. Removed `installConfig.hoistingLimits` (handled by `.npmrc`)
3. Changed `yarn` commands to `pnpm` in build scripts
4. Simplified `check-types` script in `cmd/ui` (dependency building now handled by NX)

### Docker Files Updated

1. **`dockerfiles/bloodhound.Dockerfile`**
   - Replaced yarn with pnpm in UI build stage
   - Added pnpm installation via corepack
   - Updated COPY commands for pnpm workspace files
   - Added nx.json to build context

2. **`tools/docker-compose/ui.Dockerfile`**
   - Replaced yarn with pnpm
   - Updated workspace file copies (pnpm-workspace.yaml, .npmrc, etc.)
   - Added project.json files for NX configuration
   - Changed install command to `pnpm install`

3. **`docker-compose.dev.yml`**
   - Updated bh-ui service command from `yarn dev` to `pnpm dev`

4. **`docker-compose.watch.yml`**
   - Replaced `.yarnrc.yml` watch with `.npmrc`
   - Added `pnpm-workspace.yaml` and `nx.json` to watch list

5. **`.dockerignore`**
   - Added pnpm-specific ignores (.pnpm-store, .pnpm-debug.log)
   - Added NX cache ignores (.nx/cache, .nx/workspace-data)
   - Added legacy yarn ignores for cleanup

### Justfile Updated

1. Renamed `yarn-local` to `pnpm-local`
2. Renamed `yarn` to `pnpm`
3. Updated `build-js-client` to use `pnpm build`
4. Updated `build-shared-ui` to use `pnpm build`

## Migration Steps

### Prerequisites

1. Install pnpm globally:
   ```bash
   npm install -g pnpm@9.0.0
   ```

2. Remove existing node_modules and lock files:
   ```bash
   rm -rf node_modules
   rm -rf cmd/ui/node_modules
   rm -rf packages/javascript/*/node_modules
   rm yarn.lock
   rm -rf .yarn
   rm -rf .pnpm-store
   ```

   Or use the justfile command:
   ```bash
   just reset-node-modules
   ```

### Installation

1. Install dependencies with pnpm:
   ```bash
   pnpm install
   ```

## Available Commands

All commands now use NX for task orchestration and caching:

### Development
```bash
pnpm dev          # Start UI dev server
pnpm debug        # Start UI dev server in debug mode
pnpm start        # Start UI dev server
pnpm preview      # Preview production build
```

### Building
```bash
pnpm build        # Build all packages (with dependency graph awareness)
```

### Testing & Quality
```bash
pnpm test         # Run tests in all packages
pnpm check-types  # Type check all packages
pnpm lint         # Lint all packages
pnpm format       # Format all packages
pnpm check-format # Check formatting in all packages
```

### NX-Specific Commands

Run a specific target for a specific project:
```bash
nx run bloodhound-ui:dev
nx run bh-shared-ui:build
nx run js-client-library:test
```

Run a target for all projects:
```bash
nx run-many -t build
nx run-many -t test
```

Run affected projects only (based on git changes):
```bash
nx affected -t build
nx affected -t test
```

View the project graph:
```bash
nx graph
```

## Benefits of NX

1. **Computation Caching**: NX caches task results, so repeated builds/tests are instant
2. **Task Orchestration**: Automatically runs tasks in the correct order based on dependencies
3. **Affected Commands**: Only run tasks for projects affected by your changes
4. **Parallel Execution**: Runs independent tasks in parallel for faster builds
5. **Project Graph**: Visualize dependencies between projects

## Benefits of pnpm

1. **Disk Space**: Shared dependency storage across all projects
2. **Speed**: Faster installation than npm/yarn
3. **Strict**: Better at catching dependency issues
4. **Workspace Protocol**: Native workspace support with `workspace:*` protocol

## Workspace Dependencies

The workspace dependencies are automatically resolved:
- `cmd/ui` depends on `bh-shared-ui` and `js-client-library`
- `bh-shared-ui` depends on `js-client-library`

NX automatically builds dependencies before dependent projects.

## Docker Usage

After the migration, Docker builds will use pnpm:

### Development
```bash
just bh-dev up          # Start dev environment
just bh-watch           # Use watch mode for auto-rebuild
```

### Production Build
```bash
just build-bhce-container linux/amd64 edge v5.0.0
```

The Docker builds now:
- Use pnpm instead of yarn for faster, more reliable installs
- Leverage NX for build orchestration
- Cache dependencies more efficiently

## Troubleshooting

### Peer Dependency Warnings
If you see peer dependency warnings, they should be auto-installed due to `auto-install-peers=true` in `.npmrc`.

### Cache Issues
If you encounter stale cache issues:
```bash
nx reset
```

### Hoisting Issues
The `.npmrc` is configured with `shamefully-hoist=true` for compatibility. If you encounter module resolution issues, this can be adjusted.

### Docker Build Issues
If Docker builds fail:
1. Clear Docker build cache: `docker builder prune`
2. Rebuild without cache: `just bh-clean-docker-build`
3. Ensure pnpm-lock.yaml is committed to the repository

## Next Steps

After migration:
1. Test all build commands
2. Test all development workflows
3. Update CI/CD pipelines to use pnpm and NX
4. Consider adding NX Cloud for distributed caching (optional)

## Additional Documentation

- **DOCKER_JUSTFILE_UPDATES.md** - Detailed documentation of Docker and justfile changes
- **.gitignore.pnpm** - Suggested additions to .gitignore for pnpm and NX files

## Summary of All Changes

### Configuration Files Created
- `.npmrc` - pnpm configuration
- `pnpm-workspace.yaml` - Workspace package definitions
- `nx.json` - NX workspace configuration
- `cmd/ui/project.json` - NX project config for UI
- `packages/javascript/bh-shared-ui/project.json` - NX project config for shared UI
- `packages/javascript/js-client-library/project.json` - NX project config for client library

### Files Modified
- `package.json` (root) - Updated for pnpm and NX
- `cmd/ui/package.json` - Removed yarn-specific config
- `packages/javascript/bh-shared-ui/package.json` - Removed yarn-specific config
- `packages/javascript/js-client-library/package.json` - Removed yarn-specific config
- `justfile` - Updated yarn commands to pnpm
- `dockerfiles/bloodhound.Dockerfile` - Updated for pnpm
- `tools/docker-compose/ui.Dockerfile` - Updated for pnpm
- `docker-compose.dev.yml` - Updated bh-ui command to use pnpm
- `docker-compose.watch.yml` - Updated watch paths
- `.dockerignore` - Added pnpm/NX ignores
- `DEVREADME.md` - Updated prerequisites
- `cmd/ui/README.md` - Updated commands and added NX section
- `debian/README.md` - Updated build instructions

### Files to Keep (Legacy)
- `yarn-workspaces.json` - Still used by stbernard build tool

