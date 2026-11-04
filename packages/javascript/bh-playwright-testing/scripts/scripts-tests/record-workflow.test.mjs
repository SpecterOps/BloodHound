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
import { spawn } from 'node:child_process';
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile } from 'node:fs/promises';
import { createServer } from 'node:http';
import { createRequire } from 'node:module';
import path from 'node:path';
import test from 'node:test';
import { makeMockedSpec } from '../build-and-write-specs/build-spec.mjs';
import { buildAndWriteSpec, writeSpec } from '../build-and-write-specs/write-spec.mjs';
import { sanitizeHar } from '../transform-and-write-hars/sanitize-har.mjs';
import {
    mergeHarCaptures,
    writeNewHar,
    writeRecordedHar,
    writeUpdatedHar,
} from '../transform-and-write-hars/write-har.mjs';
import { captureForUpdate } from '../capture-actions-and-hars/capture-hars.mjs';
import {
    checkHasSameOrigin,
    makeRecordingFilePaths,
    makeRecordingSession,
} from '../capture-actions-and-hars/make-recording-session.mjs';
import { listRecordedSpecs, selectRecordedSpecs } from '../update-hars/select-specs.mjs';
import {
    makeDevServerArguments,
    parseRecordingSessionArguments,
    makeRecorderArguments,
} from '../capture-actions-and-hars/prepare-arguments.mjs';
import { checkForDuplicateSpecName } from '../capture-actions-and-hars/manage-files.mjs';

const secretCapture = {
    log: {
        entries: [
            {
                request: {
                    headers: [{ name: 'Authorization', value: 'Bearer secret' }],
                    cookies: [{ value: 'secret' }],
                },
                response: {
                    headers: [{ name: 'Set-Cookie', value: 'secret' }],
                    cookies: [],
                    content: { text: 'secret' },
                },
            },
        ],
    },
};

const secretState = {
    origins: [
        { localStorage: [{ name: 'persistedState', value: JSON.stringify({ auth: { sessionToken: 'secret' } }) }] },
    ],
};

function entry(method, title) {
    return {
        request: {
            method,
            url: 'http://old-host:3000/api/v2/items',
            headers: [],
            cookies: [],
            ...(method === 'POST' ? { postData: { mimeType: 'application/json', text: '{"query":"items"}' } } : {}),
        },
        response: {
            status: 200,
            headers: [{ name: 'content-type', value: 'application/json' }],
            content: { mimeType: 'application/json', text: JSON.stringify({ title }) },
        },
    };
}

function sampleCapture() {
    const authenticatedEntry = entry('GET', 'Mocked GET');
    authenticatedEntry.request.headers.push({ name: 'Authorization', value: 'Bearer recordingSession-secret' });
    return sanitizeHar(
        { log: { entries: [entry('GET', 'Anonymous GET'), authenticatedEntry, entry('POST', 'Mocked POST')] } },
        { origins: [] }
    );
}

const recordingSessionSource = `import { test, expect } from '@playwright/test';
test.use({ storageState: 'private-auth.json' });
test('example', async ({ page }) => {
  await page.routeFromHAR('/private/temporary/capture.har', { url: '**/api/**' });
  await page.goto('http://old-host:3000/ui/');
  await expect(page.getByRole('heading')).toHaveText('Mocked GET');
  await page.getByRole('button', {name:'Load'}).click();
  await expect(page.getByRole('heading')).toHaveText('Mocked POST');
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('persistedState')).auth.sessionToken)).toBe('playwright-mock-session');
});`;

async function writeAuthState(directory, baseURL) {
    await writeFile(
        path.join(directory, 'auth.json'),
        JSON.stringify({
            cookies: [],
            origins: [
                {
                    origin: baseURL,
                    localStorage: [
                        {
                            name: 'persistedState',
                            value: JSON.stringify({ auth: { sessionToken: 'playwright-mock-session' } }),
                        },
                    ],
                },
            ],
        })
    );
}

async function writeTestConfig(directory, baseURL, testModule) {
    await writeFile(
        path.join(directory, 'tsconfig.json'),
        JSON.stringify({
            compilerOptions: { baseUrl: directory, paths: { '@playwright/test': [testModule] } },
        })
    );
    await writeFile(
        path.join(directory, 'playwright.config.mjs'),
        `export default { testDir: '.', outputDir: './results', timeout: 15000, tsconfig: './tsconfig.json', use: {baseURL: ${JSON.stringify(baseURL)}, headless: true, storageState: {cookies: [], origins: []}}, reporter: 'line' };`
    );
}

