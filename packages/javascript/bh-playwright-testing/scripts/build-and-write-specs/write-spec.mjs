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
import { makeMockedSpec } from './build-spec.mjs';

export async function writeSpec(file, spec) {
    await mkdir(path.dirname(file), { recursive: true });
    await writeFile(file, spec, { flag: 'wx' });
}

export function extractLicenseHeader(source) {
    return source.split('\n\n')[0];
}

export async function buildAndWriteSpec(recordingSession, baseUrl) {
    const source = await readFile(recordingSession.rawSource, 'utf8');
    const sourceFilePath = fileURLToPath(import.meta.url);
    const sourceFile = await readFile(sourceFilePath, 'utf8');
    const licenseHeader = extractLicenseHeader(sourceFile);
    const mockedSpec = makeMockedSpec(source, recordingSession.name, baseUrl, licenseHeader);
    const prettierConfig = await prettier.resolveConfig(recordingSession.files.spec);

    const spec = await prettier.format(mockedSpec, {
        ...prettierConfig,
        parser: 'typescript',
    });

    await writeSpec(recordingSession.files.spec, spec);
    return source;
}
