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
import { access } from 'node:fs/promises';
import path from 'node:path';

const uiDirectory = process.cwd();

export async function ensureStorageState(playwrightCli, storageState) {
    console.log(`Refreshing authentication state at ${path.relative(uiDirectory, storageState)}.`);
    const result = spawnSync(
        process.execPath,
        [playwrightCli, 'test', '-c', 'playwright.a11y.config.ts', '--project=setup'],
        { cwd: uiDirectory, env: process.env, stdio: 'inherit' }
    );
    if (result.error) throw result.error;
    if (result.status !== 0) throw new Error(`Unable to generate authentication state (exit code ${result.status}).`);
    await access(storageState);
    console.log(`Using fresh authentication state at ${path.relative(uiDirectory, storageState)}.`);
}
