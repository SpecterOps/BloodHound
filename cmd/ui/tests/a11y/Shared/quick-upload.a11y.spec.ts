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

test.describe('Quick Upload dialog', () => {
    const UPLOAD_FILE = 'tests/a11y/Shared/fixtures/playwright-upload.json';

    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/file-upload/accepted-types', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: ['application/json'],
                },
            });
        });

        await page.goto('/ui/administration/data-quality');

        await page.getByRole('heading', { name: 'Data Quality', exact: true }).waitFor({ state: 'visible' });

        const quickUploadButton = page.getByRole('button', {
            name: 'Quick Upload',
            exact: true,
        });

        await quickUploadButton.waitFor({ state: 'visible' });
        await quickUploadButton.click();

        await page.getByRole('dialog').waitFor({ state: 'visible' });
        await hideBySelector(page, '#content-wrapper');
    });

    test('default state', async ({ page, makeAxeBuilder }, testInfo) => {
        const dialog = page.getByRole('dialog');

        await dialog
            .getByText('Click here or drag and drop to upload JSON or zip/compressed JSON files', {
                exact: true,
            })
            .waitFor({ state: 'visible' });
        await dialog.getByText('View File Ingest History', { exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with files added', async ({ page, makeAxeBuilder }, testInfo) => {
        const dialog = page.getByRole('dialog');

        const fileChooserPromise = page.waitForEvent('filechooser');

        await dialog.locator('div[role="button"].size-full').click();

        const fileChooser = await fileChooserPromise;
        await fileChooser.setFiles(UPLOAD_FILE);
        await dialog.getByText('playwright-upload.json', { exact: true }).waitFor({ state: 'visible' });
        await dialog.getByRole('button', { name: 'Remove item', exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with completed uploads', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload/start', async (route) => {
            if (route.request().method() !== 'POST') {
                return route.fallback();
            }

            await route.fulfill({
                status: 201,
                json: {
                    data: {
                        id: 1,
                    },
                },
            });
        });
        await page.route('**/api/v2/file-upload/1', async (route) => {
            if (route.request().method() !== 'POST') {
                return route.fallback();
            }

            await route.fulfill({
                status: 202,
                json: {
                    data: '',
                },
            });
        });
        await page.route('**/api/v2/file-upload/1/end', async (route) => {
            if (route.request().method() !== 'POST') {
                return route.fallback();
            }

            await route.fulfill({
                status: 200,
                json: {
                    data: '',
                },
            });
        });

        const dialog = page.getByRole('dialog');

        const fileChooserPromise = page.waitForEvent('filechooser');

        await dialog.locator('div[role="button"].size-full').click();

        const fileChooser = await fileChooserPromise;
        await fileChooser.setFiles(UPLOAD_FILE);
        await dialog.getByText('playwright-upload.json', { exact: true }).waitFor({ state: 'visible' });
        await dialog.getByRole('button', { name: 'Upload', exact: true }).click();

        await dialog
            .getByText('All files have successfully been uploaded for ingest.', { exact: true })
            .waitFor({ state: 'visible' });
        await dialog.getByText('100%', { exact: true }).waitFor({ state: 'visible' });
        await dialog.getByText('Upload in progress.', { exact: true }).waitFor({ state: 'hidden' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('with failed uploads', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.route('**/api/v2/file-upload/start', async (route) => {
            if (route.request().method() !== 'POST') {
                return route.fallback();
            }

            await route.fulfill({
                status: 201,
                json: {
                    data: {
                        id: 1,
                    },
                },
            });
        });
        await page.route('**/api/v2/file-upload/1', async (route) => {
            if (route.request().method() !== 'POST') {
                return route.fallback();
            }

            await route.fulfill({
                status: 400,
                json: {
                    errors: [
                        {
                            message: 'Playwright upload failure',
                        },
                    ],
                },
            });
        });
        await page.route('**/api/v2/file-upload/1/end', async (route) => {
            if (route.request().method() !== 'POST') {
                return route.fallback();
            }

            await route.fulfill({
                status: 200,
                json: {
                    data: '',
                },
            });
        });

        const dialog = page.getByRole('dialog');

        const fileChooserPromise = page.waitForEvent('filechooser');

        await dialog.locator('div[role="button"].size-full').click();

        const fileChooser = await fileChooserPromise;
        await fileChooser.setFiles(UPLOAD_FILE);
        await dialog.getByText('playwright-upload.json', { exact: true }).waitFor({ state: 'visible' });
        await dialog.getByRole('button', { name: 'Upload', exact: true }).click();

        await dialog
            .getByText('Some files have failed to upload and have not been included for ingest.', {
                exact: true,
            })
            .waitFor({ state: 'visible' });
        await dialog.getByText('Failed to Upload', { exact: true }).waitFor({ state: 'visible' });
        await dialog.getByRole('button', { name: 'Retry upload', exact: true }).waitFor({ state: 'visible' });
        await dialog.getByText('Upload in progress.', { exact: true }).waitFor({ state: 'hidden' });

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
