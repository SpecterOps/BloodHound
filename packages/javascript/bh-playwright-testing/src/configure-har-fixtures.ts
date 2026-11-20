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
import { makeRebasedHar } from './rebase-har-urls';

async function writeReplayHar(baseUrl: string, replayHar: URL, replayPath: string) {
    // Rebase recorded API URLs so the fixture also works on another developer's port/host.
    const harSource = await readFile(replayHar, 'utf8');
    const har = JSON.parse(harSource);

    const rebasedHar = makeRebasedHar(har, baseUrl);
    const replaySource = JSON.stringify(rebasedHar);

    await writeFile(replayPath, replaySource);
}

async function initializeReplaySession(context: BrowserContext) {
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

export async function configureHarReplay(context: BrowserContext, baseURL: string, replayHar: URL, testInfo: TestInfo) {
    const replayPath = testInfo.outputPath('recording.har');

    await writeReplayHar(baseURL, replayHar, replayPath);
    await context.routeFromHAR(replayPath, { url: /\/api(?:\/|$)/, notFound: 'abort' });
    await initializeReplaySession(context);
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
