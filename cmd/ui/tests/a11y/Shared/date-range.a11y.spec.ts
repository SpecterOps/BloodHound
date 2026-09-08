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

import { hideBySelector } from 'bh-playwright-testing/axe';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

test.describe('Date Range', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/features', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [{ key: 'open_graph_phase_2', enabled: true }],
                },
            });
        });
        await page.route('**/api/v2/file-upload**', async (route) => {
            if (
                route.request().method() !== 'GET' ||
                new URL(route.request().url()).pathname !== '/api/v2/file-upload'
            ) {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    count: 0,
                    data: [],
                    limit: 10,
                    skip: 0,
                },
            });
        });
        await page.route('**/api/v2/bloodhound-users-minimal', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: {
                        users: [],
                    },
                },
            });
        });

        await page.goto('/ui/administration/file-ingest');

        await page
            .getByRole('columnheader', {
                name: 'ID / User / Status',
                exact: true,
            })
            .waitFor({ state: 'visible' });
        await page.getByText('0\u20130 of 0', { exact: true }).waitFor({ state: 'visible' });

        const filterButton = page.getByRole('button', {
            name: 'Open file ingest filters',
            exact: true,
        });

        await filterButton.waitFor({ state: 'visible' });
        await filterButton.click();

        await hideBySelector(page, '#content-wrapper');

        const filterDialog = page.getByRole('dialog');

        await filterDialog.waitFor({ state: 'visible' });
        await filterDialog
            .getByRole('heading', {
                name: 'Filter Clear All',
                exact: true,
            })
            .waitFor({ state: 'visible' });
    });

    test('default state', async ({ page, makeAxeBuilder }, testInfo) => {
        const filterDialog = page.getByRole('dialog');
        const startDate = filterDialog.getByRole('textbox', {
            name: 'Start Date',
            exact: true,
        });
        const endDate = filterDialog.getByRole('textbox', {
            name: 'End Date',
            exact: true,
        });

        await filterDialog.getByText('Date Range', { exact: true }).waitFor({ state: 'visible' });
        await startDate.waitFor({ state: 'visible' });
        await endDate.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with focus', async ({ page, makeAxeBuilder }, testInfo) => {
        const filterDialog = page.getByRole('dialog');
        const statusSelect = filterDialog.getByRole('combobox', {
            name: 'Status Select',
            exact: true,
        });
        const startDate = filterDialog.getByRole('textbox', {
            name: 'Start Date',
            exact: true,
        });

        await statusSelect.focus();
        await page.keyboard.press('Tab');

        await startDate.and(page.locator(':focus')).waitFor({ state: 'visible' });
        await filterDialog.getByPlaceholder('yyyy-mm-dd', { exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('calendar visible', async ({ page, makeAxeBuilder }, testInfo) => {
        const filterDialog = page.getByRole('dialog');
        const startDateCalendarButton = filterDialog
            .getByRole('button', {
                name: 'Choose Date',
                exact: true,
            })
            .first();

        await startDateCalendarButton.click();

        const calendarGrid = page.getByRole('grid');
        const calendarDialog = page.getByRole('dialog').filter({
            has: calendarGrid,
        });
        const visibleMonth = new Intl.DateTimeFormat('en-US', {
            month: 'long',
            year: 'numeric',
        }).format(new Date());

        await calendarDialog.waitFor({ state: 'visible' });
        await calendarGrid.waitFor({ state: 'visible' });
        await calendarGrid.getByRole('gridcell').first().waitFor({ state: 'visible' });
        await calendarDialog.getByText(visibleMonth, { exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
