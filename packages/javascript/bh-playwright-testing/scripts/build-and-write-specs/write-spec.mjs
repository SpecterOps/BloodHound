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

import { mkdir, readFile, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import prettier from 'prettier';
import { makeMockedTest } from './build-spec.mjs';

export async function writeSpec(file, spec) {
    await mkdir(path.dirname(file), { recursive: true });
    await writeFile(file, spec, { flag: 'wx' });
}

export async function buildAndWriteSpec(recording, baseUrl) {
    const source = await readFile(recording.rawSource, 'utf8');
    const licenseHeader = (await readFile(fileURLToPath(import.meta.url), 'utf8')).split('\n\n')[0];
    const spec = await prettier.format(makeMockedTest(source, recording.name, baseUrl, licenseHeader), {
        ...(await prettier.resolveConfig(recording.files.spec)),
        parser: 'typescript',
    });
    await writeSpec(recording.files.spec, spec);
    return source;
}
