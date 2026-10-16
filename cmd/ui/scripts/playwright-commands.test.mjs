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
import { resolveTestArguments } from './playwright-a11y.mjs';
import { resolveBrowsers } from './playwright-install.mjs';

test('browser installation defaults to the supported test browsers', () => {
    assert.deepEqual(resolveBrowsers([]), ['chromium', 'firefox']);
    assert.deepEqual(resolveBrowsers(['chromium']), ['chromium']);
});

test('accessibility tests default to the configured complete suite', () => {
    assert.deepEqual(resolveTestArguments([]), ['test', '-c', 'playwright.a11y.config.ts']);
    assert.deepEqual(resolveTestArguments(['login.a11y.spec.ts', '--project=chromium-light']), [
        'test',
        '-c',
        'playwright.a11y.config.ts',
        'login.a11y.spec.ts',
        '--project=chromium-light',
    ]);
});
