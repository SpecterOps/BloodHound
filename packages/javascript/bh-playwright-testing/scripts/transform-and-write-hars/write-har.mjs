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

import { mkdir, readFile, readdir, rename, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { sanitizeHar } from './sanitize-har.mjs';

export function mergeHarCaptures(capturedHars) {
    const combinedHar = structuredClone(capturedHars[0]);
    const combinedEntries = capturedHars.flatMap((har) => har.log.entries);

    combinedHar.log.entries = structuredClone(combinedEntries);

    return combinedHar;
}

function serializeHar(har) {
    return `${JSON.stringify(har, null, 2)}\n`;
}

async function readHarFile(filePath) {
    const harSource = await readFile(filePath, 'utf8');
    const har = JSON.parse(harSource);

    return har;
}

async function replaceHarFile(filePath, temporaryFilePath, harSource) {
    await writeFile(temporaryFilePath, harSource);
    await rename(temporaryFilePath, filePath);
}

export async function writeNewHar(file, har) {
    const harSource = serializeHar(har);

    await mkdir(path.dirname(file), { recursive: true });
    await writeFile(file, harSource, { flag: 'wx' });
}

export async function writeRecordedHar(recordingSession) {
    const rawHar = await readHarFile(recordingSession.rawHar);
    const authStateSource = await readFile(recordingSession.storageState, 'utf8');
    const authState = JSON.parse(authStateSource);
    const sanitizedHar = sanitizeHar(rawHar, authState);

    if (!sanitizedHar.log.entries.length) {
        throw new Error('No API responses captured. Visit your feature before closing codegen.');
    }

    await writeNewHar(recordingSession.files.har, sanitizedHar);
}

export async function writeUpdatedHar(capturesPath, authState, spec) {
    const directoryFiles = await readdir(capturesPath);
    const captureFiles = directoryFiles.filter((name) => name.endsWith('.har'));

    if (!captureFiles.length) {
        throw new Error(`No HAR files captured for ${spec.name}.`);
    }

    const capturePaths = captureFiles.map((name) => path.join(capturesPath, name));
    const capturedHars = await Promise.all(capturePaths.map(readHarFile));
    const combinedHar = mergeHarCaptures(capturedHars);
    const sanitizedHar = sanitizeHar(combinedHar, authState);

    if (!sanitizedHar.log.entries.length) {
        throw new Error(`No API responses captured for ${spec.name}.`);
    }

    const harSource = serializeHar(sanitizedHar);
    const temporaryFilePath = path.join(path.dirname(capturesPath), 'sanitized.har');

    await replaceHarFile(spec.har, temporaryFilePath, harSource);
}
