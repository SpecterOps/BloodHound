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

test.describe('Administration - BloodHound Configuration - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v2/datapipe/status', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({ json: { data: { status: 'idle' } } });
        });
        await page.route('**/api/v2/config', async (route) => {
            if (route.request().method() !== 'GET') {
                return route.fallback();
            }

            await route.fulfill({
                json: {
                    data: [
                        {
                            key: 'analysis.citrix_rdp_support',
                            value: { enabled: false },
                        },
                    ],
                },
            });
        });
    });

    test('page', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/administration/bloodhound-configuration');
        await expect(page.getByRole('heading', { name: 'BloodHound Configuration' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Run Analysis Now' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Citrix RDP Support' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Analyze Now' })).toBeEnabled();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Analyze Now dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/administration/bloodhound-configuration');
        await page.getByRole('button', { name: 'Analyze Now' }).click();
        await expect(page.getByRole('dialog', { name: 'Confirm re-run analysis' })).toBeVisible();
        await expect(page.getByText(/Analysis may take some time/)).toBeVisible();
        await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Confirm' })).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
