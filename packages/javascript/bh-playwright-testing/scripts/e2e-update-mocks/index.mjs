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
import { access, mkdir, readFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { runPlaywright } from './run-playwright.mjs';
import { selectRecordedTests } from './select-tests.mjs';
import { updateOneTest } from './update-one.mjs';

export { selectRecordedTests } from './select-tests.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

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
