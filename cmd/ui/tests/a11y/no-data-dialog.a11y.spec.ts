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

// The exact cypher query issued by `useGraphHasData`. The shared a11y fixture stubs this query
// with a populated payload by default; this spec installs a higher-priority handler that returns
// an empty payload so the "No Data Available" upload dialog renders and can be scanned.
const GRAPH_HAS_DATA_QUERY = 'MATCH (A) WHERE NOT A:MigrationData RETURN A LIMIT 1';

test.describe('No Data Available dialog accessibility', () => {
    test('upload dialog has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
        // Playwright runs route handlers in LIFO order, so this test-local handler wins for the
        // graph-has-data probe and overrides the fixture's populated payload with an empty one.
        // Other cypher requests fall through to the fixture (and then to the real backend).
        await page.route('**/api/v2/graphs/cypher', async (route) => {
            const body = route.request().postDataJSON();
            if (body?.query === GRAPH_HAS_DATA_QUERY) {
                return route.fulfill({
                    json: { data: { nodes: {}, edges: {} } },
                });
            }
            return route.fallback();
        });

        await page.goto('/ui/explore');

        // Wait for the dialog to render before proceeding.
        await page
            .getByRole('heading', { name: 'Upload Data to Start Mapping Your Environment' })
            .waitFor({ state: 'visible' });

        // Scope the scan to the dialog (rendered as `role="dialog"` by MUI) so violations from
        // the rest of the Explore page don't bleed into this test's signal.
        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
