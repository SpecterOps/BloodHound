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

import { Page, TestInfo } from '@playwright/test';
import { hideBySelector, test } from 'bh-playwright-testing';

const installSaveQueryStub = async (page: Page, testInfo: TestInfo) => {
    const queryName = `a11y-export-${testInfo.project.name}-${Date.now()}`;

    const savedQuery = {
        id: 1,
        name: queryName,
        description: '',
        query: 'MATCH (n) RETURN n LIMIT 10',
        user_id: '0e712979-d6e9-4496-9f31-f96f5565873e',
        scope: 'owned',
    };

    const self = {
        data: {
            sso_provider_id: null,
            AuthSecret: {
                digest_method: 'argon2',
                expires_at: '2026-11-15T14:32:05.868773Z',
                id: 1,
            },
            first_name: 'BloodHound',
            last_name: 'Dev',
            email_address: 'spam@example.com',
            principal_name: 'admin',
            last_login: '2026-08-24T17:12:33.222043Z',
            is_disabled: false,
            all_environments: true,
            environment_targeted_access_control: [],
            eula_accepted: true,
            id: '0e712979-d6e9-4496-9f31-f96f5565873e',
        },
    };

    // Fetch the real /api/v2/self response and override the id so the logged-in
    // user (self) owns the stubbed saved query and can edit it. The rest of the
    // self payload is preserved so the page keeps loading normally.
    await page.route('**/api/v2/self', async (route) => {
        const request = route.request();
        const pathname = new URL(request.url()).pathname;

        if (request.method() === 'GET' && pathname === '/api/v2/self') {
            return route.fulfill({ json: self });
        }

        return route.fallback();
    });

    // Stub the saved queries list so a single owned query already exists, and the
    // per-query permissions lookup opened by the Edit/Share dialog. Any other
    // saved-query requests fall through to the real API.
    await page.route('**/api/v2/saved-queries**', async (route) => {
        const request = route.request();
        const pathname = new URL(request.url()).pathname;

        if (request.method() === 'GET' && pathname === '/api/v2/saved-queries') {
            return route.fulfill({
                json: { data: [savedQuery], count: 1, limit: 10, skip: 0 },
            });
        }

        if (request.method() === 'GET' && pathname === `/api/v2/saved-queries/${savedQuery.id}/permissions`) {
            return route.fulfill({
                json: { data: { query_id: savedQuery.id, public: false, shared_to_user_ids: [] } },
            });
        }

        return route.fallback();
    });
};

