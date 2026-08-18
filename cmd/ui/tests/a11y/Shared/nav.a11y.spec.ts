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

const MAIN_NAV_SCOPE = '#app-root > nav';

test.describe('Shared navigation - has no detectable WCAG A/AA violations', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/ui/administration/data-quality');
        await page.getByRole('heading', { name: 'Data Quality', exact: true }).waitFor({ state: 'visible' });
        await page.locator(MAIN_NAV_SCOPE).waitFor({ state: 'visible' });
        await page.getByRole('button', { name: 'Administration', exact: true }).waitFor({ state: 'visible' });
        await hideBySelector(page, '#content-wrapper');
    });

    test('nav menu', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('button', { name: 'Toggle Navigation', expanded: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include(MAIN_NAV_SCOPE).analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('collapsed nav menu', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('button', { name: 'Toggle Navigation', expanded: true }).click();
        await page.getByRole('button', { name: 'Toggle Navigation', expanded: false }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include(MAIN_NAV_SCOPE).analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('administration subnav menu visible', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('button', { name: 'Administration', exact: true }).click();
        const subNav = page.getByTestId('sub-nav');
        await subNav.waitFor({ state: 'visible' });
        await subNav.getByText('File Ingest', { exact: true }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include(MAIN_NAV_SCOPE).analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
