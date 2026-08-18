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

test.describe('WCAG A/AA Violations - Explore - Pathfinding Tab', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/ui/explore?exploreSearchTab=pathfinding');
        await page.getByRole('textbox', { name: 'Start Node' }).waitFor({ state: 'visible' });
    });

    test('Pathfinding tab', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByText('Begin typing to search.').first().waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Start Node with results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchTerm = 'test';
        const searchResultName = 'TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: searchResultName,
                            objectid: 'playwright-pathfinding-result',
                            type: 'User',
                        },
                    ],
                },
            });
        });

        await page.getByLabel('Start Node').fill(searchTerm);
        await page.getByRole('option', { name: 'No results found for "' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Start Node with no results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchTerm = 'zzzznonexistentnode9999';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });

        const startNodeField = page.getByLabel('Start Node');
        await startNodeField.fill(searchTerm);

        const noResultsMessage = `No results found for "${searchTerm}"`;
        await expect(page.getByText(noResultsMessage)).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Destination Node', async ({ page, makeAxeBuilder }, testInfo) => {
        // Pathfinding autofocus opens the Start Node popup over the Destination Node field.
        await page.getByLabel('Start Node').press('Escape');

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.click();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Destination Node with results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchTerm = 'test';
        const searchResultName = 'DESTINATION TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: searchResultName,
                            objectid: 'playwright-pathfinding-destination-result',
                            type: 'User',
                        },
                    ],
                },
            });
        });

        // Pathfinding autofocus opens the Start Node popup over the Destination Node field.
        await page.getByLabel('Start Node').press('Escape');

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.click();
        await destinationNodeField.fill(searchTerm);

        const searchResult = page.getByRole('option').filter({ hasText: searchResultName });
        await expect(searchResult).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Destination Node with no results', async ({ page, makeAxeBuilder }, testInfo) => {
        const searchTerm = 'zzzznonexistentdestination9999';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });

        // Pathfinding autofocus opens the Start Node popup over the Destination Node field.
        await page.getByLabel('Start Node').press('Escape');

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.click();
        await destinationNodeField.fill(searchTerm);

        const noResultsMessage = `No results found for "${searchTerm}"`;
        await expect(page.getByText(noResultsMessage)).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Enabled pathfinding controls', async ({ page, makeAxeBuilder }, testInfo) => {
        const startResultName = 'START TEST RESULT';
        const destinationResultName = 'DESTINATION TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: startResultName,
                            objectid: 'playwright-pathfinding-start-result',
                            type: 'User',
                        },
                        {
                            name: destinationResultName,
                            objectid: 'playwright-pathfinding-destination-result',
                            type: 'Computer',
                        },
                    ],
                },
            });
        });

        const startNodeField = page.getByLabel('Start Node');
        await startNodeField.fill(startResultName);

        const startResult = page.getByRole('option').filter({ hasText: startResultName });
        await expect(startResult).toBeVisible();
        await startResult.click();

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.click();
        await destinationNodeField.fill(destinationResultName);

        const destinationResult = page.getByRole('option').filter({ hasText: destinationResultName });
        await expect(destinationResult).toBeVisible();
        await destinationResult.click();

        const swapButton = page.getByRole('button', { name: 'Swap start and destination' });
        const filterButton = page.getByRole('button', {
            name: 'Show pathfinding filter options',
        });

        await expect(swapButton).toBeEnabled();
        await expect(filterButton).toBeEnabled();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Path Edge Filtering dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const startResultName = 'START TEST RESULT';
        const destinationResultName = 'DESTINATION TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: startResultName,
                            objectid: 'playwright-pathfinding-start-result',
                            type: 'User',
                        },
                        {
                            name: destinationResultName,
                            objectid: 'playwright-pathfinding-destination-result',
                            type: 'Computer',
                        },
                    ],
                },
            });
        });

        const startNodeField = page.getByLabel('Start Node');
        await startNodeField.fill(startResultName);

        const startResult = page.getByRole('option').filter({ hasText: startResultName });
        await expect(startResult).toBeVisible();
        await startResult.click();

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.click();
        await destinationNodeField.fill(destinationResultName);

        const destinationResult = page.getByRole('option').filter({ hasText: destinationResultName });
        await expect(destinationResult).toBeVisible();
        await destinationResult.click();

        const filterButton = page.getByRole('button', {
            name: 'Show pathfinding filter options',
        });
        await expect(filterButton).toBeEnabled();
        await filterButton.click();

        const dialog = page.getByRole('dialog', { name: 'Path Edge Filtering' });
        await expect(dialog).toBeVisible();
        await expect(dialog.getByRole('checkbox', { name: 'Active Directory', exact: true })).toBeChecked();
        await expect(dialog.getByRole('checkbox', { name: 'Azure', exact: true })).toBeChecked();

        const results = await makeAxeBuilder().include('[role=dialog]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Path Edge Filtering dialog with no selections', async ({ page, makeAxeBuilder }, testInfo) => {
        const startResultName = 'START TEST RESULT';
        const destinationResultName = 'DESTINATION TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: startResultName,
                            objectid: 'playwright-pathfinding-start-result',
                            type: 'User',
                        },
                        {
                            name: destinationResultName,
                            objectid: 'playwright-pathfinding-destination-result',
                            type: 'Computer',
                        },
                    ],
                },
            });
        });

        const startNodeField = page.getByLabel('Start Node');
        await startNodeField.fill(startResultName);

        const startResult = page.getByRole('option').filter({ hasText: startResultName });
        await expect(startResult).toBeVisible();
        await startResult.click();

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.fill(destinationResultName);

        const destinationResult = page.getByRole('option').filter({ hasText: destinationResultName });
        await expect(destinationResult).toBeVisible();
        await destinationResult.click();

        const filterButton = page.getByRole('button', {
            name: 'Show pathfinding filter options',
        });
        await expect(filterButton).toBeEnabled();
        await filterButton.click();

        const dialog = page.getByRole('dialog', {
            name: 'Path Edge Filtering',
        });
        await expect(dialog).toBeVisible();

        const activeDirectoryFilter = dialog.getByRole('checkbox', {
            name: 'Active Directory',
            exact: true,
        });
        const azureFilter = dialog.getByRole('checkbox', {
            name: 'Azure',
            exact: true,
        });

        await expect(activeDirectoryFilter).toBeChecked();
        await expect(azureFilter).toBeChecked();

        await activeDirectoryFilter.click();
        await azureFilter.click();

        await expect(activeDirectoryFilter).not.toBeChecked();
        await expect(azureFilter).not.toBeChecked();
        await expect(dialog.getByRole('checkbox', { checked: true })).toHaveCount(0);

        const results = await makeAxeBuilder().include('[role=dialog]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Path Edge Filtering dialog with search results', async ({ page, makeAxeBuilder }, testInfo) => {
        const startResultName = 'START TEST RESULT';
        const destinationResultName = 'DESTINATION TEST RESULT';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: startResultName,
                            objectid: 'playwright-pathfinding-start-result',
                            type: 'User',
                        },
                        {
                            name: destinationResultName,
                            objectid: 'playwright-pathfinding-destination-result',
                            type: 'Computer',
                        },
                    ],
                },
            });
        });

        const startNodeField = page.getByLabel('Start Node');
        await startNodeField.fill(startResultName);

        const startResult = page.getByRole('option').filter({ hasText: startResultName });
        await expect(startResult).toBeVisible();
        await startResult.click();

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.fill(destinationResultName);

        const destinationResult = page.getByRole('option').filter({ hasText: destinationResultName });
        await expect(destinationResult).toBeVisible();
        await destinationResult.click();

        const filterButton = page.getByRole('button', {
            name: 'Show pathfinding filter options',
        });
        await expect(filterButton).toBeEnabled();
        await filterButton.click();

        const dialog = page.getByRole('dialog', {
            name: 'Path Edge Filtering',
        });
        await expect(dialog).toBeVisible();

        const searchTextbox = dialog.getByRole('textbox', { name: 'Search edges...' });
        await searchTextbox.fill('write');
        await expect(searchTextbox).toHaveValue('write');

        await expect(dialog.getByRole('checkbox', { name: 'GenericWrite', exact: true })).toBeVisible();
        await expect(dialog.getByRole('checkbox', { name: 'WriteOwner', exact: true })).toBeVisible();
        await expect(dialog.getByRole('checkbox', { name: 'Credential Access', exact: true })).toHaveCount(0);

        const results = await makeAxeBuilder().include('[role=dialog]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Path Edge Filtering dialog with no search results', async ({ page, makeAxeBuilder }, testInfo) => {
        const startResultName = 'START TEST RESULT';
        const destinationResultName = 'DESTINATION TEST RESULT';
        const searchTerm = 'no-match-test-value';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            name: startResultName,
                            objectid: 'playwright-pathfinding-start-result',
                            type: 'User',
                        },
                        {
                            name: destinationResultName,
                            objectid: 'playwright-pathfinding-destination-result',
                            type: 'Computer',
                        },
                    ],
                },
            });
        });

        const startNodeField = page.getByLabel('Start Node');
        await startNodeField.fill(startResultName);

        const startResult = page.getByRole('option').filter({ hasText: startResultName });
        await expect(startResult).toBeVisible();
        await startResult.click();

        const destinationNodeField = page.getByLabel('Destination Node');
        await destinationNodeField.fill(destinationResultName);

        const destinationResult = page.getByRole('option').filter({ hasText: destinationResultName });
        await expect(destinationResult).toBeVisible();
        await destinationResult.click();

        const filterButton = page.getByRole('button', {
            name: 'Show pathfinding filter options',
        });
        await expect(filterButton).toBeEnabled();
        await filterButton.click();

        const dialog = page.getByRole('dialog', {
            name: 'Path Edge Filtering',
        });
        await expect(dialog).toBeVisible();

        const searchTextbox = dialog.getByRole('textbox', { name: 'Search edges...' });
        await searchTextbox.fill(searchTerm);
        await expect(searchTextbox).toHaveValue(searchTerm);

        await expect(dialog.getByRole('checkbox', { name: 'GenericWrite', exact: true })).toHaveCount(0);
        await expect(dialog.getByRole('checkbox', { name: 'Active Directory', exact: true })).toHaveCount(0);
        await expect(dialog.getByRole('checkbox')).toHaveCount(0);

        const results = await makeAxeBuilder().include('[role=dialog]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
