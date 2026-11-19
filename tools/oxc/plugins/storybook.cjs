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

const path = require('node:path');
const { fixupPluginRules } = require('@eslint/compat');
const storybook = require('eslint-plugin-storybook');

const noUninstalledAddons = storybook.rules['no-uninstalled-addons'];

module.exports = fixupPluginRules({
    ...storybook,
    rules: {
        ...storybook.rules,
        'no-uninstalled-addons': {
            ...noUninstalledAddons,
            create(context) {
                // The upstream rule reads from the working directory by default.
                const workspaceOptions = {
                    packageJsonLocation: path.resolve(path.dirname(context.filename), '..', 'package.json'),
                };
                const workspaceContext = Object.create(context, {
                    options: { value: [workspaceOptions, ...context.options] },
                });
                return noUninstalledAddons.create(workspaceContext);
            },
        },
    },
});
