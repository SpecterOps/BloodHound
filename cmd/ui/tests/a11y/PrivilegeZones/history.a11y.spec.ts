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
import { installAssetGroupTagsHistoryStub } from 'bh-playwright-testing/stubs';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

const HISTORY_URL = '/ui/privilege-zones/history';

/** Visit the URL, collapse the nav menu (for more space), and wait for locator to be visible */
const openPage = async (page: Page, url: string, waitFor?: Locator) => {
    const waitLocator = waitFor ?? page.getByRole('heading', { name: 'History Log' });

    await page.goto(url);
    await page.getByRole('button', { name: 'Toggle Navigation' }).click();
    await waitLocator.waitFor({ state: 'visible' });
};

test.describe('WCAG A/AA violations - Privilege Zones - History tab', () => {
    test('empty history', async ({ page, makeAxeBuilder }, testInfo) => {
        // Return no history records so the History Log renders its empty state.
        await installAssetGroupTagsHistoryStub(page, { data: { records: [] } });
        await openPage(page, HISTORY_URL);

        // Wait for the DataTable's empty fallback to render before scanning.
        await page.getByText('No results.').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with history', async ({ page, makeAxeBuilder }, testInfo) => {
        // Default stub data renders a populated History Log table.
        await installAssetGroupTagsHistoryStub(page);
        await openPage(page, HISTORY_URL);

        // Wait for a stubbed record's translated action to render before scanning.
        await page.getByText('Create Tag').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With open note', async ({ page, makeAxeBuilder }, testInfo) => {
        // Default stub data includes a record with a note so the Note panel can be opened.
        await installAssetGroupTagsHistoryStub(page);
        await openPage(page, HISTORY_URL);

        // Wait for the populated table, then open the first record's note into the side panel.
        await page.getByText('Create Tag').waitFor({ state: 'visible' });
        await page.getByRole('button', { name: 'Show note' }).first().click();

        // Wait for the note contents to render in the side panel before scanning.
        await page.getByText('Created the Playwright zone for accessibility testing.').waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Filter dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        // Default stub data provides Zone/Label options for the filter dialog.
        await installAssetGroupTagsHistoryStub(page);
        await openPage(page, HISTORY_URL);

        // Open the filter dialog and wait for its content to render before scanning.
        await page.getByTestId('privilege-zones_history_filter-button').click();
        await page.getByRole('button', { name: 'Clear All' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Filter dialog with selections', async ({ page, makeAxeBuilder }, testInfo) => {
        // Default stub data provides Zone/Label options for the filter dialog.
        await installAssetGroupTagsHistoryStub(page);
        await openPage(page, HISTORY_URL);

        // Open the filter dialog and wait for its content to render before scanning.
        await page.getByTestId('privilege-zones_history_filter-button').click();
        const dialog = page.getByRole('dialog');
        await page.getByRole('button', { name: 'Clear All' }).waitFor({ state: 'visible' });

        // Select an Action (static option list, no network dependency).
        const actionSelect = dialog.getByRole('combobox', { name: 'Action' });
        await actionSelect.click();
        await page.getByRole('option', { name: 'Create Tag' }).click();

        // Select a Zone/Label (options provided by the stubbed tag list).
        const tagSelect = dialog.getByRole('combobox', { name: 'Zone/Label' });
        await tagSelect.click();
        await page.getByRole('option', { name: 'PLAYWRIGHT_ZONE' }).click();

        // Select a Made By (the BloodHound system option always renders regardless of the users list).
        const madeBySelect = dialog.getByRole('combobox', { name: 'Made By' });
        await madeBySelect.click();
        await page.getByRole('option', { name: 'BloodHound' }).click();

        // Enter a Start Date into the masked date input (local form state, no network dependency).
        const startDate = dialog.getByLabel('Start Date');
        await startDate.pressSequentially('2024-01-01');

        // Wait for every selection to reflect in its control before scanning.
        await actionSelect.getByText('Create Tag').waitFor({ state: 'visible' });
        await tagSelect.getByText('PLAYWRIGHT_ZONE').waitFor({ state: 'visible' });
        await madeBySelect.getByText('BloodHound').waitFor({ state: 'visible' });
        await startDate.and(page.locator('[value="2024-01-01"]')).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
