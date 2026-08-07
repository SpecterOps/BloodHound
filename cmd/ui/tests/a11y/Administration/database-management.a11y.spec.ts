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

test.describe('Administration - Database Management - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/graphs/source-kinds', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: {
                        kinds: [
                            { id: 1, name: 'Base' },
                            { id: 2, name: 'AZBase' },
                            { id: 3, name: 'ACustomBase' },
                            { id: 0, name: 'Sourceless' },
                        ],
                    },
                },
            });
        });
        await page.route('**/api/v2/features', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            id: 1,
                            key: 'clear_graph_data',
                            name: 'Clear Graph Data',
                            description: 'Enables the ability to delete all nodes and edges from the graph database.',
                            enabled: true,
                            user_updatable: true,
                        },
                    ],
                },
            });
        });
        await page.route('**/api/v2/self', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: {
                        sso_provider_id: null,
                        AuthSecret: null,
                        roles: [
                            {
                                name: 'Administrator',
                                permissions: [{ authority: 'db', name: 'Wipe' }],
                                id: 4,
                            },
                        ],
                        first_name: 'Test',
                        last_name: 'Administrator',
                        email_address: 'test-admin@example.com',
                        principal_name: 'test_admin',
                        is_disabled: false,
                        eula_accepted: true,
                        id: 'user-1',
                        created_at: '2026-01-01T12:00:00Z',
                        updated_at: '2026-01-01T12:00:00Z',
                        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
                    },
                },
            });
        });
    });

    test('page', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/administration/database-management');
        await page.getByRole('heading', { name: 'Database Management' }).waitFor({ state: 'visible' });
        await page.getByRole('checkbox', { name: 'All graph data' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('delete confirmation dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/administration/database-management');
        await page.getByRole('checkbox', { name: 'All asset group selectors' }).check();
        await page.getByRole('button', { name: 'Delete' }).click();
        await hideBySelector(page, '#content-wrapper');

        await page
            .getByRole('dialog', { name: 'Delete data from the current environment?' })
            .waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
