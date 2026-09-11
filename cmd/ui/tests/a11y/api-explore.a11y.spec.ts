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

import { test } from 'bh-playwright-testing';

test.describe('WCAG A/AA Violations - Explore - API Explorer', () => {
    test.beforeEach(async ({ goAndWaitFor, page }) => {
        await goAndWaitFor('/ui/api-explorer', page.getByRole('textbox', { name: 'Filter by tag or path' }));
    });

    test('default state', async ({ checkA11y }) => {
        await checkA11y();
    });

    test('expanded resource', async ({ page, checkA11y }) => {
        await page.getByRole('button', { name: 'get /api/version' }).click();
        await page.getByText('Returns the supported API versions.').waitFor({ state: 'visible' });

        await checkA11y();
    });

    test('expanded disabled resource', async ({ page, checkA11y }) => {
        await page.getByRole('button', { name: 'get /api/v2/saml', exact: true }).click();
        await page.getByText('Deprecated: This endpoint').waitFor({ state: 'visible' });

        await checkA11y();
    });

    test('filter with no results', async ({ page, checkA11y }) => {
        await page.getByRole('textbox', { name: 'Filter by tag or path' }).fill('no-matching-api-resource');
        await page.getByRole('heading', { name: 'No operations defined in spec!' }).waitFor({ state: 'visible' });

        await checkA11y();
    });

    test('expanded Schemas', async ({ page, checkA11y }) => {
        // Set filter for empty reponse for easier view of Schemas
        await page.getByRole('textbox', { name: 'Filter by tag or path' }).fill('no-matching-api-resource');

        // Expand a schema and its property
        await page.getByRole('button', { name: 'api.error-detail' }).click();
        await page.getByRole('button', { name: '[...]' }).first().click();

        await page.getByText('The context in which the').waitFor({ state: 'visible' });

        await checkA11y();
    });
});
