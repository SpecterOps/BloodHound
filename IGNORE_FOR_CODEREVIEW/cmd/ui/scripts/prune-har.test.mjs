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
import { access, mkdir, mkdtemp, rm, symlink, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import process from 'node:process';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { findOrphanedHarFiles } from './prune-har.mjs';

const script = fileURLToPath(new URL('./prune-har.mjs', import.meta.url));

test('prunes only unpaired HAR files, including nested files, without following symlinks', async () => {
    const workspace = await mkdtemp(path.join(tmpdir(), 'bh-har-prune-test-'));
    const root = path.join(workspace, 'tests/e2e');
    try {
        assert.equal(await findOrphanedHarFiles(root), null);
        await mkdir(path.join(root, 'nested'), { recursive: true });
        await writeFile(path.join(root, 'active.har'), '{}');
        await writeFile(path.join(root, 'active.spec.ts'), '');
        await writeFile(path.join(root, 'orphan.har'), '{}');
        await writeFile(path.join(root, 'nested/recording.har'), '{}');
        await writeFile(path.join(workspace, 'outside.har'), '{}');
        await symlink(path.join(workspace, 'outside.har'), path.join(root, 'linked.har'));
        await symlink(workspace, path.join(root, 'linked-directory'));
        assert.deepEqual(await findOrphanedHarFiles(root), [
            path.join(root, 'nested/recording.har'),
            path.join(root, 'orphan.har'),
        ]);
        const preview = spawnSync(process.execPath, [script], { cwd: workspace, encoding: 'utf8' });
        assert.equal(preview.status, 0, preview.stderr);
        assert.match(preview.stdout, /2 orphaned HAR files found/);
        await access(path.join(root, 'orphan.har'));
        const deletion = spawnSync(process.execPath, [script, '--delete'], { cwd: workspace, encoding: 'utf8' });
        assert.equal(deletion.status, 0, deletion.stderr);
        assert.match(deletion.stdout, /2 orphaned HAR files removed/);
        await assert.rejects(access(path.join(root, 'orphan.har')), { code: 'ENOENT' });
        await access(path.join(root, 'active.har'));
        await access(path.join(workspace, 'outside.har'));
        await symlink(tmpdir(), path.join(workspace, 'outside-link'));
        const unsafe = spawnSync(process.execPath, [script, '--root', 'outside-link', '--delete'], {
            cwd: workspace,
            encoding: 'utf8',
        });
        assert.equal(unsafe.status, 1);
        assert.match(unsafe.stderr, /leaves the UI workspace/);
    } finally {
        await rm(workspace, { recursive: true, force: true });
    }
});

for (const args of [['--unknown'], ['--root'], ['--root', '--delete'], ['--root', '..']]) {
    test(`invalid prune arguments fail: ${args.join(' ')}`, () => {
        const result = spawnSync(process.execPath, [script, ...args], { encoding: 'utf8' });
        assert.equal(result.status, 1);
        assert.match(result.stderr, /Usage: yarn har:prune|--root requires a directory|HAR root must be/);
    });
}
