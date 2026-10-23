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

import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';

const aliases = {
    'react-compat': 'react',
    'a11y-compat': 'jsx-a11y',
    'hooks-compat': 'react-hooks',
    'refresh-compat': 'react-refresh',
};

function legacyRule(code) {
    if (!code) return null;
    const match = /^(.*)\((.*)\)$/.exec(code);
    if (!match) return code;
    const [, plugin, rule] = match;
    if (plugin === 'eslint') return rule === 'no-unused-vars' ? '@typescript-eslint/no-unused-vars' : rule;
    return `${aliases[plugin] ?? (plugin === 'typescript' ? '@typescript-eslint' : plugin)}/${rule}`;
}

export function testLintParity(root, fixtures) {
    for (const fixture of fixtures) {
        test(`${fixture.workspace}: ${fixture.name}`, async () => {
            const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'bhce-oxc-fixture-'));
            try {
                const file = path.join(temporary, fixture.file);
                const source = fixture.source.replaceAll('eslint-disable', 'oxlint-disable');
                await fs.mkdir(path.dirname(file), { recursive: true });
                await fs.writeFile(file, source);
                const result = spawnSync(
                    process.execPath,
                    [
                        path.join(root, 'node_modules/oxlint/bin/oxlint'),
                        file,
                        '--config',
                        '.oxlintrc.json',
                        '--format',
                        'json',
                        '--report-unused-disable-directives-severity',
                        'error',
                        '--max-warnings',
                        '0',
                    ],
                    {
                        cwd: path.join(root, fixture.workspace),
                        encoding: 'utf8',
                        env: {
                            ...process.env,
                            NODE_OPTIONS: `${process.env.NODE_OPTIONS ?? ''} --require=${new URL('./no-eslint.cjs', import.meta.url).pathname}`,
                        },
                    }
                );
                assert.equal(result.error, undefined);
                const diagnostics = JSON.parse(result.stdout).diagnostics.map((diagnostic) => ({
                    rule: legacyRule(diagnostic.code),
                    severity: diagnostic.severity === 'error' ? 2 : 1,
                }));
                const sort = (items) => items.sort((a, b) => JSON.stringify(a).localeCompare(JSON.stringify(b)));
                assert.deepEqual(sort(diagnostics), sort([...fixture.diagnostics]), result.stdout + result.stderr);
                assert.equal(result.status, fixture.diagnostics.length ? 1 : 0);
            } finally {
                await fs.rm(temporary, { recursive: true, force: true });
            }
        });
    }
}
