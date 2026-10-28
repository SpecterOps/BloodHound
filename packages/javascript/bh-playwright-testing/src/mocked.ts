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
import { readFile, writeFile } from 'node:fs/promises';
import { test as base, expect } from '@playwright/test';

/** Self-contained API replay for generated tests, with an isolated fake session. */
export const test = base.extend<{ replayHar: URL | undefined }>({
    replayHar: [undefined, { option: true }],
    serviceWorkers: 'block',
    storageState: { cookies: [], origins: [] },
    context: async ({ context, baseURL, replayHar }, use, testInfo) => {
        if (!replayHar) throw new Error('Set replayHar to the recorded HAR URL with test.use().');
        if (!baseURL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
        // Rebase recorded API URLs so the fixture also works on another developer's port/host.
        const har = JSON.parse(await readFile(replayHar, 'utf8'));
        for (const entry of har.log.entries) {
            const url = new URL(entry.request.url);
            entry.request.url = new URL(url.pathname + url.search, baseURL).href;
        }
        const replay = testInfo.outputPath('recording.har');
        await writeFile(replay, JSON.stringify(har));
        await context.routeFromHAR(replay, { url: /\/api(?:\/|$)/, notFound: 'abort' });
        // A fake local session keeps replay independent of a live login or expiring credentials.
        await context.addInitScript(() => {
            if (!localStorage.getItem('persistedState')) {
                localStorage.setItem(
                    'persistedState',
                    JSON.stringify({ auth: { sessionToken: 'playwright-mock-session' } })
                );
            }
        });
        await use(context);
    },
});

export { expect };
