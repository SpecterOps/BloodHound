// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Optional migration tool. Never imported by the permanent tests.
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import fs from "node:fs/promises";
import { createRequire } from "node:module";
import path from "node:path";
import vm from "node:vm";
import { legacyRule } from "./parity.mjs";

const [rootArg, revision, baselineArg, ...flags] = process.argv.slice(2);
assert(
  rootArg && revision && baselineArg,
  "Usage: node baseline-audit.mjs REPO PRE_MIGRATION_REV TEMP_BASELINE_DIR [--record] [--sources]",
);
for (const flag of flags)
  assert(["--record", "--sources"].includes(flag), `Unknown flag: ${flag}`);
const root = path.resolve(rootArg);
const baseline = path.resolve(baselineArg);
assert(
  baseline !== root && !baseline.startsWith(root + path.sep),
  "Install the baseline outside the checkout",
);
const projectRequire = createRequire(path.join(root, "package.json"));
const executable = path.join(
  path.dirname(projectRequire.resolve("oxlint/package.json")),
  "bin/oxlint",
);
const baselineRequire = createRequire(path.join(baseline, "package.json"));
const { ESLint } = baselineRequire("eslint");
const { format } = projectRequire("prettier");
const json = (value) => format(JSON.stringify(value), { parser: "json" });
const manifest = JSON.parse(
  await fs.readFile(new URL("./baseline-packages.json", import.meta.url)),
);
for (const [name, expected] of Object.entries(manifest.dependencies)) {
  // Some packages do not export package.json.
  let directory = path.dirname(baselineRequire.resolve(name));
  while (true) {
    try {
      const pkg = JSON.parse(
        await fs.readFile(path.join(directory, "package.json")),
      );
      if (pkg.name === name) {
        assert.equal(pkg.version, expected, name);
        break;
      }
    } catch (error) {
      if (error.code !== "ENOENT") throw error;
    }
    const parent = path.dirname(directory);
    assert.notEqual(parent, directory, `Cannot find package metadata: ${name}`);
    directory = parent;
  }
}
// The disposable parser identifies real comments, so strings and template text
// containing directive spellings remain untouched. No parser is needed in CI.
const parser = baselineRequire("@typescript-eslint/parser");
const legacyConfig = baselineRequire("@eslint/eslintrc");
class BaselineFlatCompat extends legacyConfig.FlatCompat {
  constructor(options) {
    super({ ...options, resolvePluginsRelativeTo: baseline });
  }
  extends(...names) {
    // Preserve workspace-relative override patterns while loading the retired
    // shared config and plugins exclusively from the disposable installation.
    return super.extends(
      ...names.map((name) =>
        name === "prettier"
          ? baselineRequire.resolve("eslint-config-prettier")
          : name,
      ),
    );
  }
}

const aliases = {
  "hooks-compat/": "react-hooks/",
  "react-compat/": "react/",
  "a11y-compat/": "jsx-a11y/",
  "refresh-compat/": "react-refresh/",
  "typescript/": "@typescript-eslint/",
};
function translateComments(source, toOxlint, filePath) {
  if (!source.includes(toOxlint ? "eslint-" : "oxlint-")) return source;
  const ast = parser.parse(source, {
    comment: true,
    range: true,
    ecmaFeatures: { jsx: true },
    filePath,
    sourceType: "module",
    ecmaVersion: "latest",
  });
  for (const comment of [...ast.comments].reverse()) {
    const [start, end] = comment.range;
    let text = source.slice(start, end);
    if (!/^\/[*\/]\s*(?:eslint|oxlint)-(?:disable|enable)\b/.test(text))
      continue;
    if (toOxlint) {
      text = text
        .replace(/eslint-(disable|enable)/g, "oxlint-$1")
        .replaceAll("@typescript-eslint/no-unused-vars", "no-unused-vars");
      for (const [current, legacy] of Object.entries(aliases))
        text = text.replaceAll(legacy, current);
    } else {
      text = text.replace(/oxlint-(disable|enable)/g, "eslint-$1");
      for (const [current, legacy] of Object.entries(aliases))
        text = text.replaceAll(current, legacy);
      text = text.replace(
        /(?<![\w/-])no-unused-vars\b/g,
        "@typescript-eslint/no-unused-vars",
      );
    }
    source = source.slice(0, start) + text + source.slice(end);
  }
  return source;
}

