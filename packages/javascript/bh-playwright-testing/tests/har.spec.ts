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

import { createServer, type Server } from 'node:http';
import { access, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { expect, test } from '../src/axe';
import {
    harArtifactPaths,
    readHar,
    readJsonArtifact,
    type HarRequestSummary,
    type HarResponseSummary,
} from '../src/har';

const mode = process.env.HAR_TEST_MODE as 'record' | 'update' | 'mock' | undefined;
const url = 'http://127.0.0.1:18887/value';

test.describe('HAR lifecycle', () => {
    test.describe.configure({ mode: 'serial' });
    test.use({ harMode: mode ?? 'off', harRootDir: '../playwright/har-test-fixtures' });
    let server: Server | undefined;
    let firstRecording: string;
    let failedRecording: string;

    test.beforeAll(async () => {
        if (mode === 'mock') return;
        server = createServer((_request, response) => {
            response.setHeader('Content-Type', 'text/plain');
            response.end(mode === 'update' ? 'refreshed' : 'original');
        });
        await new Promise<void>((resolve) => server!.listen(18887, '127.0.0.1', resolve));
    });
    test.afterAll(async () => {
        if (server?.listening)
            await new Promise<void>((resolve, reject) => server!.close((error) => (error ? reject(error) : resolve())));
    });

    test('captures, updates, and replays the same response', async ({ page, harArtifacts }) => {
        firstRecording = harArtifacts.recording;
        await page.goto(url);
        await expect(page.locator('body')).toHaveText(mode === 'record' ? 'original' : 'refreshed');
    });

    test('finalizes recording after an assertion failure', async ({ page, harArtifacts }) => {
        failedRecording = harArtifacts.recording;
        test.fail(true, 'exercise fixture teardown after a test failure');
        await page.goto(url);
        expect('intentional failure').toBe('success');
    });

    test('writes summaries and rejects unmatched requests in strict mode', async ({ page, harArtifacts }) => {
        await access(firstRecording);
        const identity = await readJsonArtifact<{ testId: string }>(
            path.join(path.dirname(firstRecording), '.har-fixture.json')
        );
        expect(identity.testId).toBeTruthy();
        await access(failedRecording);
        if (mode === 'update') {
            expect(JSON.stringify(await readHar(failedRecording))).toContain('original');
        }
        const entries = await readHar(firstRecording);
        expect(entries.some((entry) => entry.request.url === url && entry.response.status === 200)).toBeTruthy();
        const firstPaths = {
            ...harArtifacts,
            recording: firstRecording,
            requests: firstRecording.replace('recording.har', 'requests.json'),
            responses: firstRecording.replace('recording.har', 'responses.json'),
        };
        const requests = await readJsonArtifact<HarRequestSummary[]>(firstPaths.requests);
        const responses = await readJsonArtifact<HarResponseSummary[]>(firstPaths.responses);
        expect(requests.some((request) => request.url === url && request.method === 'GET')).toBeTruthy();
        expect(responses.some((response) => response.url === url && response.status === 200)).toBeTruthy();

        await page.goto(url);
        if (mode === 'mock') {
            test.fail(true, 'strict HAR mocks report unmatched requests during fixture teardown');
            await page.evaluate(() => fetch('/unexpected').catch(() => undefined));
        }
    });
});

test.describe('HAR helpers without network fixtures', () => {
    test.use({ harMode: 'off' });

    test('paths are stable across retries and unique for parameterized identities', async ({
        harArtifacts: _harArtifacts,
    }, testInfo) => {
        void _harArtifacts;
        const first = harArtifactPaths(testInfo, '../playwright/har-test-fixtures');
        expect(
            harArtifactPaths({ ...testInfo, retry: testInfo.retry + 1 }, '../playwright/har-test-fixtures').recording
        ).toBe(first.recording);
        expect(
            harArtifactPaths(
                { ...testInfo, testId: `${testInfo.testId}-parameter-b` },
                '../playwright/har-test-fixtures'
            ).recording
        ).not.toBe(first.recording);
    });

    test('missing and malformed recordings report their path', async ({ harArtifacts: _harArtifacts }, testInfo) => {
        void _harArtifacts;
        const missing = testInfo.outputPath('missing.har');
        await expect(readHar(missing)).rejects.toThrow(`Cannot read HAR at ${missing}`);
        const malformed = testInfo.outputPath('malformed.har');
        await writeFile(malformed, '{"log":{"entries":[]}}');
        await expect(readHar(malformed)).rejects.toThrow(`Malformed HAR at ${malformed}`);
    });

    test('off leaves normal page behavior and creates no HAR directory', async ({ page, harArtifacts }) => {
        await page.goto('data:text/html,<main>unmocked</main>');
        await expect(page.locator('main')).toHaveText('unmocked');
        await expect(access(harArtifacts.directory)).rejects.toThrow();
    });
});
