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

import { existsSync } from 'node:fs';
import { mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';

export function checkRecordingFiles(files) {
    if (existsSync(files.spec) || existsSync(files.har)) {
        throw new Error(
            'A spec or HAR with this name already exists. Use yarn test:e2e:update-mocks to replace its HAR.'
        );
    }
}

export async function saveRecording(files, har, spec) {
    checkRecordingFiles(files);
    await mkdir(path.dirname(files.har), { recursive: true });
    await writeFile(files.har, `${JSON.stringify(har, null, 2)}\n`, { flag: 'wx' });
    await writeFile(files.spec, spec, { flag: 'wx' });
}
