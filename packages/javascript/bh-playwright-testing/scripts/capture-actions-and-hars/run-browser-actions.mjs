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

import dotenv from 'dotenv';
import { mkdir, mkdtemp, rm } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { buildAndWriteSpec } from '../build-and-write-specs/write-spec.mjs';
import { replayRecordedSpec } from '../run-specs/replay-spec.mjs';
import { writeRecordedHar } from '../transform-and-write-hars/write-har.mjs';
import { refreshAuthState } from './refresh-auth-state.mjs';
import { launchCodegenBrowser, prepareUiServer } from './launch-codegen-browser.mjs';
import { announceFilesCreated, checkForDuplicateSpecName } from './manage-files.mjs';
import { parseRecordingSessionArguments } from './prepare-arguments.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

async function makeCodegenRecordingSessionObject(args) {
    const { name, requestedUrl } = parseRecordingSessionArguments(args);
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    const files = {
        spec: path.join(uiDirectory, 'tests', 'e2e', `${name}.spec.ts`),
        har: path.join(uiDirectory, 'tests', 'e2e', `${name}.har`),
    };

    checkForDuplicateSpecName(files);

    const homeUrl = new URL(requestedUrl, process.env.A11Y_TEST_URL).toString();

    if (new URL(homeUrl).origin !== new URL(process.env.A11Y_TEST_URL).origin) {
        throw new Error('--url must use the configured A11Y_TEST_URL origin.');
    }

    const playwrightCli = createRequire(path.join(uiDirectory, 'package.json')).resolve('@playwright/test/cli');

    await mkdir(path.join(uiDirectory, 'playwright'), { recursive: true });

    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'codegen-'));

    return {
        name,
        files,
        homeUrl,
        playwrightCli,
        temporaryDirectory,
        storageState: path.join(uiDirectory, 'playwright', '.auth', 'user-light.json'),
        rawSource: path.join(temporaryDirectory, 'recordingSession.ts'),
        rawHar: path.join(temporaryDirectory, 'capture.har'),
    };
}

async function recordAndVerifyAutoMockedCodegenSpec() {
    const recordingSession = await makeCodegenRecordingSessionObject(process.argv.slice(2));
    const server = await prepareUiServer(recordingSession.homeUrl);
    try {
        await refreshAuthState(recordingSession.playwrightCli, recordingSession.storageState);
        await launchCodegenBrowser(recordingSession);
        await writeRecordedHar(recordingSession);

        const source = await buildAndWriteSpec(recordingSession, process.env.A11Y_TEST_URL);

        await rm(recordingSession.temporaryDirectory, { recursive: true });

        announceFilesCreated(recordingSession, source);
        replayRecordedSpec(recordingSession);
    } finally {
        if (server) {
            console.log('Stopping the UI server started for codegen.');
            server.kill();
        }
    }
}

recordAndVerifyAutoMockedCodegenSpec().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
});
