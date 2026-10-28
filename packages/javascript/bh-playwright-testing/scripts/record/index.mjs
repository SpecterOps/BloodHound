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
import { buildAndWriteSpec } from '../build-and-write-specs/write.mjs';
import { replayRecordedTest } from '../run-tests/replay.mjs';
import { writeRecordedHar } from '../transform-and-write-hars/write.mjs';
import { ensureStorageState } from './auth.mjs';
import { captureBrowser } from './capture.mjs';
import { announceRecording } from './files.mjs';
import { prepareRecording } from './prepare.mjs';
import { ensureUiServer } from './server.mjs';

dotenv.config({ path: path.join(process.cwd(), '.env'), quiet: true });

async function runRecording() {
    const recording = await prepareRecording(process.argv.slice(2));
    const server = await ensureUiServer(recording.homeUrl);
    try {
        await ensureStorageState(recording.playwrightCli, recording.storageState);
        await captureBrowser(recording);
        await writeRecordedHar(recording);
        const source = await buildAndWriteSpec(recording, process.env.A11Y_TEST_URL);
        await rm(recording.temporaryDirectory, { recursive: true });
        announceRecording(recording, source);
        replayRecordedTest(recording);
    } finally {
        if (server) {
            console.log('Stopping the UI server started for codegen.');
            server.kill();
        }
    }
}

runRecording().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
});
