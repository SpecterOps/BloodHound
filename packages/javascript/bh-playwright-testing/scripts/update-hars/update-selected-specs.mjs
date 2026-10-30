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
import { rm } from 'node:fs/promises';
import path from 'node:path';
import { writeUpdatedHar } from '../transform-and-write-hars/write-har.mjs';
import { authenticateForUpdate } from './refresh-update-auth.mjs';
import { captureForUpdate } from '../capture-actions-and-hars/capture-hars.mjs';
import { selectRecordedSpecs } from './select-specs.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

export async function updateSelectedSpecs(specs) {
    const { playwrightCli, authState } = await authenticateForUpdate();
    for (const spec of specs) {
        const { captures, temporaryDirectory } = await captureForUpdate(playwrightCli, spec);
        try {
            await writeUpdatedHar(captures, authState, spec);
            console.log(`Updated ${path.relative(uiDirectory, spec.har)}.`);
        } finally {
            await rm(temporaryDirectory, { recursive: true, force: true });
        }
    }
}

async function main() {
    const specs = await selectRecordedSpecs(path.join(uiDirectory, 'tests', 'e2e'), process.argv.slice(2));
    await updateSelectedSpecs(specs);
}

if (process.argv[1] && path.resolve(process.argv[1]) === new URL(import.meta.url).pathname) {
    main().catch((error) => {
        console.error(error.message);
        process.exitCode = 1;
    });
}
