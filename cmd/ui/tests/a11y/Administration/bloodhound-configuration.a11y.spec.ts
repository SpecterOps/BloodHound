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

        await page.getByRole('heading', { name: 'BloodHound Configuration' }).waitFor({ state: 'visible' });
        await page.getByRole('heading', { name: 'Run Analysis Now' }).waitFor({ state: 'visible' });
        await page.getByRole('heading', { name: 'Citrix RDP Support' }).waitFor({ state: 'visible' });
        await page.getByRole('button', { name: 'Analyze Now' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Analyze Now dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/administration/bloodhound-configuration');
        await page.getByRole('button', { name: 'Analyze Now' }).click();

        await page.getByRole('dialog', { name: 'Confirm re-run analysis' }).waitFor({ state: 'visible' });
        await page.getByText(/Analysis may take some time/).waitFor({ state: 'visible' });
        await page.getByRole('button', { name: 'Cancel' }).waitFor({ state: 'visible' });
        await page.getByRole('button', { name: 'Confirm' }).waitFor({ state: 'visible' });

        // This dialog is mixed in with the page content rather than on its own portal
        await hideBySelector(page, '[data-aria-hidden="true"]');

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
