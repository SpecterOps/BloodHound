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

import { writeFile } from 'node:fs/promises';
import path from 'node:path';
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
test('test', async ({ page }) => {
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
