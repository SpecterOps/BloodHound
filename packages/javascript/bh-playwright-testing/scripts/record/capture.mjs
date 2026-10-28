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

import { spawn } from 'node:child_process';
import { recorderArguments } from './arguments.mjs';

const uiDirectory = process.cwd();

export async function captureBrowser(recording) {
    console.log(`Opening the Playwright browser at ${recording.homeUrl}; this may take a few seconds...`);
    console.log('Record the flow, add assertions, then close the browser to save.');
    const result = await new Promise((resolve, reject) => {
        const child = spawn(
            process.execPath,
            [
                recording.playwrightCli,
                ...recorderArguments(recording.storageState, recording.rawSource, recording.rawHar, recording.homeUrl),
            ],
            { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
        );
        child.on('error', reject);
        child.on('exit', (code, signal) => resolve({ code, signal }));
    });
    if (result.signal || result.code !== 0)
        throw new Error(
            `Recording did not finish successfully. Temporary files remain in ${recording.temporaryDirectory}.`
        );
}
