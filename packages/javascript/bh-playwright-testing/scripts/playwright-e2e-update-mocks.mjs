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
import { spawnSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { access, mkdir, mkdtemp, readFile, readdir, rename, rm, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import { sanitizeRecording } from './playwright-mocked-codegen.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

export async function selectRecordedTests(directory, args) {
    if (args.length > 1 || (args[0] && !/^[a-z0-9][a-z0-9_-]*$/i.test(args[0]))) {
        throw new Error('Usage: yarn test:e2e:update-mocks [name]');
    }
    const names = args[0]
        ? [args[0]]
        : (await readdir(directory)).filter((name) => name.endsWith('.spec.ts')).map((name) => name.slice(0, -8));
    if (!names.length) throw new Error('No recorded tests found.');
    return names.map((name) => {
        const spec = path.join(directory, `${name}.spec.ts`);
        const har = path.join(directory, `${name}.har`);
        if (!existsSync(spec) || !existsSync(har)) throw new Error(`Missing spec or HAR for ${name}.`);
        return { name, spec, har };
    });
}

export function runPlaywright(playwrightCli, args, env = process.env) {
    const result = spawnSync(process.execPath, [playwrightCli, ...args], {
        cwd: uiDirectory,
        env,
        stdio: 'inherit',
    });
    if (result.error) throw result.error;
    if (result.status !== 0) throw new Error(`Playwright failed (exit code ${result.status}).`);
}

async function updateOneTest(playwrightCli, test, authState) {
    const temporaryDirectory = await mkdtemp(path.join(uiDirectory, 'playwright', 'har-update-'));
    const captures = path.join(temporaryDirectory, 'captures');
    await mkdir(captures);
    try {
        runPlaywright(
            playwrightCli,
            [
                'test',
                '-c',
                fileURLToPath(new URL('../playwright.e2e.config.mjs', import.meta.url)),
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
        const captureFiles = (await readdir(captures)).filter((name) => name.endsWith('.har'));
        if (!captureFiles.length) throw new Error(`No HAR files captured for ${test.name}.`);
        const recordings = await Promise.all(
            captureFiles.map(async (name) => JSON.parse(await readFile(path.join(captures, name), 'utf8')))
        );
        const combined = recordings[0];
        combined.log.entries = recordings.flatMap((recording) => recording.log.entries);
        const har = sanitizeRecording(combined, authState);
        if (!har.log.entries.length) throw new Error(`No API responses captured for ${test.name}.`);
        const sanitized = path.join(temporaryDirectory, 'sanitized.har');
        await writeFile(sanitized, `${JSON.stringify(har, null, 2)}\n`);
        await rename(sanitized, test.har);
        console.log(`Updated ${path.relative(uiDirectory, test.har)}.`);
    } finally {
        await rm(temporaryDirectory, { recursive: true, force: true });
    }
}

async function main() {
    const tests = await selectRecordedTests(path.join(uiDirectory, 'tests', 'e2e'), process.argv.slice(2));
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    const playwrightCli = createRequire(path.join(uiDirectory, 'package.json')).resolve('@playwright/test/cli');
    const storageState = path.join(uiDirectory, 'playwright', '.auth', 'user-light.json');
    await mkdir(path.dirname(storageState), { recursive: true });
    runPlaywright(playwrightCli, ['test', '-c', 'playwright.a11y.config.ts', '--project=setup']);
    await access(storageState);
    const authState = JSON.parse(await readFile(storageState, 'utf8'));
    for (const test of tests) await updateOneTest(playwrightCli, test, authState);
}

if (process.argv[1] && path.resolve(process.argv[1]) === new URL(import.meta.url).pathname) {
    main().catch((error) => {
        console.error(error.message);
        process.exitCode = 1;
    });
}