test.describe('WCAG A/AA Violations - Explore - Cypher Tab', () => {
    test.beforeEach(async ({ goAndWaitFor, page }, testInfo) => {
        await goAndWaitFor('/ui/explore?exploreSearchTab=cypher', page.getByRole('textbox', { name: 'Cypher Editor' }));
    });

    test('Empty query', async ({ page, checkA11y }) => {
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor();

        await checkA11y();
    });

    test('With full query', async ({ page, checkA11y }) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor();

        await cypherEditor.fill(query);

        await checkA11y();
    });

    test('Tag Results to Zone dialog', async ({ page, checkA11y }) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor();
        await cypherEditor.fill(query);

        const tagButton = page.getByRole('button', { name: 'Tag' });
        await tagButton.waitFor();
        await tagButton.click();

        await page.getByRole('button', { name: 'Zone' }).click();

        const dialog = page.getByRole('dialog', {
            name: 'Tag Results to Zone',
        });
        await dialog.waitFor();

        await hideBySelector(page, '[data-radix-popper-content-wrapper]');
        await hideBySelector(page, '#content-wrapper');

        const selectZoneControl = dialog.getByRole('combobox');
        const cancelButton = dialog.getByRole('button', { name: 'Cancel' });
        const continueButton = dialog.getByRole('button', {
            name: 'Continue',
        });

        await selectZoneControl.waitFor();
        await cancelButton.waitFor();
        await continueButton.waitFor();

        await checkA11y({ include: '[role="dialog"]' });
    });

    test('Tag Results to Label dialog', async ({ page, checkA11y }) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor();

        await cypherEditor.fill(query);

        const tagButton = page.getByRole('button', { name: 'Tag' });
        await tagButton.waitFor();
        await tagButton.click();

        await page
            .getByRole('button', {
                name: 'Label',
                exact: true,
            })
            .click();
        const dialog = page.getByRole('dialog', {
            name: 'Tag Results to Label',
        });
        await dialog.waitFor();

        await hideBySelector(page, '[data-radix-popper-content-wrapper]');
        await hideBySelector(page, '#content-wrapper');

        const selectLabelControl = dialog.getByRole('combobox');
        const cancelButton = dialog.getByRole('button', { name: 'Cancel' });
        const continueButton = dialog.getByRole('button', {
            name: 'Continue',
        });

        await selectLabelControl.waitFor();
        await cancelButton.waitFor();
        await continueButton.waitFor();

        await checkA11y({ include: '[role="dialog"]' });
    });

    test('Save Query dialog', async ({ page, checkA11y }) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor();

        await cypherEditor.fill(query);

        const saveQueryButton = page.getByRole('button', {
            name: 'Save query',
            exact: true,
        });

        await saveQueryButton.waitFor();
        await saveQueryButton.click();

        await page.getByTestId('save-query-dialog').waitFor();

        await hideBySelector(page, 'nav');
        await hideBySelector(page, '#content-wrapper');

        await checkA11y({ include: '[role="dialog"]' });
    });

    test('Save As New Query dialog', async ({ page, checkA11y }) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor();

        await cypherEditor.fill(query);

        const saveQueryButton = page.getByRole('button', {
            name: 'Save query',
            exact: true,
        });
        const saveQueryOptionsButton = page.getByRole('button', {
            name: 'Show save query options',
            exact: true,
        });

        await saveQueryButton.waitFor();
        await saveQueryOptionsButton.waitFor();
        await saveQueryOptionsButton.click();

        const saveAsButton = page.getByRole('button', {
            name: 'Save As',
            exact: true,
        });

        await saveAsButton.waitFor();
        await saveAsButton.click();

        await page
            .getByRole('dialog', {
                name: 'Save As New Query',
            })
            .waitFor();

        await hideBySelector(page, 'nav');
        await hideBySelector(page, '#content-wrapper');

        await checkA11y({ include: '[role="dialog"]' });
    });

    test('Saved Queries - Search with no results', async ({ page, checkA11y }) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor();
        await savedQueriesButton.click();

        const searchTextbox = page.getByRole('textbox', {
            name: 'Search',
            exact: true,
        });
        const searchTerm = 'a11y-no-results-9f7c2e1b';

        await searchTextbox.waitFor();
        await searchTextbox.fill(searchTerm);

        const noResultsHeading = page.getByRole('heading', {
            name: 'No Results',
            exact: true,
        });
        await noResultsHeading.waitFor();

        await checkA11y();
    });

    test('Saved Queries - Import dialog', async ({ page, checkA11y }) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor();
        await savedQueriesButton.click();

        const importButton = page.getByRole('button', {
            name: 'Import',
            exact: true,
        });

        await importButton.waitFor();
        await importButton.click();

        await page.getByRole('dialog').waitFor();
        await hideBySelector(page, '#content-wrapper');

        await checkA11y({ include: '[role="dialog"]' });
    });

    test('Saved Queries - With filter', async ({ page, checkA11y }) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor();
        await savedQueriesButton.click();

        const platformsFilter = page.getByRole('combobox', {
            name: 'Platforms',
            exact: true,
        });

        await platformsFilter.waitFor();
        await platformsFilter.click();

        const activeDirectoryOption = page.getByRole('option', {
            name: 'Active Directory',
            exact: true,
        });

        await activeDirectoryOption.waitFor();
        await activeDirectoryOption.click();
        const selectedPlatformsFilter = page.getByRole('combobox', {
            name: /Active Directory/,
        });

        await selectedPlatformsFilter.waitFor();
        await checkA11y();
    });

    test('Saved Queries - Query dots menu', async ({ page, checkA11y }) => {
        // Expand queries panel and filter down to saved queries
        await page.getByRole('button', { name: 'Saved Queries', exact: true }).click();
        await page.getByRole('combobox', { name: 'Platforms', exact: true }).click();
        await page.getByRole('option', { name: 'Saved Queries', exact: true }).click();

        // Select the dots action from the saved query
        await page.getByTestId('saved-query-action-menu-trigger').click();

        await hideBySelector(page, '#root');

        await checkA11y({ include: '[data-radix-popper-content-wrapper]' });
    });
});

test.describe('WCAG A/AA Violations - Explore - Cypher Tab - Saved Queries', () => {
    test.beforeEach(async ({ goAndWaitFor, page }, testInfo) => {
        await installSaveQueryStub(page, testInfo);
        await goAndWaitFor('/ui/explore?exploreSearchTab=cypher', page.getByRole('textbox', { name: 'Cypher Editor' }));
    });

    // NOTE: Even with the hidden sections, this test still has incompletes in the axe-results.json
    test('Edit Saved Query dialog', async ({ page, checkA11y }) => {
        // Expand queries panel and filter down to saved queries
        await page.getByRole('button', { name: 'Saved Queries', exact: true }).click();
        await page.getByRole('combobox', { name: 'Platforms', exact: true }).click();
        await page.getByRole('option', { name: 'Saved Queries', exact: true }).click();

        // Select the dots action from the saved query
        await page.getByTestId('saved-query-action-menu-trigger').click();
        await page.getByRole('button', { name: 'Edit/Share' }).click();

        await hideBySelector(page, 'nav');
        await hideBySelector(page, '#content-wrapper');

        await checkA11y({ include: '[role="dialog"]' });
    });

    test('Delete Query dialog', async ({ page, checkA11y }) => {
        // Expand queries panel and filter down to saved queries
        await page.getByRole('button', { name: 'Saved Queries', exact: true }).click();
        await page.getByRole('combobox', { name: 'Platforms', exact: true }).click();
        await page.getByRole('option', { name: 'Saved Queries', exact: true }).click();

        // Select the dots action from the saved query
        await page.getByTestId('saved-query-action-menu-trigger').click();

        // Additional specifier as there are multiple Delete buttons in scope
        await page.getByTestId('saved-query-action-menu').getByRole('button', { name: 'Delete' }).click();

        await hideBySelector(page, 'nav');
        await hideBySelector(page, '#content-wrapper');

        await checkA11y();
    });
});
