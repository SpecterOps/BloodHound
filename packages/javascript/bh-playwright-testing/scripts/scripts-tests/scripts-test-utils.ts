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
import { makeMockedSpec } from '../build-and-write-specs/build-spec.mjs';
import { sanitizeRecording } from '../transform-and-write-hars/sanitize-recording.mjs';

export const secretCapture = {
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

export const secretState = {
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

export function sampleCapture() {
    const authenticatedEntry = entry('GET', 'Mocked GET');
    authenticatedEntry.request.headers.push({ name: 'Authorization', value: 'Bearer recording-secret' });
    return sanitizeRecording(
        { log: { entries: [entry('GET', 'Anonymous GET'), authenticatedEntry, entry('POST', 'Mocked POST')] } },
        { origins: [] }
    );
}

export const recordedSource = `import { test, expect } from '@playwright/test';
test.use({ storageState: 'private-auth.json' });
test('example', async ({ page }) => {
  await page.routeFromHAR('/private/temporary/recording.har', { url: '**/api/**' });
  await page.goto('http://old-host:3000/ui/');
  await expect(page.getByRole('heading')).toHaveText('Mocked GET');
  await page.getByRole('button', {name:'Load'}).click();
  await expect(page.getByRole('heading')).toHaveText('Mocked POST');
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('persistedState')).auth.sessionToken)).toBe('playwright-mock-session');
});`;

export async function writeAuthState(directory, baseURL) {
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

export async function writeTestConfig(directory, baseURL, testModule) {
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

export async function createReplayFixture() {
    const require = createRequire(import.meta.url);
    const cli = require.resolve('@playwright/test/cli');
    const workspace = path.resolve(path.dirname(new URL(import.meta.url).pathname), '../..');
    await mkdir(path.join(workspace, 'playwright'), { recursive: true });
    const directory = await mkdtemp(path.join(workspace, 'playwright', 'codegen-test-'));
    const server = await startTestServer();
    const capture = sampleCapture();
    await writeFile(path.join(directory, 'example.har'), JSON.stringify(capture));
    const generated = makeMockedSpec(recordedSource, 'example', 'http://old-host:3000', '// Generated');
    assert.ok(!generated.includes('private-auth.json'));
    assert.ok(!generated.includes('/private/temporary/recording.har'));
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

export async function verifyReplay(fixture) {
    const result = await runBrowserTest(fixture, process.env);
    assert.equal(result.code, 0, result.output);
    assert.equal(fixture.server.state.liveApiRequests, 0);
}

export async function verifyUpdate(fixture) {
    fixture.server.state.recordApiResponses = true;
    const captures = path.join(fixture.directory, 'captures');
    await mkdir(captures);
    const update = await runBrowserTest(fixture, {
        ...process.env,
        HAR_UPDATE_CAPTURE: captures,
        HAR_UPDATE_STORAGE_STATE: path.join(fixture.directory, 'auth.json'),
    });
    assert.equal(update.code, 0, `${update.output}\nLive API requests: ${fixture.server.state.liveApiRequests}`);
    assert.ok(fixture.server.state.liveApiRequests > 0);
    const captureFiles = await readdir(captures);
    assert.equal(captureFiles.length, 1);
    assert.ok(JSON.parse(await readFile(path.join(captures, captureFiles[0]), 'utf8')).log.entries.length > 0);
    assert.deepEqual(JSON.parse(await readFile(path.join(fixture.directory, 'example.har'), 'utf8')), fixture.capture);
}

export async function closeReplayFixture(fixture) {
    await fixture.server.close();
    await rm(fixture.directory, { recursive: true, force: true });
}
