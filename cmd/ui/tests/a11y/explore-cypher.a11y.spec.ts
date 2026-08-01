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

import { expect, expectNoAccessibilityViolations, test } from '../fixtures';

test.describe('WCAG A/AA Violations - Explore - Cypher Tab', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/ui/explore');

        const cypherTab = page.getByRole('tab', { name: 'Cypher' });
        await cypherTab.click();
        await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });
    });

    test('Empty query', async ({ page, makeAxeBuilder }, testInfo) => {
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With full query', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor({ state: 'visible' });

        await cypherEditor.fill(query);

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Tag Results to Zone dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor({ state: 'visible' });

        await cypherEditor.fill(query);

        const tagButton = page.getByRole('button', { name: 'Tag' });
        await tagButton.waitFor({ state: 'visible' });
        await tagButton.click();

        await page.getByRole('button', { name: 'Zone' }).click();

        const dialog = page.getByRole('dialog', {
            name: 'Tag Results to Zone',
        });
        await dialog.waitFor({ state: 'visible' });

        const selectZoneControl = dialog.getByRole('combobox');
        const cancelButton = dialog.getByRole('button', { name: 'Cancel' });
        const continueButton = dialog.getByRole('button', {
            name: 'Continue',
        });

        await selectZoneControl.waitFor({ state: 'visible' });
        await cancelButton.waitFor({ state: 'visible' });
        await continueButton.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Tag Results to Label dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor({ state: 'visible' });

        await cypherEditor.fill(query);

        const tagButton = page.getByRole('button', { name: 'Tag' });
        await tagButton.waitFor({ state: 'visible' });
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
        await dialog.waitFor({ state: 'visible' });

        const selectLabelControl = dialog.getByRole('combobox');
        const cancelButton = dialog.getByRole('button', { name: 'Cancel' });
        const continueButton = dialog.getByRole('button', {
            name: 'Continue',
        });

        await selectLabelControl.waitFor({ state: 'visible' });
        await cancelButton.waitFor({ state: 'visible' });
        await continueButton.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Save Query dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor({ state: 'visible' });

        await cypherEditor.fill(query);

        const saveQueryButton = page.getByRole('button', {
            name: 'Save query',
            exact: true,
        });

        await saveQueryButton.waitFor({ state: 'visible' });
        await saveQueryButton.click();

        const dialog = page.getByTestId('save-query-dialog');

        await dialog.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Save As New Query dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await cypherEditor.waitFor({ state: 'visible' });

        await cypherEditor.fill(query);

        const saveQueryButton = page.getByRole('button', {
            name: 'Save query',
            exact: true,
        });
        const saveQueryOptionsButton = page.getByRole('button', {
            name: 'Show save query options',
            exact: true,
        });

        await saveQueryButton.waitFor({ state: 'visible' });
        await saveQueryOptionsButton.waitFor({ state: 'visible' });
        await saveQueryOptionsButton.click();

        const saveAsButton = page.getByRole('button', {
            name: 'Save As',
            exact: true,
        });

        await saveAsButton.waitFor({ state: 'visible' });
        await saveAsButton.click();

        const dialog = page.getByRole('dialog', {
            name: 'Save As New Query',
        });

        await dialog.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Search with no results', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor({ state: 'visible' });
        await savedQueriesButton.click();

        const searchTextbox = page.getByRole('textbox', {
            name: 'Search',
            exact: true,
        });
        const searchTerm = 'a11y-no-results-9f7c2e1b';

        await searchTextbox.waitFor({ state: 'visible' });
        await searchTextbox.fill(searchTerm);

        const noResultsHeading = page.getByRole('heading', {
            name: 'No Results',
            exact: true,
        });
        await noResultsHeading.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Import dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor({ state: 'visible' });
        await savedQueriesButton.click();

        const importButton = page.getByRole('button', {
            name: 'Import',
            exact: true,
        });

        await importButton.waitFor({ state: 'visible' });
        await importButton.click();

        const dialog = page.getByRole('dialog');

        await dialog.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Export saved query', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 1';
        const queryName = `a11y-export-${testInfo.project.name}-${Date.now()}`;
        const savedQueryAccessibleName = `Run pre-built search query: ${queryName}`;
        let testError: unknown;
        let cleanupError: unknown;

        try {
            const cypherEditor = page.getByRole('textbox', {
                name: 'Cypher Editor',
            });

            await cypherEditor.waitFor({ state: 'visible' });

            await cypherEditor.fill(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await saveQueryButton.waitFor({ state: 'visible' });
            await saveQueryButton.click();

            const saveQueryDialog = page.getByTestId('save-query-dialog');
            const queryNameTextbox = saveQueryDialog.getByRole('textbox', {
                name: 'Query Name',
                exact: true,
            });
            const saveButton = saveQueryDialog.getByRole('button', {
                name: 'Save',
                exact: true,
            });

            await saveQueryDialog.waitFor({ state: 'visible' });
            await queryNameTextbox.fill(queryName);
            await saveButton.click();
            await saveQueryDialog.waitFor({ state: 'hidden' });
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await savedQueriesButton.waitFor({ state: 'visible' });
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await searchTextbox.waitFor({ state: 'visible' });
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await savedQueryButton.waitFor({ state: 'visible' });

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await autoRunCheckbox.waitFor({ state: 'visible' });

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }

            await savedQueryButton.click();

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();

            await expectNoAccessibilityViolations(testInfo, results, { page });
            const exportButtons = page.getByRole('button', {
                name: 'Export',
                exact: true,
            });

            const exportButton = exportButtons.first();

            await exportButton.waitFor({ state: 'visible' });

            const [download] = await Promise.all([page.waitForEvent('download'), exportButton.click()]);
            expect(download.suggestedFilename()).toMatch(/\.json$/i);
        } catch (error) {
            testError = error;
        } finally {
            try {
                await page.goto('/ui/explore');

                const cypherTab = page.getByRole('tab', {
                    name: 'Cypher',
                });

                await cypherTab.click();
                await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await savedQueriesButton.waitFor({ state: 'visible' });
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await searchTextbox.waitFor({ state: 'visible' });
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await savedQueryButton.or(noResultsHeading).waitFor({ state: 'visible' });

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await actionMenuButton.waitFor({ state: 'visible' });
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await deleteButton.waitFor({ state: 'visible' });
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await deleteDialog.waitFor({ state: 'visible' });
                    await confirmButton.click();

                    await deleteDialog.waitFor({ state: 'hidden' });
                    await savedQueryButton.waitFor({ state: 'detached' });
                }
            } catch (error) {
                cleanupError = error;
            }
        }

        if (testError !== undefined) {
            throw testError;
        }

        if (cleanupError !== undefined) {
            throw cleanupError;
        }
    });

    test('Saved Queries - Filter by Platforms', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor({ state: 'visible' });
        await savedQueriesButton.click();

        const platformsFilter = page.getByRole('combobox', {
            name: 'Platforms',
            exact: true,
        });

        await platformsFilter.waitFor({ state: 'visible' });
        await platformsFilter.click();

        const activeDirectoryOption = page.getByRole('option', {
            name: 'Active Directory',
            exact: true,
        });

        await activeDirectoryOption.waitFor({ state: 'visible' });
        await activeDirectoryOption.click();
        const selectedPlatformsFilter = page.getByRole('combobox', {
            name: /Active Directory/,
        });

        await selectedPlatformsFilter.waitFor({ state: 'visible' });
        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Filter by Categories', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor({ state: 'visible' });
        await savedQueriesButton.click();

        const categoriesFilter = page.getByRole('combobox', {
            name: 'Categories',
            exact: true,
        });

        await categoriesFilter.waitFor({ state: 'visible' });
        await categoriesFilter.click();

        const domainInformationOption = page.getByRole('option', {
            name: 'Domain Information',
            exact: true,
        });

        await domainInformationOption.waitFor({ state: 'visible' });
        const selectedCategoriesFilter = page.getByRole('combobox', {
            name: /Domain Information/,
        });

        await selectedCategoriesFilter.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Filter by Source', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await savedQueriesButton.waitFor({ state: 'visible' });
        await savedQueriesButton.click();

        const sourceFilter = page.getByRole('combobox', {
            name: 'Source',
            exact: true,
        });

        await sourceFilter.waitFor({ state: 'visible' });
        await sourceFilter.click();

        const prebuiltOption = page.getByRole('option', {
            name: 'Prebuilt',
            exact: true,
        });

        await prebuiltOption.waitFor({ state: 'visible' });
        const selectedSourceFilter = page.getByRole('combobox', {
            name: /Prebuilt/,
        });

        await selectedSourceFilter.waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Query dots menu', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 1';
        const queryName = `a11y-dots-menu-${testInfo.project.name}-${Date.now()}`;
        const savedQueryAccessibleName = `Run pre-built search query: ${queryName}`;
        let testError: unknown;
        let cleanupError: unknown;

        try {
            const cypherEditor = page.getByRole('textbox', {
                name: 'Cypher Editor',
            });

            await cypherEditor.waitFor({ state: 'visible' });

            await cypherEditor.fill(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await saveQueryButton.waitFor({ state: 'visible' });
            await saveQueryButton.click();

            const saveQueryDialog = page.getByTestId('save-query-dialog');
            const queryNameTextbox = saveQueryDialog.getByRole('textbox', {
                name: 'Query Name',
                exact: true,
            });
            const saveButton = saveQueryDialog.getByRole('button', {
                name: 'Save',
                exact: true,
            });

            await saveQueryDialog.waitFor({ state: 'visible' });
            await queryNameTextbox.fill(queryName);
            await saveButton.click();
            await saveQueryDialog.waitFor({ state: 'hidden' });
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await savedQueriesButton.waitFor({ state: 'visible' });
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await searchTextbox.waitFor({ state: 'visible' });
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await savedQueryButton.waitFor({ state: 'visible' });

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await autoRunCheckbox.waitFor({ state: 'visible' });

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }
            await savedQueryButton.click();

            const actionMenuButton = page.getByRole('button', {
                name: 'Show saved query actions',
                exact: true,
            });

            await actionMenuButton.waitFor({ state: 'visible' });
            await actionMenuButton.click();

            const runButton = page.getByRole('button', {
                name: 'Run',
                exact: true,
            });
            const editShareButton = page.getByRole('button', {
                name: 'Edit/Share',
                exact: true,
            });
            const deleteButton = page.getByRole('button', {
                name: 'Delete',
                exact: true,
            });

            await runButton.waitFor({ state: 'visible' });
            await editShareButton.waitFor({ state: 'visible' });
            await deleteButton.waitFor({ state: 'visible' });

            const results = await makeAxeBuilder()
                .include('#content-wrapper')
                .include('[data-testid="saved-query-action-menu"]')
                .analyze();

            await expectNoAccessibilityViolations(testInfo, results, { page });
        } catch (error) {
            testError = error;
        } finally {
            try {
                await page.goto('/ui/explore');

                const cypherTab = page.getByRole('tab', {
                    name: 'Cypher',
                });

                await cypherTab.click();
                await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await savedQueriesButton.waitFor({ state: 'visible' });
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await searchTextbox.waitFor({ state: 'visible' });
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await savedQueryButton.or(noResultsHeading).waitFor({ state: 'visible' });

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await actionMenuButton.waitFor({ state: 'visible' });
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await deleteButton.waitFor({ state: 'visible' });
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await deleteDialog.waitFor({ state: 'visible' });
                    await confirmButton.click();

                    await deleteDialog.waitFor({ state: 'hidden' });
                    await savedQueryButton.waitFor({ state: 'detached' });
                }
            } catch (error) {
                cleanupError = error;
            }
        }

        if (testError !== undefined) {
            throw testError;
        }

        if (cleanupError !== undefined) {
            throw cleanupError;
        }
    });

    test('Edit Saved Query dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 1';
        const queryName = `a11y-edit-dialog-${testInfo.project.name}-${Date.now()}`;
        const savedQueryAccessibleName = `Run pre-built search query: ${queryName}`;
        let testError: unknown;
        let cleanupError: unknown;

        try {
            const cypherEditor = page.getByRole('textbox', {
                name: 'Cypher Editor',
            });

            await cypherEditor.waitFor({ state: 'visible' });

            await cypherEditor.fill(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await saveQueryButton.waitFor({ state: 'visible' });
            await saveQueryButton.click();

            const saveQueryDialog = page.getByTestId('save-query-dialog');
            const queryNameTextbox = saveQueryDialog.getByRole('textbox', {
                name: 'Query Name',
                exact: true,
            });
            const saveButton = saveQueryDialog.getByRole('button', {
                name: 'Save',
                exact: true,
            });

            await saveQueryDialog.waitFor({ state: 'visible' });
            await queryNameTextbox.fill(queryName);
            await saveButton.click();
            await saveQueryDialog.waitFor({ state: 'hidden' });
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await savedQueriesButton.waitFor({ state: 'visible' });
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await searchTextbox.waitFor({ state: 'visible' });
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await savedQueryButton.waitFor({ state: 'visible' });

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await autoRunCheckbox.waitFor({ state: 'visible' });

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }
            await savedQueryButton.click();

            const actionMenuButton = page.getByRole('button', {
                name: 'Show saved query actions',
                exact: true,
            });

            await actionMenuButton.waitFor({ state: 'visible' });
            await actionMenuButton.click();

            const actionMenu = page.getByTestId('saved-query-action-menu');
            await actionMenu.waitFor({ state: 'visible' });

            const editShareButton = page.getByRole('button', {
                name: 'Edit/Share',
                exact: true,
            });

            await editShareButton.waitFor({ state: 'visible' });
            await editShareButton.click();

            const editSavedQueryDialog = page.getByRole('dialog', {
                name: 'Edit Saved Query',
                exact: true,
            });

            await editSavedQueryDialog.waitFor({ state: 'visible' });

            const editQueryNameTextbox = editSavedQueryDialog.getByRole('textbox', {
                name: 'Query Name',
                exact: true,
            });
            const editQueryDescriptionTextbox = editSavedQueryDialog.getByRole('textbox', {
                name: 'Query Description',
                exact: true,
            });
            const cypherQueryEditor = editSavedQueryDialog.getByRole('textbox', {
                name: 'Cypher Query',
                exact: true,
            });
            const cancelButton = editSavedQueryDialog.getByRole('button', {
                name: 'Cancel',
                exact: true,
            });
            const editSaveButton = editSavedQueryDialog.getByRole('button', {
                name: 'Save',
                exact: true,
            });

            await editQueryNameTextbox.waitFor({ state: 'visible' });
            await editQueryDescriptionTextbox.waitFor({ state: 'visible' });
            await cypherQueryEditor.waitFor({ state: 'visible' });
            await cancelButton.waitFor({ state: 'visible' });
            await editSaveButton.waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();

            await expectNoAccessibilityViolations(testInfo, results, { page });
        } catch (error) {
            testError = error;
        } finally {
            try {
                await page.goto('/ui/explore');

                const cypherTab = page.getByRole('tab', {
                    name: 'Cypher',
                });

                await cypherTab.click();
                await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await savedQueriesButton.waitFor({ state: 'visible' });
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await searchTextbox.waitFor({ state: 'visible' });
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await savedQueryButton.or(noResultsHeading).waitFor({ state: 'visible' });

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await actionMenuButton.waitFor({ state: 'visible' });
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await deleteButton.waitFor({ state: 'visible' });
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await deleteDialog.waitFor({ state: 'visible' });
                    await confirmButton.click();

                    await deleteDialog.waitFor({ state: 'hidden' });
                    await savedQueryButton.waitFor({ state: 'detached' });
                }
            } catch (error) {
                cleanupError = error;
            }
        }

        if (testError !== undefined) {
            throw testError;
        }

        if (cleanupError !== undefined) {
            throw cleanupError;
        }
    });

    test('Delete Query dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 1';
        const queryName = `a11y-delete-dialog-${testInfo.project.name}-${Date.now()}`;
        const savedQueryAccessibleName = `Run pre-built search query: ${queryName}`;
        let testError: unknown;
        let cleanupError: unknown;

        try {
            const cypherEditor = page.getByRole('textbox', {
                name: 'Cypher Editor',
            });

            await cypherEditor.waitFor({ state: 'visible' });

            await cypherEditor.fill(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await saveQueryButton.waitFor({ state: 'visible' });
            await saveQueryButton.click();

            const saveQueryDialog = page.getByTestId('save-query-dialog');
            const queryNameTextbox = saveQueryDialog.getByRole('textbox', {
                name: 'Query Name',
                exact: true,
            });
            const saveButton = saveQueryDialog.getByRole('button', {
                name: 'Save',
                exact: true,
            });

            await saveQueryDialog.waitFor({ state: 'visible' });
            await queryNameTextbox.fill(queryName);
            await saveButton.click();
            await saveQueryDialog.waitFor({ state: 'hidden' });
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await savedQueriesButton.waitFor({ state: 'visible' });
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await searchTextbox.waitFor({ state: 'visible' });
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await savedQueryButton.waitFor({ state: 'visible' });

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await autoRunCheckbox.waitFor({ state: 'visible' });

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }
            await savedQueryButton.click();

            const actionMenuButton = page.getByRole('button', {
                name: 'Show saved query actions',
                exact: true,
            });

            await actionMenuButton.waitFor({ state: 'visible' });
            await actionMenuButton.click();

            const actionMenu = page.getByTestId('saved-query-action-menu');
            await actionMenu.waitFor({ state: 'visible' });

            const deleteButton = page.getByRole('button', {
                name: 'Delete',
                exact: true,
            });

            await deleteButton.waitFor({ state: 'visible' });
            await deleteButton.click();

            const deleteDialog = page.getByRole('dialog', {
                name: 'Delete Query',
                exact: true,
            });

            await deleteDialog.waitFor({ state: 'visible' });

            const deleteQueryHeading = deleteDialog.getByRole('heading', {
                name: 'Delete Query',
                exact: true,
            });
            const confirmationText = deleteDialog.getByText('Are you sure you want to delete this query?', {
                exact: true,
            });
            const cancelButton = deleteDialog.getByRole('button', {
                name: 'Cancel',
                exact: true,
            });
            const confirmButton = deleteDialog.getByRole('button', {
                name: 'Confirm',
                exact: true,
            });

            await deleteQueryHeading.waitFor({ state: 'visible' });
            await confirmationText.waitFor({ state: 'visible' });
            await cancelButton.waitFor({ state: 'visible' });
            await confirmButton.waitFor({ state: 'visible' });

            const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();

            await expectNoAccessibilityViolations(testInfo, results, { page });
        } catch (error) {
            testError = error;
        } finally {
            try {
                await page.goto('/ui/explore');

                const cypherTab = page.getByRole('tab', {
                    name: 'Cypher',
                });

                await cypherTab.click();
                await page.getByRole('textbox', { name: 'Cypher Editor' }).waitFor({ state: 'visible' });

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await savedQueriesButton.waitFor({ state: 'visible' });
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await searchTextbox.waitFor({ state: 'visible' });
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await savedQueryButton.or(noResultsHeading).waitFor({ state: 'visible' });

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await actionMenuButton.waitFor({ state: 'visible' });
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await deleteButton.waitFor({ state: 'visible' });
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await deleteDialog.waitFor({ state: 'visible' });
                    await confirmButton.click();

                    await deleteDialog.waitFor({ state: 'hidden' });
                    await savedQueryButton.waitFor({ state: 'detached' });
                }
            } catch (error) {
                cleanupError = error;
            }
        }

        if (testError !== undefined) {
            throw testError;
        }

        if (cleanupError !== undefined) {
            throw cleanupError;
        }
    });
});
