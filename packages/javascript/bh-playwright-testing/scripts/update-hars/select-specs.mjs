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
import { readdir } from 'node:fs/promises';
import path from 'node:path';

function recordedSpec(directory, name) {
    const spec = path.join(directory, `${name}.spec.ts`);
    const har = path.join(directory, `${name}.har`);
    if (!existsSync(spec) || !existsSync(har)) throw new Error(`Missing spec or HAR for ${name}.`);
    return { name, spec, har };
}

export async function listRecordedSpecs(directory) {
    const names = (await readdir(directory))
        .filter((name) => name.endsWith('.spec.ts'))
        .map((name) => name.slice(0, -8));
    if (!names.length) throw new Error('No recorded specs found.');
    return names.map((name) => recordedSpec(directory, name));
}

export async function selectRecordedSpecs(directory, args) {
    if (args.length !== 1 || !/^[a-z0-9][a-z0-9_-]*$/i.test(args[0])) {
        throw new Error('Usage: yarn test:e2e:update-mocks <name>');
    }
    return [recordedSpec(directory, args[0])];
}