const fixturePath = path.join(root, "tools/oxc/fixtures.json");
const fixtures = JSON.parse(await fs.readFile(fixturePath));
const engines = new Map();
const policy = {};
const git = (...args) =>
  execFileSync("git", ["-C", root, ...args], { encoding: "utf8" });
const resolvedRevision = git("rev-parse", revision).trim();
let revisions = {};
try {
  revisions = JSON.parse(
    await fs.readFile(path.join(root, "tools/oxc/baseline-revisions.json")),
  );
} catch (error) {
  if (error.code !== "ENOENT") throw error;
}
const profiles = [
  "src/parity.js",
  "src/parity.ts",
  "src/parity.tsx",
  "src/parity.test.tsx",
  "src/parity.d.ts",
  "src/parity.stories.tsx",
  ".storybook/main.ts",
  "dist/parity.ts",
  "coverage/parity.ts",
  "rollup.config.js",
];
for (const workspace of new Set(fixtures.map((item) => item.workspace))) {
  const configRevision = revisions[workspace] ?? resolvedRevision;
  const source = git(
    "show",
    `${configRevision}:${workspace}/eslint.config.cjs`,
  );
  const module = { exports: {} };
  const localRequire = createRequire(
    path.join(root, workspace, "package.json"),
  );
  vm.runInNewContext(
    source,
    {
      module,
      exports: module.exports,
      __dirname: path.join(root, workspace),
      require: (name) => {
        if (name === "eslint-plugin-sx-props") return localRequire(name);
        if (name === "@eslint/eslintrc")
          return { ...legacyConfig, FlatCompat: BaselineFlatCompat };
        return baselineRequire(name);
      },
    },
    { filename: `${workspace}/eslint.config.cjs` },
  );
  const engine = new ESLint({
    cwd: path.join(root, workspace),
    overrideConfigFile: true,
    overrideConfig: [
      { files: ["**/*.{js,jsx,ts,tsx,mjs,cjs,mts,cts}"] },
      ...module.exports,
      { linterOptions: { reportUnusedDisableDirectives: "error" } },
    ],
  });
  engines.set(workspace, engine);
  const scopes = {};
  const rulesets = {};
  for (const file of profiles) {
    const config = await engine.calculateConfigForFile(
      path.join(root, workspace, file),
    );
    if (!config) {
      scopes[file] = { ignored: true };
      continue;
    }
    const rules = Object.fromEntries(
      Object.entries(config.rules)
        .filter(([, value]) => value[0] !== 0)
        .sort(([a], [b]) => a.localeCompare(b)),
    );
    let ruleset = Object.keys(rulesets).find(
      (key) => JSON.stringify(rulesets[key]) === JSON.stringify(rules),
    );
    if (!ruleset) {
      ruleset = `rules-${Object.keys(rulesets).length + 1}`;
      rulesets[ruleset] = rules;
    }
    scopes[file] = { ruleset };
  }
  policy[workspace] = {
    revision: configRevision,
    configSha256: createHash("sha256").update(source).digest("hex"),
    scopes,
    rulesets,
  };
}
let differences = 0;
for (const fixture of fixtures) {
  process.chdir(path.join(root, fixture.workspace));
  const [result] = await engines
    .get(fixture.workspace)
    .lintText(fixture.source, { filePath: fixture.file, warnIgnored: false });
  const diagnostics = (result?.messages ?? []).map((message) => ({
    rule: message.ruleId,
    severity: message.severity,
    ...(fixture.checkLocation ? { line: message.line } : {}),
  }));
  const sorted = (values) =>
    [...values].sort((a, b) =>
      JSON.stringify(a).localeCompare(JSON.stringify(b)),
    );
  if (flags.includes("--record")) {
    fixture.diagnostics = diagnostics;
    const translated = translateComments(fixture.source, true, fixture.file);
    if (translated !== fixture.source) fixture.oxlintSource = translated;
    else delete fixture.oxlintSource;
  } else
    try {
      assert.deepEqual(sorted(diagnostics), sorted(fixture.diagnostics));
      const translated = translateComments(fixture.source, true, fixture.file);
      assert.equal(
        fixture.oxlintSource ?? fixture.source,
        translated,
        "Suppression translation snapshot is stale",
      );
    } catch {
      differences++;
      console.error(
        JSON.stringify({
          workspace: fixture.workspace,
          name: fixture.name,
          expected: fixture.diagnostics,
          eslint: diagnostics,
        }),
      );
    }
}
const snapshot = {
  revision: resolvedRevision,
  packages: manifest.dependencies,
  workspaces: policy,
};
if (!flags.includes("--record"))
  assert.deepEqual(
    JSON.parse(
      await fs.readFile(path.join(root, "tools/oxc/baseline-policy.json")),
    ),
    snapshot,
    "Resolved ESLint inventory differs from snapshot",
  );
