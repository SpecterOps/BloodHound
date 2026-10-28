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

import { mkdir, mkdtemp } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { parseBrowserActionsArguments } from './prepare-arguments.mjs';
import { checkBrowserActionsFiles } from './manage-files.mjs';

const uiDirectory = process.cwd();

export async function prepareBrowserActions(args) {
    const { name, requestedUrl } = parseBrowserActionsArguments(args);
    if (!process.env.A11Y_TEST_URL) throw new Error('Set A11Y_TEST_URL in cmd/ui/.env.');
    const files = {
        spec: path.join(uiDirectory, 'tests', 'e2e', `${name}.spec.ts`),
        har: path.join(uiDirectory, 'tests', 'e2e', `${name}.har`),
    };
    checkBrowserActionsFiles(files);
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
        temporaryDirectory,
        storageState: path.join(uiDirectory, 'playwright', '.auth', 'user-light.json'),
        rawSource: path.join(temporaryDirectory, 'browserActions.ts'),
        rawHar: path.join(temporaryDirectory, 'capture.har'),
    };
}
