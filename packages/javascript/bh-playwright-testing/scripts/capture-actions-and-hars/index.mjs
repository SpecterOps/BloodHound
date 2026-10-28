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
import { buildAndWriteSpec } from '../build-and-write-specs/write-spec.mjs';
import { replayRecordedSpec } from '../run-specs/replay-spec.mjs';
import { writeRecordedHar } from '../transform-and-write-hars/write-har.mjs';
import { ensureStorageState } from './refresh-capture-auth.mjs';
import { launchCodegenBrowser } from './launch-codegen-browser.mjs';
import { announceFilesCreated } from './manage-files.mjs';
import { prepareBrowserActions } from './prepare-browser-actions.mjs';
import { prepareUiServer } from './prepare-ui-server.mjs';

dotenv.config({ path: path.join(process.cwd(), '.env'), quiet: true });

async function runBrowserActions() {
    const browserActions = await prepareBrowserActions(process.argv.slice(2));
    const server = await prepareUiServer(browserActions.homeUrl);
    try {
        await ensureStorageState(browserActions.playwrightCli, browserActions.storageState);
        await launchCodegenBrowser(browserActions);
        await writeRecordedHar(browserActions);
        const source = await buildAndWriteSpec(browserActions, process.env.A11Y_TEST_URL);
        await rm(browserActions.temporaryDirectory, { recursive: true });
        announceFilesCreated(browserActions, source);
        replayRecordedSpec(browserActions);
    } finally {
        if (server) {
            console.log('Stopping the UI server started for codegen.');
            server.kill();
        }
    }
}

runBrowserActions().catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
});
