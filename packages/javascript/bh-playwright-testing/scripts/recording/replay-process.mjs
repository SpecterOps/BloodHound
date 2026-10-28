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
import { fileURLToPath } from 'node:url';

export function runReplay(playwrightCli, name) {
    const config = fileURLToPath(new URL('../../playwright.e2e.config.mjs', import.meta.url));
    const result = spawnSync(
        process.execPath,
        [playwrightCli, 'test', '-c', config, `tests/e2e/${name}.spec.ts`, '--project=chromium-light'],
        { cwd: process.cwd(), env: process.env, stdio: 'inherit' }
    );
    if (result.error) throw result.error;
    return result.status ?? 1;
}
