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

import assert from 'node:assert/strict';
import path from 'node:path';
import test from 'node:test';
import { selectFailingTests } from './update-hars/failed-tests.mjs';

test('selects only recorded specs with unexpected failures', () => {
    const rootDir = path.join(process.cwd(), 'tests', 'e2e');
    const tests = ['stale', 'passing', 'nested', 'flaky'].map((name) => ({
        name,
        spec: path.join(rootDir, `${name}.spec.ts`),
    }));
    const report = {
        config: { rootDir },
        suites: [
            {
                specs: [
                    { file: 'stale.spec.ts', tests: [{ status: 'unexpected' }] },
                    { file: 'passing.spec.ts', tests: [{ status: 'expected' }] },
                    { file: 'flaky.spec.ts', tests: [{ status: 'flaky' }] },
                ],
                suites: [{ specs: [{ file: 'nested.spec.ts', tests: [{ status: 'unexpected' }] }] }],
            },
        ],
    };
    assert.deepEqual(selectFailingTests(report, tests), [tests[0], tests[2]]);
});
