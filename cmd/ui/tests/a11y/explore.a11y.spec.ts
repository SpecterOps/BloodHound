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

import { expectNoAccessibilityViolations, test } from '../fixtures';

test.describe('Explore page accessibility', () => {
    test('explore page has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/explore');

        // Wait for search widget to load
        await page.getByLabel('Toggle search widget').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
