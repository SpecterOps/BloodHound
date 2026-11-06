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

import { rm } from 'node:fs/promises';
import path from 'node:path';
import { captureForUpdate } from '../capture-actions-and-hars/capture-hars.mjs';
import { writeUpdatedHar } from '../transform-and-write-hars/write-har.mjs';
import { authenticateForUpdate } from './refresh-update-auth.mjs';

const uiDirectory = process.cwd();

export async function updateSelectedSpecs(specs) {
    const { playwrightCli, authState } = await authenticateForUpdate();

    for (const spec of specs) {
        const { capturesPath, temporaryDirectory } = await captureForUpdate(playwrightCli, spec);

        try {
            await writeUpdatedHar(capturesPath, authState, spec);
            console.log(`Updated ${path.relative(uiDirectory, spec.har)}.`);
        } finally {
            await rm(temporaryDirectory, { recursive: true, force: true });
        }
    }
}
