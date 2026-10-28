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
import { mkdir, mkdtemp, rm } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const uiDirectory = process.cwd();

export function runPlaywright(playwrightCli, args, env = process.env) {
    const result = spawnSync(process.execPath, [playwrightCli, ...args], {
        cwd: uiDirectory,
        env,
        stdio: 'inherit',
    });
    if (result.error) throw result.error;
    if (result.status !== 0) throw new Error(`Playwright failed (exit code ${result.status}).`);
}

function captureTest(playwrightCli, test, captures) {
    runPlaywright(
        playwrightCli,
        [
            'test',
            '-c',
            fileURLToPath(new URL('../../playwright.e2e.config.mjs', import.meta.url)),
            path.relative(uiDirectory, test.spec),
            '--project=chromium-light',
            '--workers=1',
        ],
        {
            ...process.env,
            HAR_UPDATE_CAPTURE: captures,
            HAR_UPDATE_STORAGE_STATE: path.join(uiDirectory, 'playwright', '.auth', 'user-light.json'),
        }
    );
}

export async function captureForUpdate(playwrightCli, test) {
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'har-update-'));
    const captures = path.join(temporaryDirectory, 'captures');
    await mkdir(captures);
    try {
        captureTest(playwrightCli, test, captures);
        return { captures, temporaryDirectory };
    } catch (error) {
        await rm(temporaryDirectory, { recursive: true, force: true });
        throw error;
    }
}
