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
import { mkdir, mkdtemp, rm, symlink, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { findOrphanedHarDirectories } from './prune-har.mjs';

test('prune candidates exclude active and unmarked recordings', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'bh-har-prune-test-'));
    const configRoot = '/example/tests';
    try {
        for (const [directory, testId] of [
            ['active', 'still-here'],
            ['orphan', 'renamed-test'],
            ['other-suite', 'other-suite-test'],
        ]) {
            await mkdir(path.join(root, directory));
            await writeFile(
                path.join(root, directory, '.har-fixture.json'),
                JSON.stringify({ testId, rootDir: directory === 'other-suite' ? '/other/tests' : configRoot })
            );
            await writeFile(path.join(root, directory, 'recording.har'), '{}');
        }
        await mkdir(path.join(root, 'hand-maintained'));
        await writeFile(path.join(root, 'hand-maintained', 'recording.har'), '{}');

        assert.deepEqual(await findOrphanedHarDirectories(root, configRoot, new Set(['still-here'])), [
            path.join(root, 'orphan'),
        ]);
    } finally {
        await rm(root, { recursive: true, force: true });
    }
});

test('prune returns null for a missing root and ignores incomplete or unsafe candidates', async () => {
    const workspace = await mkdtemp(path.join(tmpdir(), 'bh-har-prune-test-'));
    const root = path.join(workspace, 'fixtures');
    const configRoot = '/example/tests';
    try {
        assert.equal(await findOrphanedHarDirectories(root, configRoot, new Set()), null);
        await mkdir(root);
        for (const name of ['invalid-json', 'missing-id', 'missing-recording']) {
            const directory = path.join(root, name);
            await mkdir(directory);
            const marker = name === 'missing-id' ? { rootDir: configRoot } : { rootDir: configRoot, testId: name };
            await writeFile(
                path.join(directory, '.har-fixture.json'),
                name === 'invalid-json' ? '{' : JSON.stringify(marker)
            );
            if (name !== 'missing-recording') await writeFile(path.join(directory, 'recording.har'), '{}');
        }
        const outside = path.join(workspace, 'outside');
        await mkdir(outside);
        await writeFile(
            path.join(outside, '.har-fixture.json'),
            JSON.stringify({ rootDir: configRoot, testId: 'old' })
        );
        await writeFile(path.join(outside, 'recording.har'), '{}');
        await symlink(outside, path.join(root, 'linked-directory'));
        await writeFile(path.join(root, 'ordinary-file'), '{}');
        assert.deepEqual(await findOrphanedHarDirectories(root, configRoot, new Set()), []);
    } finally {
        await rm(workspace, { recursive: true, force: true });
    }
});

for (const args of [['--unknown'], ['--root'], ['--root', '--delete']]) {
    test(`invalid prune arguments fail before listing tests: ${args.join(' ')}`, () => {
        const script = fileURLToPath(new URL('./prune-har.mjs', import.meta.url));
        const result = spawnSync(process.execPath, [script, ...args], { encoding: 'utf8' });
        assert.equal(result.status, 1);
        assert.match(result.stderr, args[0] === '--unknown' ? /Usage: yarn har:prune/ : /--root requires a directory/);
    });
}
