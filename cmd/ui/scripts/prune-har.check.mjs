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
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
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
