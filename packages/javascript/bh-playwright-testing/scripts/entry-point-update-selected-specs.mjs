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
import path from 'node:path';
import { selectRecordedSpecs } from './update-hars/select-specs.mjs';
import { updateSelectedSpecs } from './update-hars/update-selected-specs.mjs';

const uiDirectory = process.cwd();
dotenv.config({ path: path.join(uiDirectory, '.env'), quiet: true });

async function selectAndUpdateSpecs() {
    const specsDirectory = path.join(uiDirectory, 'tests', 'e2e');
    const commandLineArguments = process.argv.slice(2);
    const specs = await selectRecordedSpecs(specsDirectory, commandLineArguments);
    await updateSelectedSpecs(specs);
}

if (process.argv[1] && path.resolve(process.argv[1]) === new URL(import.meta.url).pathname) {
    selectAndUpdateSpecs().catch((error) => {
        console.error(error.message);
        process.exitCode = 1;
    });
}
