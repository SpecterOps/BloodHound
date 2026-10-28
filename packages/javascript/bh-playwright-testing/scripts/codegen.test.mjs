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
import { selectRecordedTests } from './e2e-update-mocks.mjs';
import { makeMockedTest, sanitizeRecording } from './mocked-codegen.mjs';
import {
    checkRecordingFiles,
    makeDevServerArguments,
    parseRecordingArguments,
    recorderArguments,
    saveRecording,
} from './recording.mjs';

test('codegen creates a new spec and HAR; updater selects existing tests', async () => {
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
    const directory = await mkdtemp(path.join(process.env.TMPDIR || '/tmp', 'bh-codegen-update-'));
    const files = { spec: path.join(directory, 'feature.spec.ts'), har: path.join(directory, 'feature.har') };
    const newHar = { log: { entries: [{ request: { url: '/api/items' } }] } };
    try {
        await assert.rejects(selectRecordedTests(directory, []), /No recorded tests found/);
        await saveRecording(files, newHar, 'edited spec');
        assert.throws(() => checkRecordingFiles(files), /already exists/);
        assert.deepEqual(await selectRecordedTests(directory, []), [{ name: 'feature', ...files }]);
        assert.equal(await readFile(files.spec, 'utf8'), 'edited spec');
        assert.deepEqual(JSON.parse(await readFile(files.har, 'utf8')), newHar);
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
    const capture = {
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
    const state = {
        origins: [
            { localStorage: [{ name: 'persistedState', value: JSON.stringify({ auth: { sessionToken: 'secret' } }) }] },
        ],
    };
    const clean = sanitizeRecording(capture, state);
    assert.ok(!JSON.stringify(clean).includes('secret'));
    assert.deepEqual(clean.log.entries[0].request.headers, [
        { name: 'Authorization', value: 'Bearer playwright-mock-session' },
    ]);
    assert.equal(capture.log.entries[0].request.headers.length, 1);
    assert.equal(capture.log.entries[0].request.headers[0].value, 'Bearer secret');
});

test('generated spec replays GET and POST APIs on a different host without login or backend', async () => {
    const require = createRequire(import.meta.url);
    const cli = require.resolve('@playwright/test/cli');
    const workspace = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
    await mkdir(path.join(workspace, 'playwright'), { recursive: true });
    const directory = await mkdtemp(path.join(workspace, 'playwright', 'codegen-test-'));
    let liveApiRequests = 0;
    let recordApiResponses = false;
    const server = createServer((request, response) => {
        if (request.url.startsWith('/api/')) {
            liveApiRequests++;
            if (!recordApiResponses) {
                response.writeHead(500).end('Backend must not be used');
            } else {
                response.setHeader('Content-Type', 'application/json');
                response.end(JSON.stringify({ title: request.method === 'POST' ? 'Mocked POST' : 'Mocked GET' }));
            }
            return;
        }
        response.setHeader('Content-Type', 'text/html');
        response.end(
            `<button onclick="fetch('/api/v2/items', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({query:'items'}) }).then(r=>r.json()).then(d=>document.querySelector('h1').textContent=d.title)">Load</button><h1>Loading</h1><script>fetch('/api/v2/items', {headers: {Authorization: 'Bearer ' + JSON.parse(localStorage.getItem('persistedState')).auth.sessionToken}}).then(r=>r.json()).then(d=>document.querySelector('h1').textContent=d.title)</script>`
        );
    });
    await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
    const baseURL = `http://127.0.0.1:${server.address().port}`;
    const entry = (method, title) => ({
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
    });
    try {
        const authenticatedEntry = entry('GET', 'Mocked GET');
        authenticatedEntry.request.headers.push({ name: 'Authorization', value: 'Bearer recording-secret' });
        const capture = sanitizeRecording(
            { log: { entries: [entry('GET', 'Anonymous GET'), authenticatedEntry, entry('POST', 'Mocked POST')] } },
            { origins: [] }
        );
        await writeFile(path.join(directory, 'example.har'), JSON.stringify(capture));
        const source = `import { test, expect } from '@playwright/test';
test.use({ storageState: 'private-auth.json' });
test('test', async ({ page }) => {
  await page.routeFromHAR('/private/temporary/recording.har', { url: '**/api/**' });
  await page.goto('http://old-host:3000/ui/');
  await expect(page.getByRole('heading')).toHaveText('Mocked GET');
  await page.getByRole('button', {name:'Load'}).click();
  await expect(page.getByRole('heading')).toHaveText('Mocked POST');
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('persistedState')).auth.sessionToken)).toBe('playwright-mock-session');
});`;
        const generated = makeMockedTest(source, 'example', 'http://old-host:3000', '// Generated');
        assert.ok(!generated.includes('private-auth.json'));
        assert.ok(!generated.includes('/private/temporary/recording.har'));
        await writeFile(path.join(directory, 'example.spec.ts'), generated);
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
        // Resolve through the consuming workspace just as the BHE config does.
        const testModule = require.resolve('@playwright/test').replace(/index\.js$/, 'index.mjs');
        await writeFile(
            path.join(directory, 'tsconfig.json'),
            JSON.stringify({
                compilerOptions: {
                    baseUrl: directory,
                    paths: { '@playwright/test': [testModule] },
                },
            })
        );
        await writeFile(
            path.join(directory, 'playwright.config.mjs'),
            `export default { testDir: '.', outputDir: './results', timeout: 15000, tsconfig: './tsconfig.json', use: {baseURL: ${JSON.stringify(baseURL)}, headless: true, storageState: {cookies: [], origins: []}}, reporter: 'line' };`
        );
        const runBrowserTest = (env) =>
            new Promise((resolve, reject) => {
                const child = spawn(
                    process.execPath,
                    [cli, 'test', '-c', path.join(directory, 'playwright.config.mjs')],
                    {
                        cwd: workspace,
                        env,
                    }
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
        const result = await runBrowserTest(process.env);
        assert.equal(result.code, 0, result.output);
        assert.equal(liveApiRequests, 0);
        recordApiResponses = true;
        const captures = path.join(directory, 'captures');
        await mkdir(captures);
        const update = await runBrowserTest({
            ...process.env,
            HAR_UPDATE_CAPTURE: captures,
            HAR_UPDATE_STORAGE_STATE: path.join(directory, 'auth.json'),
        });
        assert.equal(update.code, 0, `${update.output}\nLive API requests: ${liveApiRequests}`);
        assert.ok(liveApiRequests > 0);
        const captureFiles = await readdir(captures);
        assert.equal(captureFiles.length, 1);
        assert.ok(JSON.parse(await readFile(path.join(captures, captureFiles[0]), 'utf8')).log.entries.length > 0);
        assert.deepEqual(JSON.parse(await readFile(path.join(directory, 'example.har'), 'utf8')), capture);
    } finally {
        server.closeAllConnections();
        await new Promise((resolve) => server.close(resolve));
        await rm(directory, { recursive: true, force: true });
    }
});
