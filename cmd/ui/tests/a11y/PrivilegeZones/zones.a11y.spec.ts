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

import {
    installAssetGroupTagMemberStub,
    installAssetGroupTagSelectorStub,
    installAssetGroupTagStub,
    installAssetGroupTagsSearchStub,
    installAssetGroupTagsZoneDetailsStub,
} from 'bh-playwright-testing/stubs';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

test.describe('WCAG A/AA violations - Privilege Zones - Zones tab', () => {
    test.beforeEach('setup', async ({ page }) => {
        await page.goto('/ui/privilege-zones/zones/1/details');
        await page.getByRole('heading', { name: 'Zone Details' }).waitFor({ state: 'visible' });
    });

    test.describe('Zone details panel', () => {
        test('default state', async ({ page, makeAxeBuilder }, testInfo) => {
            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('search', async ({ page, makeAxeBuilder }, testInfo) => {
            await installAssetGroupTagsSearchStub(page);

            // Type at least 3 characters to trigger the debounced search and open the results popover.
            await page.getByTestId('privilege-zone-detail-search-bar').fill('ADMIN');

            // Wait for the filtered Objects to render before scanning.
            await page.getByText('PLAYWRIGHT_ADMIN_USER').waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('[data-radix-popper-content-wrapper]').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('search with no results', async ({ page, makeAxeBuilder }, testInfo) => {
            await installAssetGroupTagsSearchStub(page);

            // Type at query that will have no matches
            await page.getByTestId('privilege-zone-detail-search-bar').fill('XXXYYY');

            // Wait for the filtered Objects to render before scanning.
            await page.getByText('No results').first().waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('[data-radix-popper-content-wrapper]').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('expanded rules and objects', async ({ page, makeAxeBuilder }, testInfo) => {
            // Stub the Zone Details data so Rules (including Default Rules) and Objects
            // (Computers, Domains, Groups) render deterministically, then reload to apply the routes.
            await installAssetGroupTagsZoneDetailsStub(page);
            await page.reload();
            await page.getByRole('heading', { name: 'Zone Details' }).waitFor({ state: 'visible' });

            // Expand the Default Rules accordion and wait for a stubbed default rule to render.
            await page.getByTestId('privilege-zones_details_default_selectors-accordion_open-toggle-button').click();
            await page.getByText('PLAYWRIGHT_DEFAULT_RULE_1').first().waitFor({ state: 'visible' });

            // Expand the Computers objects accordion and wait for a stubbed computer to render.
            await page.getByTestId('privilege-zones_details_Computer-accordion_open-toggle-button').click();
            await page.getByText('PLAYWRIGHT_COMPUTER_1').first().waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });
    });

    test.describe('Side panels', () => {
        test('Rule side panel tab', async ({ page, makeAxeBuilder }, testInfo) => {
            // Stub the single tag + selector so the Rule tab renders deterministic rule details.
            await installAssetGroupTagStub(page);
            await installAssetGroupTagSelectorStub(page);

            // Navigating to the rule details route enables (but does not auto-select) the Rule tab; the
            // active tab is local state that defaults to the Zone tab, so click the Rule tab to open it.
            await page.goto('/ui/privilege-zones/zones/1/rules/3003/details');
            await page.getByRole('tab', { name: 'Rule' }).click();
            await page.getByTestId('privilege-zones_selector-details-card').waitFor({ state: 'visible' });
            await page.getByText('PLAYWRIGHT_CUSTOM_RULE_1').first().waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('Object side panel tab', async ({ page, makeAxeBuilder }, testInfo) => {
            // Stub the single tag + member (and its node properties) so the Object tab renders the
            // shared Entity Info panel deterministically.
            await installAssetGroupTagStub(page);
            await installAssetGroupTagMemberStub(page);

            // Navigating to the member details route enables (but does not auto-select) the Object tab;
            // the active tab is local state that defaults to the Zone tab, so click it to open it.
            await page.goto('/ui/privilege-zones/zones/1/objects/4001/details');
            await page.getByRole('tab', { name: 'Object' }).click();
            await page.getByTestId('explore_entity-information-panel').waitFor({ state: 'visible' });
            await page.getByText('PLAYWRIGHT_COMPUTER_1').first().waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });
    });
});

test.describe('WCAG A/AA violations - Privilege Zones - Save pages', () => {
    test('Edit Zone page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Stub the single tag so the Edit Zone form fields populate instead of showing a skeleton.
        await installAssetGroupTagStub(page);

        await page.goto('/ui/privilege-zones/zones/1/save');
        await page.getByTestId('privilege-zones_save_tag-form_name-input').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Create rule page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Create mode does not fetch a selector; only the tag info is needed for form context.
        await installAssetGroupTagStub(page);

        await page.goto('/ui/privilege-zones/zones/1/rules/save');
        await page.getByTestId('rule-form').waitFor({ state: 'visible' });
        await page.getByTestId('privilege-zones_save_rule-form_name-input').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Edit rule page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Edit mode fetches both the tag (for context) and the selector being edited.
        await installAssetGroupTagStub(page);
        await installAssetGroupTagSelectorStub(page);

        await page.goto('/ui/privilege-zones/zones/1/rules/3003/save');
        await page.getByTestId('rule-form').waitFor({ state: 'visible' });
        await page.getByTestId('privilege-zones_save_rule-form_name-input').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
