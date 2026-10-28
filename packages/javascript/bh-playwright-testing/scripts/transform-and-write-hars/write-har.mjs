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
import { sanitizeRecording } from './sanitize-recording.mjs';

export async function writeNewHar(file, har) {
    await mkdir(path.dirname(file), { recursive: true });
    await writeFile(file, `${JSON.stringify(har, null, 2)}\n`, { flag: 'wx' });
}

export async function writeRecordedHar(recording) {
    const rawHar = JSON.parse(await readFile(recording.rawHar, 'utf8'));
    const authState = JSON.parse(await readFile(recording.storageState, 'utf8'));
    const har = sanitizeRecording(rawHar, authState);
    if (!har.log.entries.length) {
        throw new Error('No API responses captured. Visit your feature before closing codegen.');
    }
    await writeNewHar(recording.files.har, har);
}

export async function writeUpdatedHar(captures, authState, spec) {
    const files = (await readdir(captures)).filter((name) => name.endsWith('.har'));
    if (!files.length) throw new Error(`No HAR files captured for ${spec.name}.`);
    const recordings = await Promise.all(
        files.map(async (name) => JSON.parse(await readFile(path.join(captures, name), 'utf8')))
    );
    const combined = recordings[0];
    combined.log.entries = recordings.flatMap((recording) => recording.log.entries);
    const har = sanitizeRecording(combined, authState);
    if (!har.log.entries.length) throw new Error(`No API responses captured for ${spec.name}.`);
    const temporaryFile = path.join(path.dirname(captures), 'sanitized.har');
    await writeFile(temporaryFile, `${JSON.stringify(har, null, 2)}\n`);
    await rename(temporaryFile, spec.har);
}