function respond(request, response, state) {
    if (request.url.startsWith('/api/')) {
        state.liveApiRequests++;
        if (!state.recordApiResponses) return response.writeHead(500).end('Backend must not be used');
        response.setHeader('Content-Type', 'application/json');
        return response.end(JSON.stringify({ title: request.method === 'POST' ? 'Mocked POST' : 'Mocked GET' }));
    }
    response.setHeader('Content-Type', 'text/html');
    response.end(
        `<button onclick="fetch('/api/v2/items', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({query:'items'}) }).then(r=>r.json()).then(d=>document.querySelector('h1').textContent=d.title)">Load</button><h1>Loading</h1><script>fetch('/api/v2/items', {headers: {Authorization: 'Bearer ' + JSON.parse(localStorage.getItem('persistedState')).auth.sessionToken}}).then(r=>r.json()).then(d=>document.querySelector('h1').textContent=d.title)</script>`
    );
}

async function startTestServer() {
    const state = { liveApiRequests: 0, recordApiResponses: false };
    const server = createServer((request, response) => respond(request, response, state));
    await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
    return {
        baseURL: `http://127.0.0.1:${server.address().port}`,
        state,
        close: async () => {
            server.closeAllConnections();
            await new Promise((resolve) => server.close(resolve));
        },
    };
}

async function createReplayFixture() {
    const require = createRequire(import.meta.url);
    const cli = require.resolve('@playwright/test/cli');
    const workspace = path.resolve(path.dirname(new URL(import.meta.url).pathname), '../..');
    await mkdir(path.join(workspace, 'playwright'), { recursive: true });
    const directory = await mkdtemp(path.join(workspace, 'playwright', 'codegen-test-'));
    const server = await startTestServer();
    const capture = sampleCapture();
    await writeFile(path.join(directory, 'example.har'), JSON.stringify(capture));
    const generated = makeMockedSpec(recordingSessionSource, 'example', 'http://old-host:3000', '// Generated');
    assert.ok(!generated.includes('private-auth.json'));
    assert.ok(!generated.includes('/private/temporary/capture.har'));
    await writeFile(path.join(directory, 'example.spec.ts'), generated);
    await writeAuthState(directory, server.baseURL);
    const testModule = require.resolve('@playwright/test').replace(/index\.js$/, 'index.mjs');
    await writeTestConfig(directory, server.baseURL, testModule);
    return { cli, workspace, directory, server, capture };
}

function runBrowserTest(fixture, env) {
    return new Promise((resolve, reject) => {
        const child = spawn(
            process.execPath,
            [fixture.cli, 'test', '-c', path.join(fixture.directory, 'playwright.config.mjs')],
            { cwd: fixture.workspace, env }
        );
        let output = '';
        child.stdout.on('data', (data) => {
            output += data;
        });
        child.stderr.on('data', (data) => {
            output += data;
        });
        child.on('error', reject);
        child.on('exit', (code) => resolve({ code, output }));
    });
}

async function verifyReplay(fixture) {
    const result = await runBrowserTest(fixture, process.env);
    assert.equal(result.code, 0, result.output);
    assert.equal(fixture.server.state.liveApiRequests, 0);
}

async function verifyUpdate(fixture) {
    fixture.server.state.recordApiResponses = true;
    const captures = path.join(fixture.directory, 'captures');
    await mkdir(captures);
    const update = await runBrowserTest(fixture, {
        ...process.env,
        HAR_UPDATE_CAPTURE_DIRECTORY: captures,
        HAR_UPDATE_STORAGE_STATE: path.join(fixture.directory, 'auth.json'),
    });
    assert.equal(update.code, 0, `${update.output}\nLive API requests: ${fixture.server.state.liveApiRequests}`);
    assert.ok(fixture.server.state.liveApiRequests > 0);
    const captureFiles = await readdir(captures);
    assert.equal(captureFiles.length, 1);
    assert.ok(JSON.parse(await readFile(path.join(captures, captureFiles[0]), 'utf8')).log.entries.length > 0);
    assert.deepEqual(JSON.parse(await readFile(path.join(fixture.directory, 'example.har'), 'utf8')), fixture.capture);
}

async function closeReplayFixture(fixture) {
    await fixture.server.close();
    await rm(fixture.directory, { recursive: true, force: true });
}

