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

import { test as playwrightTest } from 'bh-playwright-testing';
import { installGraphHasDataStub } from 'bh-playwright-testing/stubs';

// Wraps `bh-playwright-testing`'s `test` with a `page` fixture that installs the shared cypher
// stub so `useGraphHasData` resolves to "true" and the "No Data Available" upload dialog never
// settles open in accessibility tests that don't care about that state.
//
// Individual tests can override the stub by registering their own `page.route` for the cypher
// endpoint — Playwright runs handlers in LIFO order, so a test-local handler wins for the cases
// it cares about.
export const test = playwrightTest.extend({
    page: async ({ page }, use) => {
        await installGraphHasDataStub(page);
        await use(page);
    },
});

export { expect, expectNoAccessibilityViolations } from 'bh-playwright-testing';
