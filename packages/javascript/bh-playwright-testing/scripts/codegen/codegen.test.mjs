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
import { mkdtemp, readFile, rm } from 'node:fs/promises';
import path from 'node:path';
import test from 'node:test';
import { selectRecordedTests } from '../e2e-update-mocks/index.mjs';
import { sanitizeRecording } from '../mocked-codegen/index.mjs';
import {
    checkRecordingFiles,
    makeDevServerArguments,
    parseRecordingArguments,
    recorderArguments,
    saveRecording,
} from '../recording/index.mjs';
import { closeReplayFixture, createReplayFixture, verifyReplay, verifyUpdate } from './replay-test.mjs';
import { secretCapture, secretState } from './test-data.mjs';

test('codegen requires one valid test name', () => {
    assert.deepEqual(parseRecordingArguments(['feature', '--url', '/ui/explore']), {
        name: 'feature',
        requestedUrl: '/ui/explore',
    });
    for (const args of [
        [],
        ['../feature'],
        ['feature.spec.ts'],
        ['feature', '--bogus'],
        ['feature', '--url'],
        ['feature', '--update'],
    ]) {
        assert.throws(() => parseRecordingArguments(args), /Usage:/);
    }
});

test('codegen creates a new spec and HAR; updater selects existing tests', async () => {
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-codegen-update-'));
    const files = { spec: path.join(directory, 'feature.spec.ts'), har: path.join(directory, 'feature.har') };
    const har = { log: { entries: [{ request: { url: '/api/items' } }] } };
    try {
        await assert.rejects(selectRecordedTests(directory, []), /No recorded tests found/);
        await saveRecording(files, har, 'edited spec');
        assert.throws(() => checkRecordingFiles(files), /already exists/);
        assert.deepEqual(await selectRecordedTests(directory, []), [{ name: 'feature', ...files }]);
        assert.equal(await readFile(files.spec, 'utf8'), 'edited spec');
        assert.deepEqual(JSON.parse(await readFile(files.har, 'utf8')), har);
    } finally {
        await rm(directory, { recursive: true, force: true });
    }
});

test('codegen records source alongside HAR data', () => {
    const args = recorderArguments('auth.json', 'capture.ts', 'capture.har', 'http://localhost:3000/ui/');
    assert.ok(args.includes('--output'));
    assert.ok(args.includes('--save-har'));
});

test('builds the default Vite server command from the configured URL', () => {
    assert.deepEqual(makeDevServerArguments('http://127.0.0.1:3000'), [
        'dev',
        '--mode',
        'playwright',
        '--host',
        '127.0.0.1',
        '--port',
        '3000',
    ]);
    assert.deepEqual(makeDevServerArguments('http://bloodhound.localhost'), [
        'dev',
        '--mode',
        'playwright',
        '--host',
        'bloodhound.localhost',
        '--port',
        '3000',
    ]);
});

test('removes credentials without mutating the original capture', () => {
    const clean = sanitizeRecording(secretCapture, secretState);
    assert.ok(!JSON.stringify(clean).includes('secret'));
    assert.deepEqual(clean.log.entries[0].request.headers, [
        { name: 'Authorization', value: 'Bearer playwright-mock-session' },
    ]);
    assert.equal(secretCapture.log.entries[0].request.headers.length, 1);
    assert.equal(secretCapture.log.entries[0].request.headers[0].value, 'Bearer secret');
});

test('generated spec replays GET and POST APIs on a different host without login or backend', async () => {
    const fixture = await createReplayFixture();
    try {
        await verifyReplay(fixture);
        await verifyUpdate(fixture);
    } finally {
        await closeReplayFixture(fixture);
    }
});
