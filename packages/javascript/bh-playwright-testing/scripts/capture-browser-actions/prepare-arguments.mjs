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

import { parseArgs } from 'node:util';
import { requestCaptureArguments } from '../capture-requests/prepare-arguments.mjs';

export function parseBrowserActionsArguments(args) {
    const usage = 'Usage: yarn codegen <name> [--url /ui/feature]';
    let parsed;
    try {
        parsed = parseArgs({
            args,
            options: { url: { type: 'string' } },
            allowPositionals: true,
        });
    } catch {
        throw new Error(usage);
    }
    const [name] = parsed.positionals;
    const requestedUrl = parsed.values.url ?? '/ui/';
    if (parsed.positionals.length !== 1 || !/^[a-z0-9][a-z0-9_-]*$/i.test(name) || !requestedUrl) {
        throw new Error(usage);
    }
    return { name, requestedUrl };
}

export function recorderArguments(storageState, source, har, url) {
    return [
        'codegen',
        '--load-storage',
        storageState,
        '--output',
        source,
        ...requestCaptureArguments(har),
        '--viewport-size',
        '1280,720',
        url,
    ];
}
