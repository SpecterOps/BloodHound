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
import { expect, expectNoAccessibilityViolations, test } from '../../fixtures';

test.describe('WCAG A/AA Violations - Explore - Search Tab', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/ui/explore');
    });

    test('Search tab', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByText('Begin typing to search').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Search with results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchTerm = 'test';
        const searchResultName = 'TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [{ name: searchResultName, objectid: 'playwright-search-result', type: 'User' }],
                },
            });
        });

        const searchField = page.getByLabel('Search Nodes');
        await searchField.fill(searchTerm);

        const searchResult = page.getByRole('option').filter({ hasText: searchResultName });
        await expect(searchResult).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Search with no results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchTerm = 'zzzznonexistentnode9999';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });

        const searchField = page.getByLabel('Search Nodes');
        await expect(searchField).toBeVisible();
        await searchField.fill(searchTerm);

        const noResultsMessage = `No results found for "${searchTerm}"`;
        await expect(page.getByText(noResultsMessage)).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
