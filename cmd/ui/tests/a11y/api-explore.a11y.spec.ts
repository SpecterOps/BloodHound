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

import { expect, expectNoAccessibilityViolations, test } from '../fixtures';

test.describe('API Explorer page accessibility', () => {
    test('explore page has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/api-explorer');

        // Wait for the filter input to load
        await page.getByRole('textbox', { name: 'Filter by tag or path' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('expanded resource', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/api-explorer');

        // Wait for the filter input to load
        await page.getByRole('textbox', { name: 'Filter by tag or path' }).waitFor({ state: 'visible' });

        const resourceButton = page.getByTestId('api-explorer').getByRole('button', { name: /^get.*api.*version$/i });

        await resourceButton.click();
        await expect(resourceButton).toHaveAttribute('aria-expanded', 'true');

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('expanded disabled resource', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/api-explorer');

        // Wait for the filter input to load
        await page.getByRole('textbox', { name: 'Filter by tag or path' }).waitFor({ state: 'visible' });

        const resourceButton = page
            .getByTestId('api-explorer')
            .getByRole('button', { name: /^put.*api.*v2.*accept-eula$/i });

        await resourceButton.click();
        await expect(resourceButton).toHaveAttribute('aria-expanded', 'true');

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('filter with no results', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/api-explorer');

        // Wait for the filter input to load
        const filterInput = page.getByRole('textbox', { name: 'Filter by tag or path' });
        await filterInput.waitFor({ state: 'visible' });

        await filterInput.fill('no-matching-api-resource');

        const apiExplorer = page.getByTestId('api-explorer');
        await expect(
            apiExplorer.getByRole('button', { name: /^(delete|get|head|options|patch|post|put|trace)\b/i })
        ).toHaveCount(0);

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('expanded Schemas', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/api-explorer');

        // Wait for the filter input to load
        await page.getByRole('textbox', { name: 'Filter by tag or path' }).waitFor({ state: 'visible' });

        const schemasButton = page.getByTestId('api-explorer').getByRole('button', { name: 'Schemas' });

        await expect(schemasButton).toHaveAttribute('aria-expanded', 'true');

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
