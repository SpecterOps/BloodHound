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

import { test as base } from '@playwright/test';
import { access, copyFile, mkdir } from 'node:fs/promises';
import path from 'node:path';
import { makeHarArtifactPaths, publishHar, readHar, writeJsonArtifact } from './har-artifacts-utils';
import type { HarArtifactPaths, HarMode, HarNotFound } from './har-artifacts-utils';

type HarFixtures = {
    /** HAR mode for the test context. Defaults to off. */
    harMode: HarMode;
    /** Canonical fixture root, relative to the Playwright config root or absolute. */
    harRootDir: string;
    /** Behavior for requests missing from a replayed HAR. Defaults to abort. */
    harNotFound: HarNotFound;
    /** Canonical paths for this test, stable across retries and runs. */
    harArtifacts: HarArtifactPaths;
    harSession: { contextStarted: boolean };
};

export type HarOptions = Pick<HarFixtures, 'harMode' | 'harRootDir' | 'harNotFound'>;

// Extend this test so HAR setup precedes page requests and teardown publishes after context closure.
export const test = base.extend<HarFixtures>({
    harMode: ['off', { option: true }],
    harRootDir: ['test-artifacts/har', { option: true }],
    harNotFound: ['abort', { option: true }],
    harArtifacts: async ({ harRootDir }, use, testInfo) => {
        const paths = makeHarArtifactPaths(testInfo, harRootDir);
        await use(paths);
    },
    harSession: async ({ harMode: _harMode }, use) => {
        void _harMode;
        await use({ contextStarted: false });
    },
    contextOptions: async ({ contextOptions, harMode, harArtifacts, harSession }, use, testInfo) => {
        if (harMode === 'off' || harMode === 'mock') {
            await use(contextOptions);
            return;
        }

        // Capture to this attempt's output directory before replacing the shared fixture.
        const temporaryRecordingPath = testInfo.outputPath('har', 'recording.har');
        await mkdir(harArtifacts.directory, { recursive: true });
        await mkdir(path.dirname(temporaryRecordingPath), { recursive: true });
        if (harMode === 'update') {
            await readHar(harArtifacts.recording);
            await copyFile(harArtifacts.recording, temporaryRecordingPath);
        }

        if (harMode === 'record') {
            await use({
                ...contextOptions,
                recordHar: { path: temporaryRecordingPath, content: 'embed', mode: 'full' },
            });
        } else {
            // Updates are captured by routeFromHAR in the context fixture.
            await use(contextOptions);
        }

        // Record mode preserves captures even after failures. A failed update must keep the old fixture.
        if (harMode === 'update' && testInfo.status !== 'passed') return;

        try {
            await access(temporaryRecordingPath);
        } catch {
            // Setup may have failed before Playwright created a browser context.
            if (!harSession.contextStarted) return;
            throw new Error(`HAR recording was not written after context close: ${temporaryRecordingPath}`);
        }
        await publishHar(temporaryRecordingPath, harArtifacts);
        await writeJsonArtifact(path.join(harArtifacts.directory, '.har-fixture.json'), {
            testId: testInfo.testId,
            project: testInfo.project.name,
            rootDir: testInfo.config.rootDir,
        });
    },
    context: async ({ context, harMode, harNotFound, harArtifacts, harSession }, use, testInfo) => {
        const failedRequests: string[] = [];
        if (harMode === 'mock') {
            await readHar(harArtifacts.recording);
            if (harNotFound === 'abort') {
                context.on('requestfailed', (request) => {
                    const failureReason = request.failure()?.errorText ?? 'failed';
                    failedRequests.push(`${request.method()} ${request.url()}: ${failureReason}`);
                });
            }
            await context.routeFromHAR(harArtifacts.recording, { notFound: harNotFound });
        } else if (harMode === 'update') {
            await readHar(harArtifacts.recording);
            await context.routeFromHAR(testInfo.outputPath('har', 'recording.har'), {
                update: true,
                updateContent: 'embed',
                updateMode: 'full',
            });
        }
        harSession.contextStarted = true;
        await use(context);
        if (harMode === 'record' || harMode === 'update') {
            // Native HAR output is written by close(). Playwright's base context fixture
            // also closes the context, but its teardown can follow contextOptions teardown.
            await context.close();
        }
        if (failedRequests.length) {
            throw new Error(`HAR mock blocked unexpected or failed requests:\n${failedRequests.join('\n')}`);
        }
    },
});
