# Oxlint migration

Workspace `lint` scripts use Oxlint 1.88.0. Prettier, import organization, Vite,
TypeScript, and Yarn retain their existing roles. The tool version clears the
repository's three-day package-age gate as of October 20, 2026.

## Rule policy

Core and TypeScript rules are explicitly enabled from the existing ESLint policy;
new recommended presets and type-aware checking are not enabled. React, hooks,
accessibility, refresh, Storybook, and the custom `sx-props` rule use their existing
JavaScript implementations through small `@eslint/compat` adapters. Oxlint's JS
plugin support is still alpha, so updates need the parity checks below.

Two Yarn patches remove runtime ESLint imports from the React and Storybook
plugins without replacing their rule logic. Their package names and ESLint peer
requirements remain; installation reports missing ESLint peers intentionally.
The ESLint engine and TypeScript ESLint parser/plugin are removed. Prettier and
its dependencies remain installed.

Suppressions use `oxlint-disable` with the configured plugin aliases, such as
`hooks-compat/exhaustive-deps`. The core unused-variable rule is `no-unused-vars`.
Unused suppressions and warnings still fail workspace lint commands. Browser
globals use Oxlint's browser environment.

## Verification

From BHCE:

```sh
yarn lint
yarn test:lint
GOWORK=off go test ./packages/go/stbernard/analyzers/js ./packages/go/stbernard/analyzers
```

From BHE, `yarn test:lint` additionally runs the enterprise fixtures. The fixtures
capture original ESLint rule identities and severities; unused-directive severity
matches the existing CLI's error setting. Tests reject runtime imports of ESLint
from the JavaScript plugins. Browser-global fixtures cover both recognized and
unknown identifiers.

Stbernard accepts Oxlint JSON, converts workspace-relative filenames to paths,
and preserves severity and line locations. Exit 1 with findings is a normal lint
failure; missing/invalid output and other exit codes are tooling errors. Only
Oxlint JSON is accepted; ESLint output is rejected as invalid tooling output.

## Known differences and limits

- Oxlint does not report the existing declaration-merge generic parameters in
  DoodleUI's `tanstack-table.d.ts`, or the loop assignment in CE's `edge.self.ts`.
  Their now-unused suppressions were removed. These native detection differences
  remain; the focused fixtures do not prove equivalence for every possible input.
- Strict-mode parser errors replace some ESLint syntax-rule findings, so rule IDs
  and diagnostic wording may differ.
- Oxlint follows optional chains through TypeScript assertions. BHE's GraphView
  immediately dereferences the asserted label; removing its redundant optional
  chain retains the requirement that the graph node exists.
- BHE's existing `useMarkdownComponents.tsx` Fast Refresh warning is still present
  and still fails `--max-warnings 0`. It also failed the baseline ESLint run.
- BHE's direct `tsc --noEmit` check reports string-versus-number collection schedule
  ID errors against the installed shared declarations. This migration does not
  rebuild those declarations or change the affected contracts. An isolated check
  of the original BHE sources with the same installed declarations produces
  identical diagnostics.

The Oxc VS Code extension is recommended; the ESLint extension is disabled for
this workspace. Prettier remains the formatter. See the official
[editor setup](https://oxc.rs/docs/guide/usage/linter/editors.html) and
[JS plugin documentation](https://oxc.rs/docs/guide/usage/linter/js-plugins.html).
