# Oxlint configuration

Workspace `lint` scripts use the repository-pinned Oxlint version. Prettier remains
the formatter, and TypeScript retains its existing build and type-checking roles.

## Shared policy

-   `configs/base.json` defines common JavaScript and TypeScript rules and the
    TypeScript file overrides.
-   `configs/react.json` extends the base with React, hooks, and accessibility rules.
-   Each workspace's `.oxlintrc.json` extends the appropriate shared config and
    declares its policy differences, JavaScript plugins, environment, and ignores.
    BHE consumes the shared configs through its `bhce` submodule.

`extends` paths resolve relative to the declaring config. Environments,
categories, settings, and ignore patterns stay in workspace configs because those
fields are not inherited. An explicit `plugins: []` inherits the shared native
plugin list without adding Oxlint's defaults.

The initial rule policy preserves the previous ESLint setup. Recommended category
presets and type-aware linting are not enabled. The Oxlint configs are the lint
policy going forward and can be improved directly.

## JavaScript rule adapters

The five files in `plugins/` adapt the existing React, hooks, accessibility, Fast
Refresh, and Storybook plugins through `@eslint/compat`. The custom `sx-props`
plugin is loaded directly. These preserve the existing rule implementations
without installing the ESLint engine or TypeScript ESLint parser/plugin.

Two Yarn patches remove runtime ESLint imports from the React and Storybook
plugins. Their ESLint peer requirements remain, so missing-ESLint peer warnings
are expected during installation. Oxlint's JavaScript plugin API is still alpha;
plugin changes should be checked with the affected workspace's lint command.

Suppressions use the configured rule aliases, for example:

```ts
// oxlint-disable-next-line hooks-compat/exhaustive-deps
```

Other aliases are `react-compat`, `a11y-compat`, and `refresh-compat`. The native
unused-variable rule is `no-unused-vars`. Warnings and unused suppressions fail
lint. Native `no-unused-vars` skips declaration files (`.d.ts`).

## Verification

Run `yarn lint` from either repository root. Stbernard consumes Oxlint JSON for
code-quality reporting and preserves severity and line locations. Its parser
checks remain normal Go unit tests.

The migration fixtures, ESLint policy snapshots, and disposable audit tool are
kept separately on `bfaulkner/oxlint-parity-evidence`. They document the initial
migration and do not run in the migration branch's routine CI.

The Oxc VS Code extension is recommended; Prettier remains the formatter. See the
[editor setup](https://oxc.rs/docs/guide/usage/linter/editors.html) and
[JavaScript plugin documentation](https://oxc.rs/docs/guide/usage/linter/js-plugins.html).
