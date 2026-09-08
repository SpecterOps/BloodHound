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

import { expectNoAccessibilityViolations, test } from '../../fixtures';

test.describe('Simple Environment Selector - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/available-domains', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            collected: true,
                            exposures: [],
                            hygiene_attack_paths: 0,
                            id: 'example-domain-id',
                            impactValue: 0,
                            name: 'EXAMPLE.COM',
                            type: 'active-directory',
                        },
                    ],
                },
            });
        });

        await page.goto('/ui/administration/data-quality');
        await page.getByRole('heading', { name: 'Data Quality', exact: true }).waitFor({ state: 'visible' });
        await page.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).waitFor({ state: 'visible' });
    });

    test('no query', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).click();

        const popover = page.getByTestId('data-quality_context-selector-popover');
        await popover.waitFor({ state: 'visible' });
        await popover.getByRole('textbox', { name: 'Search' }).waitFor({ state: 'visible' });
        await popover.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder()
            .include('[data-testid="data-quality_context-selector-popover"]')
            .analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with query', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).click();

        const popover = page.getByTestId('data-quality_context-selector-popover');
        await popover.waitFor({ state: 'visible' });

        const searchInput = popover.getByRole('textbox', { name: 'Search' });
        await searchInput.waitFor({ state: 'visible' });
        await searchInput.fill('EXAMPLE');

        await searchInput.and(page.locator('[value="EXAMPLE"]')).waitFor({ state: 'visible' });
        await popover.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder()
            .include('[data-testid="data-quality_context-selector-popover"]')
            .analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('no results query', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'zzzznonexistentenvironment9999';

        await page.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).click();

        const popover = page.getByTestId('data-quality_context-selector-popover');
        await popover.waitFor({ state: 'visible' });

        const searchInput = popover.getByRole('textbox', { name: 'Search' });
        await searchInput.waitFor({ state: 'visible' });
        await searchInput.fill(query);

        await searchInput.and(page.locator(`[value="${query}"]`)).waitFor({ state: 'visible' });
        await popover.getByRole('button', { name: 'EXAMPLE.COM', exact: true }).waitFor({ state: 'hidden' });

        const results = await makeAxeBuilder()
            .include('[data-testid="data-quality_context-selector-popover"]')
            .analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
