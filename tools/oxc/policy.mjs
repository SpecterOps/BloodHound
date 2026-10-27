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

import assert from "node:assert/strict";
import fs from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";
import test from "node:test";
import { loadConfig } from "./config.mjs";

const aliases = {
  "react-compat": "react",
  "a11y-compat": "jsx-a11y",
  "hooks-compat": "react-hooks",
  "refresh-compat": "react-refresh",
  typescript: "@typescript-eslint",
};

function ruleName(name) {
  if (
    ["no-unused-vars", "no-array-constructor", "no-loss-of-precision"].includes(
      name,
    )
  )
    return `@typescript-eslint/${name}`;
  const slash = name.indexOf("/");
  if (slash < 0) return name;
  return `${aliases[name.slice(0, slash)] ?? name.slice(0, slash)}${name.slice(slash)}`;
}

function rulePolicy(value) {
  const [severity, ...options] = Array.isArray(value) ? value : [value];
  const level = { off: 0, warn: 1, error: 2 }[severity] ?? severity;
  return [level, ...options];
}

function normalizeRules(rules, legacy = false) {
  const result = {};
  for (const [name, value] of Object.entries(rules)) {
    const canonical = ruleName(name);
    const policy = rulePolicy(value);
    if (legacy && canonical === "@typescript-eslint/no-unused-vars") {
      policy[1] = {
        args: "after-used",
        caughtErrors: "none",
        ignoreRestSiblings: false,
        ...(policy[1] ?? {}),
      };
    }
    if (policy[0] !== 0) result[canonical] = policy;
  }
  return result;
}

function matches(file, patterns = []) {
  // Oxlint applies overrides to dot directories too; path.matchesGlob uses
  // shell-style hidden-file exclusion. Escape leading dots on both operands
  // so its wildcard matching has Oxlint's dot-directory behavior.
  const visible = (value) => value.replace(/(^|\/)\./g, "$1__dot__");
  return patterns.some((pattern) =>
    path.matchesGlob(visible(file), visible(pattern)),
  );
}

export function testLintPolicy(root, baseline) {
  for (const workspace of [".", ...Object.keys(baseline.workspaces)]) {
    test(`${workspace}: no legacy lint engine dependency`, async () => {
      const packageFile = path.join(root, workspace, "package.json");
      const manifest = JSON.parse(await fs.readFile(packageFile, "utf8"));
      const require = createRequire(packageFile);
      for (const name of [
        "eslint",
        "@typescript-eslint/parser",
        "@typescript-eslint/eslint-plugin",
      ]) {
        for (const field of [
          "dependencies",
          "devDependencies",
          "optionalDependencies",
        ]) {
          assert.equal(
            manifest[field]?.[name],
            undefined,
            `${field}: ${name} must remain disposable`,
          );
        }
        assert.throws(
          () => require.resolve(name),
          { code: "MODULE_NOT_FOUND" },
          `${name} is installed in the repository`,
        );
      }
    });
  }
  for (const [workspace, snapshot] of Object.entries(baseline.workspaces)) {
    for (const [file, expected] of Object.entries(snapshot.scopes)) {
      test(`${workspace}: legacy policy for ${file}`, async () => {
        const config = await loadConfig(
          path.join(root, workspace, ".oxlintrc.json"),
        );
        // Recommended category presets would enable rules outside this inventory.
        for (const category of [
          "correctness",
          "suspicious",
          "pedantic",
          "perf",
          "style",
          "restriction",
          "nursery",
        ]) {
          assert.equal(
            rulePolicy(config.categories?.[category])[0],
            0,
            `${category} category must remain explicitly off`,
          );
        }
        const ancestors = file
          .split("/")
          .map((_, index, parts) => parts.slice(0, index + 1).join("/"));
        const ignored = ancestors.some((candidate) =>
          matches(candidate, config.ignorePatterns),
        );
        assert.equal(
          ignored,
          Boolean(expected.ignored),
          `${file}: ignored-file policy changed`,
        );
        if (ignored) return;
        const rules = { ...config.rules };
        for (const override of config.overrides ?? []) {
          if (
            matches(file, override.files) &&
            !matches(file, override.excludedFiles)
          )
            Object.assign(rules, override.rules);
        }
        const actual = normalizeRules(rules);
        const legacyRules =
          expected.rules ?? snapshot.rulesets[expected.ruleset];
        const legacy = normalizeRules(legacyRules, true);
        const differences = [];
        for (const [name, policy] of Object.entries(legacy)) {
          // Modules are strict-mode code. Oxc rejects these at parse time rather
          // than exposing ESLint's syntax-rule IDs; malformed-source fixtures
          // must continue to assert that rejection.
          if (name === "no-dupe-args" || name === "no-octal") continue;
          try {
            assert.deepEqual(actual[name], policy);
          } catch {
            differences.push({
              rule: name,
              eslint: policy,
              oxlint: actual[name] ?? null,
            });
          }
        }
        for (const [name, policy] of Object.entries(actual)) {
          if (!legacy[name])
            differences.push({ rule: name, eslint: null, oxlint: policy });
        }
        assert.deepEqual(
          differences,
          [],
          `Lost or changed ESLint policy: ${JSON.stringify(differences, null, 2)}`,
        );
      });
    }
  }
}
