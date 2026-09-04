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
import {
    installAssetGroupLabelDetailsStub,
    installAssetGroupLabelMemberStub,
    installAssetGroupLabelSelectorStub,
    installAssetGroupLabelsSearchStub,
    installAssetGroupLabelStub,
} from 'bh-playwright-testing/stubs';

const LABELS_URL = '/ui/privilege-zones/labels/2/details';
const EDIT_LABEL_URL = '/ui/privilege-zones/labels/2/save';
const CREATE_RULE_URL = '/ui/privilege-zones/labels/2/rules/save';
const EDIT_RULE_URL = '/ui/privilege-zones/labels/2/rules/3303/save';

test.describe('WCAG A/AA violations - Privilege Zones - Labels tab', () => {
    test.describe('Label details panel', () => {
        test('default state', async ({ page, goAndWaitFor, checkA11y }) => {
            await goAndWaitFor(LABELS_URL, page.getByRole('heading', { name: 'Label Details' }));

            await checkA11y();
        });

        test('search', async ({ page, goAndWaitFor, checkA11y }) => {
            await installAssetGroupLabelsSearchStub(page);
            await goAndWaitFor(LABELS_URL, page.getByRole('heading', { name: 'Label Details' }));

            // Type at least 3 characters to trigger the debounced search and open the results popover.
            await page.getByTestId('privilege-zone-detail-search-bar').fill('ADMIN');

            // Wait for the filtered Objects to render before scanning.
            await page.getByText('PLAYWRIGHT_LABEL_ADMIN_USER').waitFor({ state: 'visible' });

            await checkA11y({ include: '[data-radix-popper-content-wrapper]' });
        });

        test('search with no results', async ({ page, goAndWaitFor, checkA11y }) => {
            await installAssetGroupLabelsSearchStub(page);
            await goAndWaitFor(LABELS_URL, page.getByRole('heading', { name: 'Label Details' }));

            // Type at query that will have no matches
            await page.getByTestId('privilege-zone-detail-search-bar').fill('XXXYYY');

            // Wait for the filtered Objects to render before scanning.
            await page.getByText('No results').first().waitFor({ state: 'visible' });

            await checkA11y({ include: '[data-radix-popper-content-wrapper]' });
        });

        test('expanded rules and objects', async ({ page, goAndWaitFor, checkA11y }) => {
            // Stub Label data so custom Rules and Objects render deterministically
            // Labels deliberately do not expose Default Rules.
            await installAssetGroupLabelDetailsStub(page);
            await installAssetGroupLabelMemberStub(page);
            await goAndWaitFor(LABELS_URL, page.getByRole('heading', { name: 'Label Details' }));

            const caret = page.getByTestId('privilege-zones_details_Computer-accordion_open-toggle-button');
            await caret.waitFor({ state: 'visible' });
            await caret.click();

            await checkA11y();
        });
    });

    test.describe('Side panels', () => {
        test('Rule side panel tab', async ({ page, goAndWaitFor, checkA11y }) => {
            // Stub Label Details with custom rules, plus the first rule's detail response.
            await installAssetGroupLabelDetailsStub(page);
            await installAssetGroupLabelSelectorStub(page);
            await goAndWaitFor(LABELS_URL, page.getByRole('heading', { name: 'Label Details' }));

            // Selecting the first Custom Rule updates the route and switches the details panel to Rule.
            await page.getByRole('button', { name: 'PLAYWRIGHT_LABEL_RULE_1' }).click();
            await page.getByText('Playwright stubbed label rule').waitFor({ state: 'visible' });

            await checkA11y();
        });

        test('Object side panel tab', async ({ page, goAndWaitFor, checkA11y }) => {
            // Stub Label Details with objects grouped by node type, plus the first object's detail response.
            await installAssetGroupLabelDetailsStub(page);
            await installAssetGroupLabelMemberStub(page);
            await goAndWaitFor(LABELS_URL, page.getByRole('heading', { name: 'Label Details' }));

            // Open the first node type and select its first object to switch the details panel to Object.
            await page.getByTestId('privilege-zones_details_Computer-accordion_open-toggle-button').click();
            await page.getByRole('button', { name: 'PLAYWRIGHT_LABEL_COMPUTER_1' }).click();
            await page.getByText('Node Type:').waitFor({ state: 'visible' });

            await checkA11y();
        });
    });
});

test.describe('WCAG A/AA violations - Privilege Zones - Label save pages', () => {
    test('Edit Label page', async ({ page, goAndWaitFor, checkA11y }) => {
        // Stub the single label so the Edit Label form fields populate instead of showing a skeleton.
        await installAssetGroupLabelStub(page);
        await goAndWaitFor(EDIT_LABEL_URL, page.getByTestId('privilege-zones_save_tag-form_name-input'));

        await checkA11y();
    });

    test('Create Label rule page', async ({ page, goAndWaitFor, checkA11y }) => {
        // Create mode does not fetch a selector; only the tag info is needed for form context.
        await installAssetGroupLabelStub(page);
        await goAndWaitFor(CREATE_RULE_URL, page.getByTestId('rule-form'));

        await checkA11y();
    });

    test('Edit Label rule page', async ({ page, goAndWaitFor, checkA11y }) => {
        // Edit mode fetches both the label (for context) and the selector being edited.
        await installAssetGroupLabelStub(page);
        await installAssetGroupLabelSelectorStub(page);
        await goAndWaitFor(EDIT_RULE_URL, page.getByTestId('rule-form'));

        await checkA11y();
    });
});
