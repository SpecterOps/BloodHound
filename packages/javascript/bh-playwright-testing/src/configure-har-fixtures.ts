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
import path from 'node:path';
import type { BrowserContext, TestInfo } from '@playwright/test';

export async function configureHarReplay(context: BrowserContext, baseURL: string, replayHar: URL, testInfo: TestInfo) {
    // Rebase recorded API URLs so the fixture also works on another developer's port/host.
    const harSource = await readFile(replayHar, 'utf8');
    const har = JSON.parse(harSource);

    for (const entry of har.log.entries) {
        const url = new URL(entry.request.url);
        entry.request.url = new URL(url.pathname + url.search, baseURL).href;
    }

    const replay = testInfo.outputPath('recording.har');

    const replaySource = JSON.stringify(har);

    await writeFile(replay, replaySource);

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
}

export async function configureHarCapture(context: BrowserContext, captureDirectory: string, testInfo: TestInfo) {
    const captureName = `${encodeURIComponent(testInfo.testId)}.har`;
    const capture = path.join(captureDirectory, captureName);

    await context.routeFromHAR(capture, {
        url: /\/api(?:\/|$)/,
        update: true,
        updateContent: 'embed',
        updateMode: 'full',
    });
}
