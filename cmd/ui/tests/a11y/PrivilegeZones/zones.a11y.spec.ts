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

import { Locator, Page } from '@playwright/test';
import {
    installAssetGroupTagMemberStub,
    installAssetGroupTagSelectorStub,
    installAssetGroupTagStub,
    installAssetGroupTagsSearchStub,
    installAssetGroupTagsZoneDetailsStub,
} from 'bh-playwright-testing/stubs';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

const ZONES_URL = '/ui/privilege-zones/zones/1/details';
const EDIT_ZONE_URL = '/ui/privilege-zones/zones/1/save';
const CREATE_RULE_URL = '/ui/privilege-zones/zones/1/rules/save';
const EDIT_RULE_URL = '/ui/privilege-zones/zones/1/rules/3003/save';

/** Visit the URL, collapse the nav menu (for more space), and wait for locator to be visible */
const openPage = async (page: Page, url: string, waitFor?: Locator) => {
    const waitLocator = waitFor ?? page.getByRole('heading', { name: 'Zone Details' });

    await page.goto(url);
    await page.getByRole('button', { name: 'Toggle Navigation' }).click();
    await waitLocator.waitFor({ state: 'visible' });
};

test.describe('WCAG A/AA violations - Privilege Zones - Zones tab', () => {
    test.describe('Zone details panel', () => {
        test('default state', async ({ page, makeAxeBuilder }, testInfo) => {
            await openPage(page, ZONES_URL);

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('search', async ({ page, makeAxeBuilder }, testInfo) => {
            await installAssetGroupTagsSearchStub(page);
            await openPage(page, ZONES_URL);

            // Type at least 3 characters to trigger the debounced search and open the results popover.
            await page.getByTestId('privilege-zone-detail-search-bar').fill('ADMIN');

            // Wait for the filtered Objects to render before scanning.
            await page.getByText('PLAYWRIGHT_ADMIN_USER').waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('[data-radix-popper-content-wrapper]').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('search with no results', async ({ page, makeAxeBuilder }, testInfo) => {
            await installAssetGroupTagsSearchStub(page);
            await openPage(page, ZONES_URL);

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
            await openPage(page, ZONES_URL);

            // Expand the Computers objects accordion and wait for a stubbed computer to render.
            await page.getByTestId('privilege-zones_details_Computer-accordion_open-toggle-button').click();
            await page.getByText('PLAYWRIGHT_COMPUTER_1').waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });
    });

    test.describe('Side panels', () => {
        test('Rule side panel tab', async ({ page, makeAxeBuilder }, testInfo) => {
            // Stub the single tag + selector so the Rule tab renders deterministic rule details.
            await installAssetGroupTagStub(page);
            await installAssetGroupTagSelectorStub(page);
            await openPage(page, ZONES_URL);

            // Selecting the first Default Rule updates the route and switches the details panel to Rule.
            await page.getByTestId('privilege-zones_details_default_selectors-accordion_open-toggle-button').click();
            await page.getByRole('button', { name: 'Account Operators' }).click();
            await page.getByText('Playwright stubbed rule').waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });

        test('Object side panel tab', async ({ page, makeAxeBuilder }, testInfo) => {
            // Stub Zone Details with objects grouped by node type, plus the first object's detail response.
            await installAssetGroupTagsZoneDetailsStub(page);
            await installAssetGroupTagMemberStub(page);
            await openPage(page, ZONES_URL);

            // Open the first node type and select its first object to switch the details panel to Object.
            await page.getByTestId('privilege-zones_details_Computer-accordion_open-toggle-button').click();

            const computer = page.getByRole('button', { name: 'PLAYWRIGHT_COMPUTER_1' });
            await computer.waitFor({ state: 'visible' });
            await computer.click();

            // Once the selected computer's node data resolves, the Object tab renders its details.
            await page.getByText('Node Type:').waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();
            await expectNoAccessibilityViolations(testInfo, results, { page });
        });
    });
});

test.describe('WCAG A/AA violations - Privilege Zones - Save pages', () => {
    test('Edit Zone page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Stub the single tag so the Edit Zone form fields populate instead of showing a skeleton.
        await installAssetGroupTagStub(page);
        await openPage(page, EDIT_ZONE_URL, page.getByTestId('privilege-zones_save_tag-form_name-input'));

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Create rule page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Create mode does not fetch a selector; only the tag info is needed for form context.
        await installAssetGroupTagStub(page);
        await openPage(page, CREATE_RULE_URL, page.getByTestId('privilege-zones_save_rule-form_name-input'));

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Edit rule page', async ({ page, makeAxeBuilder }, testInfo) => {
        // Edit mode fetches both the tag (for context) and the selector being edited.
        await installAssetGroupTagStub(page);
        await installAssetGroupTagSelectorStub(page);
        await openPage(page, EDIT_RULE_URL, page.getByTestId('privilege-zones_save_rule-form_name-input'));

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
