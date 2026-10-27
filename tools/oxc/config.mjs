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
import { createRequire } from "node:module";
import path from "node:path";

// Oxlint 1.89 applies only the entry config's categories, environment, settings,
// and ignore patterns. Rules, native plugins, overrides, and JS plugins inherit.
// Resolve JS plugins at their declaring config before flattening for fixtures.
export async function loadConfig(filename, ancestors = []) {
  const absolute = await fs.realpath(filename);
  assert(
    !ancestors.includes(absolute),
    `Circular Oxlint extends: ${[...ancestors, absolute].join(" -> ")}`,
  );
  const config = JSON.parse(await fs.readFile(absolute, "utf8"));
  const require = createRequire(absolute);
  const inherited = [];
  for (const reference of config.extends ?? []) {
    assert.equal(
      typeof reference,
      "string",
      `${absolute}: extends must contain paths`,
    );
    inherited.push(
      await loadConfig(path.resolve(path.dirname(absolute), reference), [
        ...ancestors,
        absolute,
      ]),
    );
  }
  const resolvePlugin = (plugin) => {
    if (typeof plugin === "string") {
      assert(
        !plugin.startsWith(".") && !path.isAbsolute(plugin),
        `${absolute}: relative JS plugins require an explicit name`,
      );
      return {
        name: plugin.replace(/(^|\/)eslint-plugin-/, "$1"),
        specifier: require.resolve(plugin),
      };
    }
    return { ...plugin, specifier: require.resolve(plugin.specifier) };
  };
  const rules = Object.assign(
    {},
    ...inherited.map((parent) => parent.rules),
    config.rules,
  );
  const plugins = [
    ...new Set([
      ...inherited.flatMap((parent) => parent.plugins),
      ...(config.plugins ?? ["typescript", "unicorn", "oxc"]),
    ]),
  ];
  const jsPlugins = new Map();
  for (const plugin of [
    ...inherited.flatMap((parent) => parent.jsPlugins),
    ...(config.jsPlugins ?? []).map(resolvePlugin),
  ]) {
    jsPlugins.set(plugin.name, plugin);
  }
  const overrides = [
    ...inherited.flatMap((parent) => parent.overrides),
    ...(config.overrides ?? []).map((override) => ({
      ...override,
      ...(override.jsPlugins
        ? { jsPlugins: override.jsPlugins.map(resolvePlugin) }
        : {}),
    })),
  ];
  const local = { ...config };
  delete local.extends;
  return {
    ...local,
    rules,
    plugins,
    jsPlugins: [...jsPlugins.values()],
    overrides,
  };
}