test('codegen requires one valid spec name', () => {
    assert.deepEqual(parseRecordingSessionArguments(['feature', '--url', '/ui/explore']), {
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
        assert.throws(() => parseRecordingSessionArguments(args), /Usage:/);
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
        checkForDuplicateSpecName(files);
        await writeNewHar(files.har, har);
        await writeSpec(files.spec, 'edited spec');
        assert.throws(() => checkForDuplicateSpecName(files), /already exists/);
        assert.deepEqual(await selectRecordedSpecs(directory, ['feature']), [{ name: 'feature', ...files }]);
        assert.deepEqual(await listRecordedSpecs(directory), [{ name: 'feature', ...files }]);
        assert.equal(await readFile(files.spec, 'utf8'), 'edited spec');
        assert.deepEqual(JSON.parse(await readFile(files.har, 'utf8')), har);
    } finally {
        await rm(directory, { recursive: true, force: true });
    }
});

test('codegen records source alongside HAR data', () => {
    const args = makeRecorderArguments('auth.json', 'capture.ts', 'capture.har', 'http://localhost:3000/ui/');
    assert.ok(args.includes('--output'));
    assert.ok(args.includes('--save-har'));
});

test('recordingSession writers save a sanitized HAR and generated spec', async () => {
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-recordingSession-writers-'));
    const recordingSession = {
        name: 'feature',
        rawHar: path.join(directory, 'raw.har'),
        rawSource: path.join(directory, 'raw.ts'),
        storageState: path.join(directory, 'auth.json'),
        files: { har: path.join(directory, 'feature.har'), spec: path.join(directory, 'feature.spec.ts') },
    };
    try {
        await writeFile(recordingSession.rawHar, JSON.stringify(secretCapture));
        await writeFile(recordingSession.rawSource, recordingSessionSource);
        await writeFile(recordingSession.storageState, JSON.stringify(secretState));
        await writeRecordedHar(recordingSession);
        await buildAndWriteSpec(recordingSession, 'http://old-host:3000');
        assert.ok(!(await readFile(recordingSession.files.har, 'utf8')).includes('secret'));
        const spec = await readFile(recordingSession.files.spec, 'utf8');
        assert.ok(spec.startsWith('// Copyright 2026 Specter Ops, Inc.'));
        assert.ok(spec.includes('// SPDX-License-Identifier: Apache-2.0'));
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
    const originalCapture = structuredClone(secretCapture);
    const originalState = structuredClone(secretState);
    const clean = sanitizeHar(secretCapture, secretState);
    assert.ok(!JSON.stringify(clean).includes('secret'));
    assert.deepEqual(clean.log.entries[0].request.headers, [
        { name: 'Authorization', value: 'Bearer playwright-mock-session' },
    ]);
    assert.equal(secretCapture.log.entries[0].request.headers.length, 1);
    assert.equal(secretCapture.log.entries[0].request.headers[0].value, 'Bearer secret');
    assert.deepEqual(secretCapture, originalCapture);
    assert.deepEqual(secretState, originalState);
});

test('merges every HAR capture without changing inputs or sharing mutable entries', () => {
    const capturedHars = [sampleCapture(), sampleCapture()];
    const originalCaptures = structuredClone(capturedHars);
    const combinedHar = mergeHarCaptures(capturedHars);

    assert.deepEqual(
        combinedHar.log.entries,
        originalCaptures.flatMap((har) => har.log.entries)
    );
    assert.deepEqual(capturedHars, originalCaptures);

    combinedHar.log.entries[0].request.url = '/changed';

    assert.deepEqual(capturedHars, originalCaptures);
});

test('builds recording paths from supplied directories and rejects cross-origin URLs', () => {
    const recordingFiles = makeRecordingFilePaths('/workspace/ui', 'feature');
    const recordingSession = makeRecordingSession({
        name: 'feature',
        files: recordingFiles,
        homeUrl: 'http://localhost:3000/ui/',
        playwrightCli: '/tools/playwright',
        temporaryDirectory: '/temporary/recording',
        uiDirectory: '/workspace/ui',
    });

    assert.deepEqual(recordingFiles, {
        spec: '/workspace/ui/tests/e2e/feature.spec.ts',
        har: '/workspace/ui/tests/e2e/feature.har',
    });
    assert.equal(recordingSession.rawHar, '/temporary/recording/capture.har');
    assert.equal(recordingSession.storageState, '/workspace/ui/playwright/.auth/user-light.json');
    assert.equal(checkHasSameOrigin('http://localhost:3000/ui/explore', 'http://localhost:3000'), true);
    assert.equal(checkHasSameOrigin('http://other-host/ui/', 'http://localhost:3000'), false);
});

test('successful capture returns the directory containing the recorded HAR', async () => {
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-capture-result-'));
    const playwrightCli = path.join(directory, 'playwright-cli.cjs');
    let temporaryDirectory;

    try {
        await mkdir(path.join(process.cwd(), 'playwright'), { recursive: true });
        await writeFile(
            playwrightCli,
            `const fs = require('node:fs');
const path = require('node:path');
fs.writeFileSync(path.join(process.env.HAR_UPDATE_CAPTURE_DIRECTORY, 'capture.har'), '{"log":{"entries":[]}}');`
        );

        const captureResult = await captureForUpdate(playwrightCli, { spec: path.join(directory, 'feature.spec.ts') });
        temporaryDirectory = captureResult.temporaryDirectory;
        const captureFiles = await readdir(captureResult.capturesPath);

        assert.deepEqual(captureFiles, ['capture.har']);
    } finally {
        if (temporaryDirectory) {
            await rm(temporaryDirectory, { recursive: true, force: true });
        }

        await rm(directory, { recursive: true, force: true });
    }
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
