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

import { expect, test } from '@playwright/test';
import { mkdir, readFile, readdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { harArtifactPaths, publishHar, readHar, readJsonArtifact, summarizeHar, writeJsonArtifact } from '../src/har';

const capturedEntries = [
    {
        request: { method: 'POST', url: 'https://example.test/items', bodySize: 12 },
        response: { status: 201, content: { mimeType: 'application/json', size: 24 } },
    },
    {
        request: { method: 'GET', url: 'https://example.test/empty', bodySize: 0 },
        response: { status: 204, content: { mimeType: '', size: 0 } },
    },
    {
        request: { method: 'GET', url: 'https://example.test/unknown' },
        response: { status: 200 },
    },
];

function recording(entries: unknown = capturedEntries) {
    return { log: { version: '1.2', entries } };
}

const expectedRequests = [
    { method: 'POST', url: 'https://example.test/items', bodySize: 12 },
    { method: 'GET', url: 'https://example.test/empty', bodySize: 0 },
    { method: 'GET', url: 'https://example.test/unknown', bodySize: null },
];
const expectedResponses = [
    { method: 'POST', url: 'https://example.test/items', status: 201, mimeType: 'application/json', bodySize: 24 },
    { method: 'GET', url: 'https://example.test/empty', status: 204, mimeType: '', bodySize: 0 },
    { method: 'GET', url: 'https://example.test/unknown', status: 200, mimeType: '', bodySize: null },
];

test('canonical paths preserve their existing identity and ignore execution details', async () => {
    const testInfo = test.info();
    const identity = {
        ...testInfo,
        config: { ...testInfo.config, rootDir: '/suite' },
        file: '/suite/nested/login.spec.ts',
        titlePath: ['nested/login.spec.ts', 'Account / settings', 'sign in?'],
        testId: 'stable-test-id',
        project: { ...testInfo.project, name: 'chromium' },
        repeatEachIndex: 0,
    };
    const paths = harArtifactPaths(identity, 'fixtures');
    expect(paths).toEqual({
        directory: '/suite/fixtures/login-spec-Account-settings-sign-in--chromium-5552b0302c45',
        recording: '/suite/fixtures/login-spec-Account-settings-sign-in--chromium-5552b0302c45/recording.har',
        requests: '/suite/fixtures/login-spec-Account-settings-sign-in--chromium-5552b0302c45/requests.json',
        responses: '/suite/fixtures/login-spec-Account-settings-sign-in--chromium-5552b0302c45/responses.json',
    });
    expect(harArtifactPaths({ ...identity, retry: 2, workerIndex: 7, parallelIndex: 3 }, 'fixtures')).toEqual(paths);
    expect(harArtifactPaths(identity, '/external/fixtures').directory).toBe(
        path.join('/external/fixtures', path.basename(paths.directory))
    );
    for (const variation of [
        { file: '/suite/other/login.spec.ts' },
        { titlePath: [...identity.titlePath, 'parameter b'] },
        { testId: 'another-test-id' },
        { project: { ...identity.project, name: 'firefox' } },
        { repeatEachIndex: 1 },
    ]) {
        expect(harArtifactPaths({ ...identity, ...variation }, 'fixtures').directory).not.toBe(paths.directory);
    }
});

test('summaries preserve order, zero sizes, and defaults for absent content', async () => {
    const testInfo = test.info();
    const paths = harArtifactPaths(testInfo, testInfo.outputPath('fixtures'));
    await writeJsonArtifact(paths.recording, recording());
    await summarizeHar(paths.recording, paths);
    expect(await readJsonArtifact(paths.requests)).toEqual(expectedRequests);
    expect(await readJsonArtifact(paths.responses)).toEqual(expectedResponses);
});

test('valid empty HARs produce empty summaries', async () => {
    const testInfo = test.info();
    const paths = harArtifactPaths(testInfo, testInfo.outputPath('fixtures'));
    await writeJsonArtifact(paths.recording, recording([]));
    expect(await readHar(paths.recording)).toEqual([]);
    await summarizeHar(paths.recording, paths);
    expect(await readJsonArtifact(paths.requests)).toEqual([]);
    expect(await readJsonArtifact(paths.responses)).toEqual([]);
});

for (const [description, value] of [
    ['null document', null],
    ['missing version', { log: { entries: [] } }],
    ['non-array entries', recording({})],
    ['null entry', recording([null])],
    ['missing request', recording([{ response: { status: 200 } }])],
    ['non-string method', recording([{ request: { method: 1, url: '/' }, response: { status: 200 } }])],
    ['non-string URL', recording([{ request: { method: 'GET', url: 1 }, response: { status: 200 } }])],
    ['non-numeric status', recording([{ request: { method: 'GET', url: '/' }, response: { status: '200' } }])],
] as const) {
    test(`rejects ${description} with the HAR path`, async () => {
        const testInfo = test.info();
        const file = testInfo.outputPath('invalid.har');
        await writeJsonArtifact(file, value);
        await expect(readHar(file)).rejects.toThrow(`Malformed HAR at ${file}`);
    });
}

test('invalid JSON reports the file and preserves the parsing error', async () => {
    const testInfo = test.info();
    const file = testInfo.outputPath('invalid.har');
    await writeFile(file, '{');
    await expect(readHar(file)).rejects.toMatchObject({
        message: expect.stringContaining(`Cannot read HAR at ${file}`),
        cause: expect.any(SyntaxError),
    });
});

test('publishing invalid capture preserves all canonical artifacts', async () => {
    const testInfo = test.info();
    const paths = harArtifactPaths(testInfo, testInfo.outputPath('fixtures'));
    const temporary = testInfo.outputPath('invalid.har');
    await writeJsonArtifact(paths.recording, recording());
    await summarizeHar(paths.recording, paths);
    const files = [paths.recording, paths.requests, paths.responses];
    const originalContents = await Promise.all(files.map((file) => readFile(file, 'utf8')));
    await writeJsonArtifact(temporary, recording([null]));
    await expect(publishHar(temporary, paths)).rejects.toThrow('Malformed HAR');
    expect(await Promise.all(files.map((file) => readFile(file, 'utf8')))).toEqual(originalContents);
});

test('publishing replaces the recording and its summaries', async () => {
    const testInfo = test.info();
    const paths = harArtifactPaths(testInfo, testInfo.outputPath('fixtures'));
    const temporary = testInfo.outputPath('capture.har');
    await writeJsonArtifact(paths.recording, recording([]));
    await summarizeHar(paths.recording, paths);
    await writeJsonArtifact(temporary, recording());
    await publishHar(temporary, paths);
    expect(await readHar(paths.recording)).toEqual(capturedEntries);
    expect(await readJsonArtifact(paths.requests)).toEqual(expectedRequests);
    expect(await readJsonArtifact(paths.responses)).toEqual(expectedResponses);
    await expect(readFile(temporary)).rejects.toMatchObject({ code: 'ENOENT' });
});

test('concurrent JSON writes leave a complete formatted artifact and no temporary files', async () => {
    const testInfo = test.info();
    const directory = testInfo.outputPath('nested', 'artifacts');
    const file = path.join(directory, 'metadata.json');
    await mkdir(directory, { recursive: true });
    const values = Array.from({ length: 12 }, (_, index) => ({ index, text: 'payload'.repeat(100) }));
    await Promise.all(values.map((value) => writeJsonArtifact(file, value)));
    const result = await readJsonArtifact(file);
    expect(values).toContainEqual(result);
    expect(await readFile(file, 'utf8')).toBe(`${JSON.stringify(result, null, 2)}\n`);
    expect(await readdir(directory)).toEqual(['metadata.json']);
});
