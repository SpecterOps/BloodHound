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
import { Page } from '@playwright/test';
import { expect, test } from 'bh-playwright-testing';

const installPathfindingStub = async (page: Page) => {
    await page.route('**/api/v2/graphs/shortest-path**', async (route) => {
        if (route.request().method() !== 'GET') {
            return route.fallback();
        }

        await route.fulfill({
            json: {
                data: {
                    nodes: {},
                    edges: [],
                },
            },
        });
    });
};

test.describe('WCAG A/AA Violations - Explore - Pathfinding Tab', () => {
    test.beforeEach(async ({ page, goAndWaitFor }) => {
        await goAndWaitFor(
            '/ui/explore?exploreSearchTab=pathfinding',
            page.getByRole('textbox', { name: 'Start node' })
        );
    });

    test('Pathfinding tab', async ({ page, checkA11y }) => {
        await page.getByText('Begin typing to search.').first().waitFor();
        await checkA11y();
    });

    test('Start node with results', async ({ page, checkA11y }) => {
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

        await page.getByRole('textbox', { name: 'Start Node' }).fill(searchTerm);
        await page.getByText('TEST RESULT').waitFor();

        await checkA11y();
    });

    test('Start node with no results', async ({ page, checkA11y }) => {
        const searchTerm = 'zzzznonexistentnode9999';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });

        const startNodeField = page.getByRole('textbox', { name: 'Start Node' });
        await startNodeField.fill(searchTerm);

        const noResultsMessage = `No results found for "${searchTerm}"`;
        await expect(page.getByText(noResultsMessage)).toBeVisible();

        await checkA11y();
    });

    test('Destination node', async ({ page, checkA11y }) => {
        // Pathfinding autofocus opens the Start node popup over the Destination Node field.
        await page.getByRole('textbox', { name: 'Start Node' }).press('Escape');
        await page.getByRole('textbox', { name: 'Destination Node' }).click();

        await checkA11y();
    });

    test('Destination node with results', async ({ page, checkA11y }) => {
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

        // Pathfinding autofocus opens the Start node popup over the Destination Node field.
        await page.getByRole('textbox', { name: 'Start Node' }).press('Escape');
        await page.getByRole('textbox', { name: 'Destination Node' }).fill(searchTerm);
        await page.getByText('DESTINATION TEST RESULT').waitFor();

        await checkA11y();
    });

    test('Destination node with no results', async ({ page, checkA11y }) => {
        const searchTerm = 'zzzznonexistentdestination9999';

        await page.route('**/api/v2/search**', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: [] } });
        });

        // Pathfinding autofocus opens the Start node popup over the Destination Node field.
        await page.getByRole('textbox', { name: 'Start Node' }).press('Escape');
        await page.getByRole('textbox', { name: 'Destination Node' }).fill(searchTerm);
        await page.getByText('No results found for "').waitFor();

        await checkA11y();
    });

    test('Enabled pathfinding controls', async ({ page, checkA11y }) => {
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

        await installPathfindingStub(page);

        await page.getByRole('textbox', { name: 'Start Node' }).fill(startResultName);
        await page.getByRole('option').filter({ hasText: startResultName }).click();

        await page.getByRole('textbox', { name: 'Destination Node' }).fill(destinationResultName);
        await page.getByRole('option').filter({ hasText: destinationResultName }).click();

        await page.getByRole('button', { name: 'Swap start and destination' }).waitFor();
        await page.getByRole('button', { name: 'Show pathfinding filter options' }).waitFor();

        await checkA11y();
    });

    test('Path edge filtering dialog', async ({ page, checkA11y }) => {
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

        await installPathfindingStub(page);

        await page.getByRole('textbox', { name: 'Start Node' }).fill(startResultName);
        await page.getByRole('option').filter({ hasText: startResultName }).click();
        await page.getByRole('textbox', { name: 'Destination Node' }).fill(destinationResultName);

        await page.getByRole('option').filter({ hasText: destinationResultName }).click();
        await page.getByRole('button', { name: 'Show pathfinding filter options' }).click();
        await page.getByRole('dialog', { name: 'Path Edge Filtering' }).waitFor();

        await checkA11y({ include: '[role=dialog]' });
    });

    test('Path edge filtering dialog with no selections', async ({ page, checkA11y }) => {
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

        await installPathfindingStub(page);

        await page.getByRole('textbox', { name: 'Start Node' }).fill(startResultName);
        await page.getByRole('option').filter({ hasText: startResultName }).click();

        await page.getByRole('textbox', { name: 'Destination Node' }).fill(destinationResultName);
        await page.getByRole('option').filter({ hasText: destinationResultName }).click();
        await page.getByRole('button', { name: 'Show pathfinding filter options' }).click();

        const dialog = page.getByRole('dialog', { name: 'Path Edge Filtering' });
        await dialog.waitFor();

        await dialog.getByRole('checkbox', { name: 'Active Directory', exact: true }).click();
        await dialog.getByRole('checkbox', { name: 'Azure', exact: true }).click();

        await expect(dialog.getByRole('checkbox', { checked: true })).toHaveCount(0);

        await checkA11y({ include: '[role=dialog]' });
    });

    test('Path edge filtering dialog with search results', async ({ page, checkA11y }) => {
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

        await installPathfindingStub(page);

        await page.getByRole('textbox', { name: 'Start Node' }).fill(startResultName);
        await page.getByRole('option').filter({ hasText: startResultName }).click();
        await page.getByRole('textbox', { name: 'Destination Node' }).fill(destinationResultName);

        await page.getByRole('option').filter({ hasText: destinationResultName }).click();
        await page.getByRole('button', { name: 'Show pathfinding filter options' }).click();

        const dialog = page.getByRole('dialog', { name: 'Path Edge Filtering' });
        await dialog.waitFor();

        const searchTextbox = dialog.getByRole('textbox', { name: 'Search edges...' });
        await searchTextbox.fill('write');
        await expect(searchTextbox).toHaveValue('write');

        await checkA11y({ include: '[role=dialog]' });
    });

    test('Path edge filtering dialog with no search results', async ({ page, checkA11y }) => {
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

        await installPathfindingStub(page);

        await page.getByRole('textbox', { name: 'Start Node' }).fill(startResultName);
        await page.getByRole('option').filter({ hasText: startResultName }).click();
        await page.getByRole('textbox', { name: 'Destination Node' }).fill(destinationResultName);

        await page.getByRole('option').filter({ hasText: destinationResultName }).click();
        await page.getByRole('button', { name: 'Show pathfinding filter options' }).click();

        const dialog = page.getByRole('dialog', { name: 'Path Edge Filtering' });
        await dialog.waitFor();

        await dialog.getByRole('textbox', { name: 'Search edges...' }).fill(searchTerm);
        await expect(dialog.getByRole('checkbox')).toHaveCount(0);

        await checkA11y({ include: '[role=dialog]' });
    });
});
