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

import { access, mkdir, readFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { runPlaywright, updateOneTest } from './update-one.mjs';

const uiDirectory = process.cwd();

export async function updateSelectedTests(tests) {
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    const playwrightCli = createRequire(path.join(uiDirectory, 'package.json')).resolve('@playwright/test/cli');
    const storageState = path.join(uiDirectory, 'playwright', '.auth', 'user-light.json');
    await mkdir(path.dirname(storageState), { recursive: true });
    runPlaywright(playwrightCli, ['test', '-c', 'playwright.a11y.config.ts', '--project=setup']);
    await access(storageState);
    const authState = JSON.parse(await readFile(storageState, 'utf8'));
    for (const test of tests) await updateOneTest(playwrightCli, test, authState);
}
