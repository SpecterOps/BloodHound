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
import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import test from 'node:test';
import { buildAndWriteSpec, writeSpec } from '../build-and-write-specs/write-spec.mjs';
import {
    closeReplayFixture,
    createReplayFixture,
    browserActionsSource,
    secretCapture,
    secretState,
    verifyReplay,
    verifyUpdate,
} from './scripts-test-utils.ts';
import { sanitizeHar } from '../transform-and-write-hars/sanitize-har.mjs';
import { writeNewHar, writeRecordedHar, writeUpdatedHar } from '../transform-and-write-hars/write-har.mjs';
import { listRecordedSpecs, selectRecordedSpecs } from '../update-hars/select-specs.mjs';
import { parseBrowserActionsArguments, recorderArguments } from '../record/prepare-arguments.mjs';
import { checkBrowserActionsFiles } from '../record/manage-files.mjs';
import { makeDevServerArguments } from '../record/ensure-server.mjs';

test('codegen requires one valid spec name', () => {
    assert.deepEqual(parseBrowserActionsArguments(['feature', '--url', '/ui/explore']), {
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
        assert.throws(() => parseBrowserActionsArguments(args), /Usage:/);
    }
});

test('codegen creates a new spec and HAR; updater selects existing specs', async () => {
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-codegen-update-'));
    const files = { spec: path.join(directory, 'feature.spec.ts'), har: path.join(directory, 'feature.har') };
    const har = { log: { entries: [{ request: { url: '/api/items' } }] } };
    try {
        await assert.rejects(selectRecordedSpecs(directory, []), /Usage:.*<name>/);
        await assert.rejects(selectRecordedSpecs(directory, ['feature', 'other']), /Usage:.*<name>/);
        await assert.rejects(listRecordedSpecs(directory), /No recorded specs found/);
        checkBrowserActionsFiles(files);
        await writeNewHar(files.har, har);
        await writeSpec(files.spec, 'edited spec');
        assert.throws(() => checkBrowserActionsFiles(files), /already exists/);
        assert.deepEqual(await selectRecordedSpecs(directory, ['feature']), [{ name: 'feature', ...files }]);
        assert.deepEqual(await listRecordedSpecs(directory), [{ name: 'feature', ...files }]);
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

test('browserActions writers save a sanitized HAR and generated spec', async () => {
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-browserActions-writers-'));
    const browserActions = {
        name: 'feature',
        rawHar: path.join(directory, 'raw.har'),
        rawSource: path.join(directory, 'raw.ts'),
        storageState: path.join(directory, 'auth.json'),
        files: { har: path.join(directory, 'feature.har'), spec: path.join(directory, 'feature.spec.ts') },
    };
    try {
        await writeFile(browserActions.rawHar, JSON.stringify(secretCapture));
        await writeFile(browserActions.rawSource, browserActionsSource);
        await writeFile(browserActions.storageState, JSON.stringify(secretState));
        await writeRecordedHar(browserActions);
        await buildAndWriteSpec(browserActions, 'http://old-host:3000');
        assert.ok(!(await readFile(browserActions.files.har, 'utf8')).includes('secret'));
        const spec = await readFile(browserActions.files.spec, 'utf8');
        assert.ok(spec.includes('feature.har'));
        assert.ok(spec.includes('replayHar'));
        assert.ok(!spec.includes('private-auth.json'));
    } finally {
        await rm(directory, { recursive: true, force: true });
    }
});

test('HAR updater replaces captured data after sanitizing it', async () => {
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-update-writer-'));
    const captures = path.join(directory, 'captures');
    const selected = { name: 'feature', har: path.join(directory, 'feature.har') };
    try {
        await mkdir(captures);
        await writeFile(selected.har, 'old HAR');
        await writeFile(path.join(captures, 'capture.har'), JSON.stringify(secretCapture));
        await writeUpdatedHar(captures, secretState, selected);
        const updated = await readFile(selected.har, 'utf8');
        assert.ok(!updated.includes('secret'));
        assert.equal(JSON.parse(updated).log.entries.length, 1);
    } finally {
        await rm(directory, { recursive: true, force: true });
    }
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
    const clean = sanitizeHar(secretCapture, secretState);
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
