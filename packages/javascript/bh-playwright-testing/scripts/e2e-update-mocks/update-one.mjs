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

import { mkdir, mkdtemp, readFile, readdir, rename, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { sanitizeRecording } from '../mocked-codegen/index.mjs';
import { runPlaywright } from './run-playwright.mjs';

const uiDirectory = process.cwd();

async function readCapturedHar(captures, authState, testName) {
    const captureFiles = (await readdir(captures)).filter((name) => name.endsWith('.har'));
    if (!captureFiles.length) throw new Error(`No HAR files captured for ${testName}.`);
    const recordings = await Promise.all(
        captureFiles.map(async (name) => JSON.parse(await readFile(path.join(captures, name), 'utf8')))
    );
    const combined = recordings[0];
    combined.log.entries = recordings.flatMap((recording) => recording.log.entries);
    const har = sanitizeRecording(combined, authState);
    if (!har.log.entries.length) throw new Error(`No API responses captured for ${testName}.`);
    return har;
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

export async function updateOneTest(playwrightCli, test, authState) {
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'har-update-'));
    const captures = path.join(temporaryDirectory, 'captures');
    await mkdir(captures);
    try {
        captureTest(playwrightCli, test, captures);
        const har = await readCapturedHar(captures, authState, test.name);
        const sanitized = path.join(temporaryDirectory, 'sanitized.har');
        await writeFile(sanitized, `${JSON.stringify(har, null, 2)}\n`);
        await rename(sanitized, test.har);
        console.log(`Updated ${path.relative(uiDirectory, test.har)}.`);
    } finally {
        await rm(temporaryDirectory, { recursive: true, force: true });
    }
}
