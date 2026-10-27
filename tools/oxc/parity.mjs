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
import { spawnSync } from "node:child_process";
import fs from "node:fs/promises";
import os from "node:os";
import { createRequire } from "node:module";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import { loadConfig } from "./config.mjs";

const aliases = {
  "react-compat": "react",
  "a11y-compat": "jsx-a11y",
  "hooks-compat": "react-hooks",
  "refresh-compat": "react-refresh",
};

export function legacyRule(code) {
  if (!code) return null;
  const match = /^(.*)\((.*)\)$/.exec(code);
  if (!match) return code;
  const [, plugin, rule] = match;
  if (plugin === "eslint")
    return [
      "no-unused-vars",
      "no-array-constructor",
      "no-loss-of-precision",
    ].includes(rule)
      ? `@typescript-eslint/${rule}`
      : rule;
  return `${aliases[plugin] ?? (plugin === "typescript" ? "@typescript-eslint" : plugin)}/${rule}`;
}

const sorted = (items) =>
  [...items].sort((a, b) => JSON.stringify(a).localeCompare(JSON.stringify(b)));

export function testLintParity(root, fixtures) {
  const require = createRequire(path.join(root, "package.json"));
  const executable = path.join(
    path.dirname(require.resolve("oxlint/package.json")),
    "bin/oxlint",
  );
  const workspaces = Map.groupBy(fixtures, (fixture) => fixture.workspace);
  for (const [workspace, cases] of workspaces) {
    test(`${workspace}: lint parity`, async (context) => {
      const temporary = await fs.mkdtemp(
        path.join(os.tmpdir(), "bhce-oxc-fixture-"),
      );
      try {
        const workspaceDirectory = path.join(root, workspace);
        const configPath = path.join(workspaceDirectory, ".oxlintrc.json");
        // Flatten inheritance before relocating the fixture config. Relative JS
        // plugins are resolved from the config file that declared each one.
        const config = await loadConfig(configPath);
        await fs.writeFile(
          path.join(temporary, ".oxlintrc.json"),
          JSON.stringify(config),
        );
        // Some rules, including Storybook's addon check, intentionally read
        // package.json relative to process.cwd() rather than the source file.
        await fs.copyFile(
          path.join(workspaceDirectory, "package.json"),
          path.join(temporary, "package.json"),
        );
        const files = [];
        const expectedFiles = new Map();
        const names = new Set();
        for (const [index, fixture] of cases.entries()) {
          assert.equal(typeof fixture.name, "string");
          assert(
            !names.has(fixture.name),
            `Duplicate fixture: ${fixture.name}`,
          );
          names.add(fixture.name);
          assert.equal(typeof fixture.source, "string", fixture.name);
          assert(Array.isArray(fixture.diagnostics), fixture.name);
          if ("oxlintDiagnostics" in fixture) {
            assert(Array.isArray(fixture.oxlintDiagnostics), fixture.name);
            assert.equal(
              typeof fixture.differenceReason,
              "string",
              fixture.name,
            );
            assert(
              fixture.differenceReason.trim().length > 0,
              `Missing difference explanation: ${fixture.name}`,
            );
            assert.notDeepEqual(
              sorted(fixture.oxlintDiagnostics),
              sorted(fixture.diagnostics),
              `Redundant exception: ${fixture.name}`,
            );
          } else {
            assert.equal(
              fixture.differenceReason,
              undefined,
              `Difference has no explicit expectation: ${fixture.name}`,
            );
          }
          assert(!path.isAbsolute(fixture.file), fixture.name);
          // Root-anchored Storybook overrides require the real relative path.
          // The .ts and .js fixtures use separate files; all direct paths must
          // be unique so one case cannot overwrite another case's source.
          const directory = fixture.file.startsWith(".storybook/")
            ? temporary
            : path.join(temporary, String(index));
          const file = path.resolve(directory, fixture.file);
          assert(
            file.startsWith(directory + path.sep),
            `Fixture escapes its directory: ${fixture.name}`,
          );
          assert(
            !expectedFiles.has(file),
            `Duplicate fixture path: ${fixture.file}`,
          );
          files.push(file);
          expectedFiles.set(file, []);
          await fs.mkdir(path.dirname(file), { recursive: true });
          await fs.writeFile(file, fixture.oxlintSource ?? fixture.source);
        }
        const result = spawnSync(
          process.execPath,
          [
            executable,
            ...files.map((file) => path.relative(temporary, file)),
            "--config",
            ".oxlintrc.json",
            "--format",
            "json",
            "--report-unused-disable-directives-severity",
            "error",
            "--max-warnings",
            "0",
          ],
          {
            cwd: temporary,
            encoding: "utf8",
            timeout: 60_000,
            maxBuffer: 16 * 1024 * 1024,
            env: {
              ...process.env,
              NODE_OPTIONS: `${process.env.NODE_OPTIONS ?? ""} --require=${JSON.stringify(fileURLToPath(new URL("./no-eslint.cjs", import.meta.url)))}`,
            },
          },
        );
        const output = result.stdout + result.stderr;
        assert.equal(result.error, undefined, output);
        assert.equal(result.signal, null, output);
        assert(
          [0, 1].includes(result.status),
          `Oxlint tooling failure: ${output}`,
        );
        let report;
        try {
          report = JSON.parse(result.stdout);
        } catch {
          assert.fail(`Oxlint returned invalid JSON: ${output}`);
        }
        assert(Array.isArray(report.diagnostics), output);
        for (const diagnostic of report.diagnostics) {
          assert(["error", "warning"].includes(diagnostic.severity), output);
          assert.equal(typeof diagnostic.filename, "string", output);
          const file = path.resolve(temporary, diagnostic.filename);
          assert(
            expectedFiles.has(file),
            `Unexpected diagnostic file: ${diagnostic.filename}`,
          );
          expectedFiles.get(file).push(diagnostic);
        }
        assert.equal(result.status, report.diagnostics.length ? 1 : 0, output);
        // A successful batch must have visited every requested file, including
        // clean fixtures. Otherwise an ignored file could masquerade as a pass.
        assert.equal(report.number_of_files, files.length, output);
        for (const [index, fixture] of cases.entries()) {
          await context.test(fixture.name, () => {
            const diagnostics = expectedFiles
              .get(files[index])
              .map((diagnostic) => {
                const value = {
                  rule: legacyRule(diagnostic.code),
                  severity: diagnostic.severity === "error" ? 2 : 1,
                };
                // Opt in to location assertions for fixtures that need them.
                if (
                  (fixture.oxlintDiagnostics ?? fixture.diagnostics).some(
                    (item) => "line" in item,
                  )
                ) {
                  value.line = diagnostic.labels?.[0]?.span?.line;
                }
                return value;
              });
            assert.deepEqual(
              sorted(diagnostics),
              sorted(fixture.oxlintDiagnostics ?? fixture.diagnostics),
              JSON.stringify(
                {
                  source: fixture.source,
                  diagnostics: expectedFiles.get(files[index]),
                },
                null,
                2,
              ),
            );
          });
        }
      } finally {
        await fs.rm(temporary, { recursive: true, force: true });
      }
    });
  }
}
