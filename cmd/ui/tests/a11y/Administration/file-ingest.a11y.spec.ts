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
const completedIngest = {
    created_at: '2026-08-01T12:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    end_time: '2026-08-01T12:01:00Z',
    failed_files: 0,
    id: 1,
    last_ingest: '2026-08-01T12:01:00Z',
    start_time: '2026-08-01T12:00:00Z',
    status: 2,
    status_message: 'Completed',
    total_files: 1,
    updated_at: '2026-08-01T12:01:00Z',
    user_email_address: 'analyst@example.com',
    user_id: 'user-1',
};

const failedIngest = {
    ...completedIngest,
    failed_files: 1,
    id: 2,
    status: 5,
    status_message: 'Failed',
};

test.describe('Administration - File Ingest - has no detectable WCAG A/AA violations', () => {
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
    });

    test('empty history', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload**', async (route) => {
            if (
                route.request().method() !== 'GET' ||
                new URL(route.request().url()).pathname !== '/api/v2/file-upload'
            ) {
                return route.fallback();
            }

            await route.fulfill({ json: { count: 0, data: [], limit: 10, skip: 0 } });
        });

        await page.goto('/ui/administration/file-ingest');
        await page.getByRole('columnheader', { name: 'ID / User / Status' }).waitFor({ state: 'visible' });
        await page.getByText('0–0 of 0').waitFor({ state: 'visible' });
        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with history', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload**', async (route) => {
            if (
                route.request().method() !== 'GET' ||
                new URL(route.request().url()).pathname !== '/api/v2/file-upload'
            ) {
                return route.fallback();
            }

            await route.fulfill({ json: { count: 1, data: [completedIngest], limit: 10, skip: 0 } });
        });

        await page.goto('/ui/administration/file-ingest');
        await page.getByRole('button', { name: 'View ingest 1 details' }).waitFor({ state: 'visible' });
        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
    test('with ingest selected', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload**', async (route) => {
            if (
                route.request().method() !== 'GET' ||
                new URL(route.request().url()).pathname !== '/api/v2/file-upload'
            ) {
                return route.fallback();
            }

            await route.fulfill({ json: { count: 1, data: [completedIngest], limit: 10, skip: 0 } });
        });
        await page.route('**/api/v2/file-upload/1/completed-tasks', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            created_at: '2026-08-01T12:01:00Z',
                            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
                            errors: [],
                            file_name: 'bloodhound-data.zip',
                            id: 1,
                            parent_file_name: '',
                            updated_at: '2026-08-01T12:01:00Z',
                            warnings: [],
                        },
                    ],
                },
            });
        });

        await page.goto('/ui/administration/file-ingest');
        await page.getByRole('button', { name: 'View ingest 1 details' }).click();
        await page.getByText('bloodhound-data.zip').waitFor({ state: 'visible' });
        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with errored ingest selected', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload**', async (route) => {
            if (
                route.request().method() !== 'GET' ||
                new URL(route.request().url()).pathname !== '/api/v2/file-upload'
            ) {
                return route.fallback();
            }

            await route.fulfill({ json: { count: 1, data: [failedIngest], limit: 10, skip: 0 } });
        });
        await page.route('**/api/v2/file-upload/2/completed-tasks', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            created_at: '2026-08-01T12:01:00Z',
                            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
                            errors: ['The uploaded file could not be parsed.'],
                            file_name: 'invalid-data.json',
                            id: 2,
                            parent_file_name: '',
                            updated_at: '2026-08-01T12:01:00Z',
                            warnings: [],
                        },
                    ],
                },
            });
        });

        await page.goto('/ui/administration/file-ingest');
        await page.getByRole('button', { name: 'View ingest 2 details' }).click();
        await page.getByRole('button', { name: /invalid-data\.json Failure/ }).click();
        await page.getByText('The uploaded file could not be parsed.').waitFor({ state: 'visible' });
        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('ingest filter', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload**', async (route) => {
            if (
                route.request().method() !== 'GET' ||
                new URL(route.request().url()).pathname !== '/api/v2/file-upload'
            ) {
                return route.fallback();
            }

            await route.fulfill({ json: { count: 0, data: [], limit: 10, skip: 0 } });
        });
        await page.route('**/api/v2/bloodhound-users-minimal', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: { users: [] } } });
        });

        await page.goto('/ui/administration/file-ingest');
        await page.getByTestId('file_ingest_log-open_filter_dialog').click();

        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('heading', { name: 'Filter' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
