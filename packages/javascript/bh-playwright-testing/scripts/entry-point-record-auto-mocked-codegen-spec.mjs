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
import { rm } from 'node:fs/promises';
import path from 'node:path';
import { buildAndWriteSpec } from './build-and-write-specs/write-spec.mjs';
import { launchCodegenBrowser, prepareUiServer } from './capture-actions-and-hars/launch-codegen-browser.mjs';
import { makeCodegenRecordingSessionObject } from './capture-actions-and-hars/make-recording-session.mjs';
import { announceFilesCreated } from './capture-actions-and-hars/manage-files.mjs';
import { refreshAuthState } from './capture-actions-and-hars/refresh-auth-state.mjs';
import { replayRecordedSpec } from './run-specs/replay-spec.mjs';
import { writeRecordedHar } from './transform-and-write-hars/write-har.mjs';

dotenv.config({ path: path.join(process.cwd(), '.env'), quiet: true });

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
