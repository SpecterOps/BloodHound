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
import test from 'node:test';
import { makeDraftFileName } from './playwright-codegen.mjs';

test('uses the requested draft file name', () => {
    assert.equal(makeDraftFileName('findings-table.ts', 'ignored', 'ignored'), 'findings-table.ts');
    assert.equal(makeDraftFileName('findings table', 'ignored', 'ignored'), 'findings-table.ts');
});

test('uses the branch and a random suffix when no name is supplied', () => {
    assert.equal(
        makeDraftFileName(undefined, 'BED-10004/playwright-hars', 'a1b2c3d4'),
        'BED-10004-playwright-hars-a1b2c3d4.ts'
    );
    assert.equal(makeDraftFileName(undefined, '', 'a1b2c3d4'), 'playwright-draft-a1b2c3d4.ts');
});

test('rejects paths and empty draft names', () => {
    assert.throws(() => makeDraftFileName('../draft.ts', 'ignored', 'ignored'), /without a directory/);
    assert.throws(() => makeDraftFileName('...', 'ignored', 'ignored'), /letter or number/);
});
