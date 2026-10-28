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
import { parseRecordingArguments } from './arguments.mjs';
import { ensureStorageState } from './auth.mjs';
import { buildRecordedTest, captureBrowser } from './capture.mjs';
import { checkRecordingFiles, saveRecording } from './files.mjs';
import { recordingFiles, temporaryRecordingPaths } from './paths.mjs';
import { announceRecording, replayRecordedTest } from './replay.mjs';
import { ensureUiServer } from './server.mjs';

export { parseRecordingArguments, recorderArguments } from './arguments.mjs';
export { checkRecordingFiles, saveRecording } from './files.mjs';
export { makeDevServerArguments } from './server.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

async function prepareRecording(args) {
    const { name, requestedUrl } = parseRecordingArguments(args);
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    const files = recordingFiles(uiDirectory, name);
    checkRecordingFiles(files);
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
        ...temporaryRecordingPaths(uiDirectory, temporaryDirectory),
    };
}

export async function runRecording() {
    const recording = await prepareRecording(process.argv.slice(2));
    const server = await ensureUiServer(recording.homeUrl);
    try {
        await ensureStorageState(recording.playwrightCli, recording.storageState);
        await captureBrowser(recording);
        const { source, har, spec } = await buildRecordedTest(recording);
        await saveRecording(recording.files, har, spec);
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
