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

test.describe('Marketplace page accessibility', () => {
    test('Marketplace content has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/marketplace');
        await page.getByRole('heading', { name: 'Marketplace', exact: true }).waitFor({ state: 'visible' });
        await page.getByRole('region', { name: 'Community Extensions' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });

        const search = page.getByRole('searchbox', { name: 'Search marketplace' });
        const typeFilter = page.getByRole('combobox', { name: 'Filter Marketplace items by type' });
        const publisherFilter = page.getByRole('combobox', { name: 'Filter Marketplace items by publisher' });
        const availabilityFilter = page.getByRole('combobox', {
            name: 'Filter Marketplace items by availability',
        });

        await search.focus();
        await page.keyboard.press('Tab');
        await expect(typeFilter).toBeFocused();
        await page.keyboard.press('Tab');
        await expect(publisherFilter).toBeFocused();
        await page.keyboard.press('Tab');
        await expect(availabilityFilter).toBeFocused();
    });
});
