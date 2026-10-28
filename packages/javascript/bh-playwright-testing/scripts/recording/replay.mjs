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

import { spawnSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const uiDirectory = process.cwd();

export function announceRecording(recording, source) {
    if (!/\bexpect\s*\(/.test(source)) {
        console.log('No assertions were recorded. Add an expected result before treating this as regression coverage.');
    }
    console.log(
        `Created ${path.relative(uiDirectory, recording.files.spec)} and ${path.relative(uiDirectory, recording.files.har)}.`
    );
    console.log('Review the assertions and captured data before committing.');
    console.log(`Run: yarn test:e2e tests/e2e/${recording.name}.spec.ts`);
}

export function replayRecordedTest(recording) {
    console.log('Checking the new test with API replay and no live login...');
    const config = fileURLToPath(new URL('../../playwright.e2e.config.mjs', import.meta.url));
    const result = spawnSync(
        process.execPath,
        [
            recording.playwrightCli,
            'test',
            '-c',
            config,
            `tests/e2e/${recording.name}.spec.ts`,
            '--project=chromium-light',
        ],
        { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
    );
    if (result.error) throw result.error;
    const exitCode = result.status ?? 1;
    process.exitCode = exitCode;
    console.log(
        exitCode === 0
            ? 'Mocked test passed. Review the spec and HAR before committing.'
            : 'Replay failed; the files are saved. Review the failure and record again if needed.'
    );
}
