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

import { readFile } from 'node:fs/promises';
import { selectFailingSpecs } from './select-failures.mjs';

export async function readFailures(reportFile, specs) {
    let report;

    try {
        const reportSource = await readFile(reportFile, 'utf8');
        report = JSON.parse(reportSource);
    } catch {
        throw new Error('E2E specs failed before Playwright could report the failing specs.');
    }

    const failed = selectFailingSpecs(report, specs);

    if (!failed.length) throw new Error('E2E specs failed, but no recorded spec was identified as failing.');

    return failed;
}