if (flags.includes("--record")) {
  await fs.writeFile(fixturePath, await json(fixtures));
  await fs.writeFile(
    path.join(root, "tools/oxc/baseline-policy.json"),
    await json(snapshot),
  );
}
if (flags.includes("--sources")) {
  // Read the same current source with legacy suppression names. Never rewrite it.
  const report = [];
  const coverage = [];
  const normalize = (items) =>
    items.sort((a, b) => JSON.stringify(a).localeCompare(JSON.stringify(b)));
  const difference = (left, right) => {
    const remaining = [...right];
    return left.filter((item) => {
      const index = remaining.findIndex(
        (other) => JSON.stringify(item) === JSON.stringify(other),
      );
      if (index < 0) return true;
      remaining.splice(index, 1);
      return false;
    });
  };
  let visited = 0;
  for (const [workspace, engine] of engines) {
    const directory = path.join(root, workspace);
    process.chdir(directory);
    const previousPackage = JSON.parse(
      git("show", `${policy[workspace].revision}:${workspace}/package.json`),
    );
    const currentPackage = JSON.parse(
      await fs.readFile(path.join(directory, "package.json")),
    );
    const lintScope = (script, command) => {
      const match = new RegExp(`^${command} (src|\\.)(?:\\s|$)`).exec(script);
      assert(match, `${workspace}: unsupported lint command scope: ${script}`);
      return match[1];
    };
    const eslintScope = lintScope(previousPackage.scripts.lint, "eslint");
    const oxlintScope = lintScope(currentPackage.scripts.lint, "oxlint");
    const within = (file, scope) =>
      scope === "." || file.startsWith(`${scope}/`);
    const candidates = git("ls-files", "-z", "--", workspace)
      .split("\0")
      .filter((file) => /\.(?:[cm]?[jt]sx?)$/.test(file))
      .map((file) => path.relative(directory, path.join(root, file)));
    const tracked = [];
    for (const file of candidates) {
      // Git's index also lists files removed by this migration. This audit
      // compares both engines on the current files that still exist.
      try {
        if ((await fs.stat(path.join(directory, file))).isFile())
          tracked.push(file);
      } catch (error) {
        if (error.code !== "ENOENT") throw error;
      }
    }
    const eslintFiles = [];
    const eslintIgnored = [];
    for (const file of tracked.filter((file) => within(file, eslintScope))) {
      ((await engine.isPathIgnored(path.join(directory, file)))
        ? eslintIgnored
        : eslintFiles
      ).push(file);
    }
    const debug = spawnSync(
      process.execPath,
      [executable, oxlintScope, "--debug", "files"],
      {
        cwd: directory,
        encoding: "utf8",
        timeout: 60_000,
        maxBuffer: 32 * 1024 * 1024,
      },
    );
    assert(!debug.error && debug.status === 0, debug.stderr || debug.stdout);
    const discovered = debug.stdout
      .split(/\r?\n/)
      .filter(Boolean)
      .map((file) => path.relative(directory, path.resolve(directory, file)));
    const trackedSet = new Set(tracked);
    const oxlintFiles = discovered.filter((file) => trackedSet.has(file));
    const eslintSet = new Set(eslintFiles);
    const oxlintSet = new Set(oxlintFiles);
    const eslintOnlyFiles = eslintFiles.filter((file) => !oxlintSet.has(file));
    const oxlintOnlyFiles = oxlintFiles.filter((file) => !eslintSet.has(file));
    coverage.push({
      workspace,
      eslintScope,
      oxlintScope,
      tracked: tracked.length,
      eslintIncluded: eslintFiles.length,
      eslintIgnored,
      oxlintIncluded: oxlintFiles.length,
      eslintOnly: eslintOnlyFiles,
      oxlintOnly: oxlintOnlyFiles,
      oxlintUntracked: discovered.filter((file) => !trackedSet.has(file)),
    });
    differences += eslintOnlyFiles.length + oxlintOnlyFiles.length;
    // Compare diagnostics over the union: a newly included or excluded file
    // must remain visible in both the coverage report and the findings.
    const files = [...new Set([...eslintFiles, ...oxlintFiles])];
    if (!files.length) continue;
    const execution = oxlintFiles.length
      ? spawnSync(
          process.execPath,
          [
            executable,
            ...oxlintFiles,
            "--format",
            "json",
            "--report-unused-disable-directives-severity",
            "error",
            "--max-warnings",
            "0",
          ],
          {
            cwd: directory,
            encoding: "utf8",
            timeout: 60_000,
            maxBuffer: 32 * 1024 * 1024,
          },
        )
      : { status: 0, stdout: '{"diagnostics":[],"number_of_files":0}' };
    assert(
      !execution.error && [0, 1].includes(execution.status),
      execution.stderr || execution.stdout,
    );
    const oxlint = JSON.parse(execution.stdout);
    assert(Array.isArray(oxlint.diagnostics));
    assert.equal(
      oxlint.number_of_files,
      oxlintFiles.length,
      `${workspace}: explicit-file traversal differs from actual CLI coverage`,
    );
    for (const workspaceFile of files) {
      const file = path.join(workspace, workspaceFile);
      let source = await fs.readFile(path.join(root, file), "utf8");
      source = translateComments(source, false, file);
      const [result] = eslintSet.has(workspaceFile)
        ? await engine.lintText(source, {
            filePath: path.join(root, file),
            warnIgnored: false,
          })
        : [];
      visited++;
      const eslintDiagnostics = normalize(
        (result?.messages ?? []).map((item) => ({
          rule: item.ruleId,
          severity: item.severity,
          line: item.line,
        })),
      );
      const oxlintDiagnostics = normalize(
        oxlint.diagnostics
          .filter(
            (item) =>
              path.resolve(root, workspace, item.filename) ===
              path.join(root, file),
          )
          .map((item) => ({
            rule: legacyRule(item.code),
            severity: item.severity === "error" ? 2 : 1,
            line: item.labels?.[0]?.span?.line,
          })),
      );
      const eslintOnly = difference(eslintDiagnostics, oxlintDiagnostics);
      const oxlintOnly = difference(oxlintDiagnostics, eslintDiagnostics);
      if (eslintOnly.length || oxlintOnly.length)
        report.push({ file, eslintOnly, oxlintOnly });
    }
  }
  const reportPath = path.join(
    baseline,
    `source-report-${path.basename(root)}.json`,
  );
  await fs.writeFile(
    reportPath,
    JSON.stringify(
      {
        oxlintVersion: projectRequire("oxlint/package.json").version,
        visited,
        coverage,
        differences: report,
      },
      null,
      2,
    ) + "\n",
  );
  console.log(
    `Same-source differential report: ${reportPath} (${visited} files, ${report.length} differing files)`,
  );
  differences += report.length;
}
console.log(
  `${fixtures.length} fixtures checked; ${differences} differences; ${Object.keys(policy).length} workspace policies resolved`,
);
process.exitCode = differences ? 1 : 0;
