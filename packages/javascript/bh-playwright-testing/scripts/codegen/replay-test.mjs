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
import { createRequire } from 'node:module';
import path from 'node:path';
import { makeMockedTest } from '../mocked-codegen/index.mjs';
import { recordedSource, sampleCapture, writeAuthState, writeTestConfig } from './test-data.mjs';
import { startTestServer } from './test-server.mjs';

export async function createReplayFixture() {
    const require = createRequire(import.meta.url);
    const cli = require.resolve('@playwright/test/cli');
    const workspace = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..', '..');
    await mkdir(path.join(workspace, 'playwright'), { recursive: true });
    const directory = await mkdtemp(path.join(workspace, 'playwright', 'codegen-test-'));
    const server = await startTestServer();
    const capture = sampleCapture();
    await writeFile(path.join(directory, 'example.har'), JSON.stringify(capture));
    const generated = makeMockedTest(recordedSource, 'example', 'http://old-host:3000', '// Generated');
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
