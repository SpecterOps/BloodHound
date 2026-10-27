# Oxlint migration evidence

Workspace `lint` scripts use Oxlint 1.89.0. Prettier, import organization, Vite,
TypeScript, and Yarn retain their existing roles.

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

The evidence suite also checks the active rule inventory against the resolved
ESLint policy for JavaScript, TypeScript, TSX, tests, declarations, stories, and
Storybook configuration files, ignored output directories, and root Rollup configuration.
It checks both lost and newly enabled rules, preserves explicit options, and
asserts that the retired engine/parser packages are neither declared nor
resolvable. Explicit exceptions document reviewed differences.
CI runs `yarn test:lint` after dependency installation. Neither the tests nor CI
install or invoke the ESLint engine: the oracle is the checked-in baseline, not a
second linter retained in the project.

Stbernard accepts Oxlint JSON, converts workspace-relative filenames to paths,
and preserves severity and line locations. Exit 1 with findings is a normal lint
failure; missing/invalid output and other exit codes are tooling errors. Only
Oxlint JSON is accepted; ESLint output is rejected as invalid tooling output.

## Disposable ESLint audit

The optional `baseline-audit.mjs` tool independently recreates the original
ESLint expectations. Its package template is audit data, not a Yarn workspace or
an addition to the project's dependencies. Run it manually before accepting the
migration or deliberately changing baseline fixtures.

From the BHCE repository root, create a directory outside the checkout and install
the exact direct dependency versions recorded in `baseline-packages.json`:

```sh
lint_baseline=$(mktemp -d)
cp tools/oxc/baseline-packages.json "$lint_baseline/package.json"
cp tools/oxc/baseline-install-lock.json "$lint_baseline/package-lock.json"
npm ci --prefix "$lint_baseline" --ignore-scripts --legacy-peer-deps \
  --no-audit --no-fund
node tools/oxc/baseline-audit.mjs . \
  5760db071774bd0b6b7dcad2446091aaa5f8d5c4 "$lint_baseline"
```

The audit reads the original configurations from the specified Git revision.
`baseline-revisions.json` supplies the original Dev Portal commit, because that
workspace was introduced after the main migration baseline. Each policy records
its source revision and config hash. The audit
resolves their presets and overrides, and compares ESLint's diagnostics with the
fixture expectations. The revision must be available locally; a shallow CI
checkout is not used for this audit. The tool checks the direct dependency
versions before running. The saved `baseline-install-lock.json` pins transitive
dependencies as well; `npm ci` recreates that tree without updating the lockfile.

Add `--sources` to run both engines over the current workspace sources and write
`source-report-<repository>.json` into the temporary directory. The report records
actual old/current CLI scopes and included file lists, coverage differences,
untracked files seen by Oxlint, and ESLint-only/Oxlint-only diagnostics, normalized by rule
alias, severity, and line. Suppression names are translated in memory; source
files are not rewritten. Inspect the report and investigate every ESLint-only
finding. This source pass supplements the deliberately invalid fixtures, because
clean source alone cannot demonstrate that a rule still detects violations.

Only add `--record` when intentionally capturing original ESLint expectations. It
writes `fixtures.json` and `baseline-policy.json` under the target repository's
`tools/oxc` directory. Review the diff, run `yarn test:lint`, and investigate any
Oxlint mismatch. Do not regenerate expectations merely to make a failing test
pass. Do not use Oxlint results to populate the ESLint baseline.

From BHE, use `bhce/tools/oxc/baseline-audit.mjs` and the BHE pre-migration revision
documented in [the enterprise guide](../../../tools/oxc/README.md). Audit BHCE
separately with the BHCE revision above. Both use the shared package template.

After reviewing reports, delete the disposable environment:

```sh
rm -rf -- "$lint_baseline"
unset lint_baseline
```

Keep only the fixture sources, baseline diagnostics, resolved policy, provenance,
and documented exceptions in the repository. The temporary ESLint installation is deleted after the audit.
`source-audit.json` keeps the reviewed source comparison as migration evidence;
it is not a live claim about later source changes.

## Known differences and limits

-   Native `no-unused-vars` ignores `.d.ts` files entirely, including the existing
    declaration-merge generic parameters in DoodleUI's `tanstack-table.d.ts`.
    The old unused-generic suppression was removed. Fixtures preserve the ESLint
    diagnostics and specify separate `oxlintDiagnostics` with a required
    `differenceReason` for this exception. This documents the observed generic
    gap; it does not establish complete unused-variable coverage in declarations
    or equivalence for every possible input or autofix.
-   Strict-mode parser errors replace some ESLint syntax-rule findings, so rule IDs
    and diagnostic wording may differ.
-   Oxlint follows optional chains through TypeScript assertions. BHE's GraphView
    immediately dereferences the asserted label; removing its redundant optional
    chain retains the requirement that the graph node exists.
-   BHE's direct `tsc --noEmit` check reports string-versus-number collection schedule
    ID errors against the installed shared declarations. This migration does not
    rebuild those declarations or change the affected contracts. An isolated check
    of the original BHE sources with the same installed declarations produces
    identical diagnostics.

The Oxc VS Code extension is recommended; the ESLint extension is disabled for
this workspace. Prettier remains the formatter. See the official
[editor setup](https://oxc.rs/docs/guide/usage/linter/editors.html) and
[JS plugin documentation](https://oxc.rs/docs/guide/usage/linter/js-plugins.html).

## Evidence files

The migration branch treats its Oxlint configs as the lint policy. This branch
retains the one-time comparison tooling for reviewing that transition.

| File                                                   | Purpose                                                                                   |
| ------------------------------------------------------ | ----------------------------------------------------------------------------------------- |
| `fixtures.json`                                        | Source examples and original ESLint diagnostics, including explicit known differences.    |
| `parity.mjs`, `parity.test.mjs`                        | Run the examples through Oxlint and compare rule IDs, severities, and selected locations. |
| `config.mjs`                                           | Resolve shared configs and plugin paths for the isolated checks.                          |
| `policy.mjs`                                           | Compare inherited rules, options, and file scopes with the captured ESLint policy.        |
| `baseline-policy.json`                                 | Resolved original ESLint rules with Git and dependency provenance.                        |
| `baseline-audit.mjs`                                   | Recreate the old ESLint oracle and compare fixtures or current sources.                   |
| `baseline-packages.json`, `baseline-install-lock.json` | Pin the disposable ESLint installation outside the checkout.                              |
| `baseline-revisions.json`                              | Identify the original config revision for workspaces added after the main baseline.       |
| `source-audit.json`                                    | Migration-time findings and file-coverage comparison.                                     |
| `no-eslint.cjs`                                        | Reject runtime engine/parser imports while fixture plugins run.                           |

Both repositories use `bfaulkner/oxlint-parity-evidence`. Initialize the nested
repository, install dependencies, then run `yarn test:lint` from BHE to check both
repositories, or from BHCE for its checks alone.
