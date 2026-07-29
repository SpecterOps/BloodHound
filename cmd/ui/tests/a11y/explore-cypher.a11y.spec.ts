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
        await expect(cypherTab).toHaveAttribute('aria-selected', 'true');
    });

    test('Empty query', async ({ page, makeAxeBuilder }, testInfo) => {
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await expect(cypherEditor).toBeVisible();
        await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('With full query', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await expect(cypherEditor).toBeVisible();
        await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

        await cypherEditor.fill(query);
        await expect(cypherEditor).toContainText(query);

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Tag Results to Zone dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await expect(cypherEditor).toBeVisible();
        await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

        await cypherEditor.fill(query);
        await expect(cypherEditor).toContainText(query);

        const tagButton = page.getByRole('button', { name: 'Tag' });
        await expect(tagButton).toBeVisible();
        await expect(tagButton).toBeEnabled();
        await tagButton.click();

        await page.getByRole('button', { name: 'Zone' }).click();

        const dialog = page.getByRole('dialog', {
            name: 'Tag Results to Zone',
        });
        await expect(dialog).toBeVisible();

        const selectZoneControl = dialog.getByRole('combobox');
        const cancelButton = dialog.getByRole('button', { name: 'Cancel' });
        const continueButton = dialog.getByRole('button', {
            name: 'Continue',
        });

        await expect(selectZoneControl).toBeVisible();
        await expect(cancelButton).toBeVisible();
        await expect(continueButton).toBeVisible();
        await expect(continueButton).toBeDisabled();

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Tag Results to Label dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await expect(cypherEditor).toBeVisible();
        await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

        await cypherEditor.fill(query);
        await expect(cypherEditor).toContainText(query);

        const tagButton = page.getByRole('button', { name: 'Tag' });
        await expect(tagButton).toBeVisible();
        await expect(tagButton).toBeEnabled();
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
        await expect(dialog).toBeVisible();

        const selectLabelControl = dialog.getByRole('combobox');
        const cancelButton = dialog.getByRole('button', { name: 'Cancel' });
        const continueButton = dialog.getByRole('button', {
            name: 'Continue',
        });

        await expect(selectLabelControl).toBeVisible();
        await expect(cancelButton).toBeVisible();
        await expect(continueButton).toBeVisible();
        await expect(continueButton).toBeDisabled();

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Save Query dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await expect(cypherEditor).toBeVisible();
        await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

        await cypherEditor.fill(query);
        await expect(cypherEditor).toContainText(query);

        const saveQueryButton = page.getByRole('button', {
            name: 'Save query',
            exact: true,
        });

        await expect(saveQueryButton).toBeVisible();
        await expect(saveQueryButton).toBeEnabled();
        await saveQueryButton.click();

        const dialog = page.getByTestId('save-query-dialog');

        await expect(dialog).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Save As New Query dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const query = 'MATCH (n) RETURN n LIMIT 10';
        const cypherEditor = page.getByRole('textbox', {
            name: 'Cypher Editor',
        });

        await expect(cypherEditor).toBeVisible();
        await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

        await cypherEditor.fill(query);
        await expect(cypherEditor).toContainText(query);

        const saveQueryButton = page.getByRole('button', {
            name: 'Save query',
            exact: true,
        });
        const saveQueryOptionsButton = page.getByRole('button', {
            name: 'Show save query options',
            exact: true,
        });

        await expect(saveQueryButton).toBeVisible();
        await expect(saveQueryButton).toBeEnabled();
        await expect(saveQueryOptionsButton).toBeVisible();
        await expect(saveQueryOptionsButton).toBeEnabled();
        await saveQueryOptionsButton.click();

        const saveAsButton = page.getByRole('button', {
            name: 'Save As',
            exact: true,
        });

        await expect(saveAsButton).toBeVisible();
        await saveAsButton.click();

        const dialog = page.getByRole('dialog', {
            name: 'Save As New Query',
        });

        await expect(dialog).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').include('[role="dialog"]').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Search with no results', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await expect(savedQueriesButton).toBeVisible();
        await expect(savedQueriesButton).toBeEnabled();
        await savedQueriesButton.click();

        const searchTextbox = page.getByRole('textbox', {
            name: 'Search',
            exact: true,
        });
        const searchTerm = 'a11y-no-results-9f7c2e1b';

        await expect(searchTextbox).toBeVisible();
        await searchTextbox.fill(searchTerm);
        await expect(searchTextbox).toHaveValue(searchTerm);

        const noResultsHeading = page.getByRole('heading', {
            name: 'No Results',
            exact: true,
        });
        await expect(noResultsHeading).toBeVisible();

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Import dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await expect(savedQueriesButton).toBeVisible();
        await expect(savedQueriesButton).toBeEnabled();
        await savedQueriesButton.click();

        const importButton = page.getByRole('button', {
            name: 'Import',
            exact: true,
        });

        await expect(importButton).toBeVisible();
        await expect(importButton).toBeEnabled();
        await importButton.click();

        const dialog = page.getByRole('dialog');

        await expect(dialog).toBeVisible();

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

            await expect(cypherEditor).toBeVisible();
            await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

            await cypherEditor.fill(query);
            await expect(cypherEditor).toContainText(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await expect(saveQueryButton).toBeVisible();
            await expect(saveQueryButton).toBeEnabled();
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

            await expect(saveQueryDialog).toBeVisible();
            await queryNameTextbox.fill(queryName);
            await expect(saveButton).toBeEnabled();
            await saveButton.click();
            await expect(saveQueryDialog).toBeHidden();
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await expect(savedQueriesButton).toBeVisible();
            await expect(savedQueriesButton).toBeEnabled();
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await expect(searchTextbox).toBeVisible();
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await expect(savedQueryButton).toBeVisible();

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await expect(autoRunCheckbox).toBeVisible();

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }

            await expect(autoRunCheckbox).not.toBeChecked();

            await savedQueryButton.click();

            const results = await makeAxeBuilder().include('#content-wrapper').analyze();

            await expectNoAccessibilityViolations(testInfo, results, { page });
            const exportButtons = page.getByRole('button', {
                name: 'Export',
                exact: true,
            });

            await expect(exportButtons).toHaveCount(2);

            const exportButton = exportButtons.first();

            await expect(exportButton).toBeVisible();
            await expect(exportButton).toBeEnabled();

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
                await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await expect(savedQueriesButton).toBeVisible();
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await expect(searchTextbox).toBeVisible();
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await expect(savedQueryButton.or(noResultsHeading)).toBeVisible();

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await expect(actionMenuButton).toBeVisible();
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await expect(deleteButton).toBeVisible();
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await expect(deleteDialog).toBeVisible();
                    await expect(confirmButton).toBeEnabled();
                    await confirmButton.click();

                    await expect(deleteDialog).toBeHidden();
                    await expect(savedQueryButton).toHaveCount(0);
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

        await expect(savedQueriesButton).toBeVisible();
        await expect(savedQueriesButton).toBeEnabled();
        await savedQueriesButton.click();

        const platformsFilter = page.getByRole('combobox', {
            name: 'Platforms',
            exact: true,
        });

        await expect(platformsFilter).toBeVisible();
        await expect(platformsFilter).toBeEnabled();
        await platformsFilter.click();

        const activeDirectoryOption = page.getByRole('option', {
            name: 'Active Directory',
            exact: true,
        });

        await expect(activeDirectoryOption).toBeVisible();
        await activeDirectoryOption.click();
        const selectedPlatformsFilter = page.getByRole('combobox', {
            name: /Active Directory/,
        });

        await expect(selectedPlatformsFilter).toBeVisible();
        await expect(selectedPlatformsFilter).toContainText('Active Directory');
        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Filter by Categories', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await expect(savedQueriesButton).toBeVisible();
        await expect(savedQueriesButton).toBeEnabled();
        await savedQueriesButton.click();

        const categoriesFilter = page.getByRole('combobox', {
            name: 'Categories',
            exact: true,
        });

        await expect(categoriesFilter).toBeVisible();
        await expect(categoriesFilter).toBeEnabled();
        await categoriesFilter.click();

        const domainInformationOption = page.getByRole('option', {
            name: 'Domain Information',
            exact: true,
        });

        await expect(domainInformationOption).toBeVisible();
        const selectedCategoriesFilter = page.getByRole('combobox', {
            name: /Domain Information/,
        });

        await expect(selectedCategoriesFilter).toBeVisible();
        await expect(selectedCategoriesFilter).toContainText('Domain Information');

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();

        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Saved Queries - Filter by Source', async ({ page, makeAxeBuilder }, testInfo) => {
        const savedQueriesButton = page.getByRole('button', {
            name: 'Saved Queries',
            exact: true,
        });

        await expect(savedQueriesButton).toBeVisible();
        await expect(savedQueriesButton).toBeEnabled();
        await savedQueriesButton.click();

        const sourceFilter = page.getByRole('combobox', {
            name: 'Source',
            exact: true,
        });

        await expect(sourceFilter).toBeVisible();
        await expect(sourceFilter).toBeEnabled();
        await sourceFilter.click();

        const prebuiltOption = page.getByRole('option', {
            name: 'Prebuilt',
            exact: true,
        });

        await expect(prebuiltOption).toBeVisible();
        const selectedSourceFilter = page.getByRole('combobox', {
            name: /Prebuilt/,
        });

        await expect(selectedSourceFilter).toBeVisible();
        await expect(selectedSourceFilter).toContainText('Prebuilt');

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

            await expect(cypherEditor).toBeVisible();
            await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

            await cypherEditor.fill(query);
            await expect(cypherEditor).toContainText(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await expect(saveQueryButton).toBeVisible();
            await expect(saveQueryButton).toBeEnabled();
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

            await expect(saveQueryDialog).toBeVisible();
            await queryNameTextbox.fill(queryName);
            await expect(saveButton).toBeEnabled();
            await saveButton.click();
            await expect(saveQueryDialog).toBeHidden();
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await expect(savedQueriesButton).toBeVisible();
            await expect(savedQueriesButton).toBeEnabled();
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await expect(searchTextbox).toBeVisible();
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await expect(savedQueryButton).toBeVisible();

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await expect(autoRunCheckbox).toBeVisible();

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }

            await expect(autoRunCheckbox).not.toBeChecked();
            await savedQueryButton.click();

            const actionMenuButton = page.getByRole('button', {
                name: 'Show saved query actions',
                exact: true,
            });

            await expect(actionMenuButton).toBeVisible();
            await expect(actionMenuButton).toBeEnabled();
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

            await expect(runButton).toBeVisible();
            await expect(editShareButton).toBeVisible();
            await expect(deleteButton).toBeVisible();

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
                await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await expect(savedQueriesButton).toBeVisible();
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await expect(searchTextbox).toBeVisible();
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await expect(savedQueryButton.or(noResultsHeading)).toBeVisible();

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await expect(actionMenuButton).toBeVisible();
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await expect(deleteButton).toBeVisible();
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await expect(deleteDialog).toBeVisible();
                    await expect(confirmButton).toBeEnabled();
                    await confirmButton.click();

                    await expect(deleteDialog).toBeHidden();
                    await expect(savedQueryButton).toHaveCount(0);
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

            await expect(cypherEditor).toBeVisible();
            await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

            await cypherEditor.fill(query);
            await expect(cypherEditor).toContainText(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await expect(saveQueryButton).toBeVisible();
            await expect(saveQueryButton).toBeEnabled();
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

            await expect(saveQueryDialog).toBeVisible();
            await queryNameTextbox.fill(queryName);
            await expect(saveButton).toBeEnabled();
            await saveButton.click();
            await expect(saveQueryDialog).toBeHidden();
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await expect(savedQueriesButton).toBeVisible();
            await expect(savedQueriesButton).toBeEnabled();
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await expect(searchTextbox).toBeVisible();
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await expect(savedQueryButton).toBeVisible();

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await expect(autoRunCheckbox).toBeVisible();

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }

            await expect(autoRunCheckbox).not.toBeChecked();
            await savedQueryButton.click();

            const actionMenuButton = page.getByRole('button', {
                name: 'Show saved query actions',
                exact: true,
            });

            await expect(actionMenuButton).toBeVisible();
            await expect(actionMenuButton).toBeEnabled();
            await actionMenuButton.click();

            const actionMenu = page.getByTestId('saved-query-action-menu');
            await expect(actionMenu).toBeVisible();

            const editShareButton = page.getByRole('button', {
                name: 'Edit/Share',
                exact: true,
            });

            await expect(editShareButton).toBeVisible();
            await editShareButton.click();

            const editSavedQueryDialog = page.getByRole('dialog', {
                name: 'Edit Saved Query',
                exact: true,
            });

            await expect(editSavedQueryDialog).toBeVisible();

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

            await expect(editQueryNameTextbox).toBeVisible();
            await expect(editQueryDescriptionTextbox).toBeVisible();
            await expect(cypherQueryEditor).toBeVisible();
            await expect(cancelButton).toBeVisible();
            await expect(editSaveButton).toBeVisible();

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
                await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await expect(savedQueriesButton).toBeVisible();
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await expect(searchTextbox).toBeVisible();
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await expect(savedQueryButton.or(noResultsHeading)).toBeVisible();

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await expect(actionMenuButton).toBeVisible();
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await expect(deleteButton).toBeVisible();
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await expect(deleteDialog).toBeVisible();
                    await expect(confirmButton).toBeEnabled();
                    await confirmButton.click();

                    await expect(deleteDialog).toBeHidden();
                    await expect(savedQueryButton).toHaveCount(0);
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

            await expect(cypherEditor).toBeVisible();
            await expect(cypherEditor).toHaveAttribute('contenteditable', 'true');

            await cypherEditor.fill(query);
            await expect(cypherEditor).toContainText(query);

            const saveQueryButton = page.getByRole('button', {
                name: 'Save query',
                exact: true,
            });

            await expect(saveQueryButton).toBeVisible();
            await expect(saveQueryButton).toBeEnabled();
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

            await expect(saveQueryDialog).toBeVisible();
            await queryNameTextbox.fill(queryName);
            await expect(saveButton).toBeEnabled();
            await saveButton.click();
            await expect(saveQueryDialog).toBeHidden();
            await page.goto('/ui/explore');

            const cypherTab = page.getByRole('tab', {
                name: 'Cypher',
            });

            await cypherTab.click();
            await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

            const savedQueriesButton = page.getByRole('button', {
                name: 'Saved Queries',
                exact: true,
            });

            await expect(savedQueriesButton).toBeVisible();
            await expect(savedQueriesButton).toBeEnabled();
            await savedQueriesButton.click();

            const searchTextbox = page.getByRole('textbox', {
                name: 'Search',
                exact: true,
            });

            await expect(searchTextbox).toBeVisible();
            await searchTextbox.fill(queryName);

            const savedQueryButton = page.getByRole('button', {
                name: savedQueryAccessibleName,
                exact: true,
            });

            await expect(savedQueryButton).toBeVisible();

            const autoRunCheckbox = page.getByRole('checkbox', {
                name: 'Auto-run selected query',
                exact: true,
            });

            await expect(autoRunCheckbox).toBeVisible();

            if (await autoRunCheckbox.isChecked()) {
                await autoRunCheckbox.uncheck();
            }

            await expect(autoRunCheckbox).not.toBeChecked();
            await savedQueryButton.click();

            const actionMenuButton = page.getByRole('button', {
                name: 'Show saved query actions',
                exact: true,
            });

            await expect(actionMenuButton).toBeVisible();
            await expect(actionMenuButton).toBeEnabled();
            await actionMenuButton.click();

            const actionMenu = page.getByTestId('saved-query-action-menu');
            await expect(actionMenu).toBeVisible();

            const deleteButton = page.getByRole('button', {
                name: 'Delete',
                exact: true,
            });

            await expect(deleteButton).toBeVisible();
            await deleteButton.click();

            const deleteDialog = page.getByRole('dialog', {
                name: 'Delete Query',
                exact: true,
            });

            await expect(deleteDialog).toBeVisible();

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

            await expect(deleteQueryHeading).toBeVisible();
            await expect(confirmationText).toBeVisible();
            await expect(cancelButton).toBeVisible();
            await expect(confirmButton).toBeVisible();

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
                await expect(cypherTab).toHaveAttribute('aria-selected', 'true');

                const savedQueriesButton = page.getByRole('button', {
                    name: 'Saved Queries',
                    exact: true,
                });

                await expect(savedQueriesButton).toBeVisible();
                await savedQueriesButton.click();

                const searchTextbox = page.getByRole('textbox', {
                    name: 'Search',
                    exact: true,
                });

                await expect(searchTextbox).toBeVisible();
                await searchTextbox.fill(queryName);

                const savedQueryButton = page.getByRole('button', {
                    name: savedQueryAccessibleName,
                    exact: true,
                });
                const noResultsHeading = page.getByRole('heading', {
                    name: 'No Results',
                    exact: true,
                });

                await expect(savedQueryButton.or(noResultsHeading)).toBeVisible();

                if (await savedQueryButton.isVisible()) {
                    const actionMenuButton = page.getByRole('button', {
                        name: 'Show saved query actions',
                        exact: true,
                    });

                    await expect(actionMenuButton).toBeVisible();
                    await actionMenuButton.click();

                    const deleteButton = page.getByRole('button', {
                        name: 'Delete',
                        exact: true,
                    });

                    await expect(deleteButton).toBeVisible();
                    await deleteButton.click();

                    const deleteDialog = page.getByRole('dialog', {
                        name: 'Delete Query',
                        exact: true,
                    });
                    const confirmButton = deleteDialog.getByRole('button', {
                        name: 'Confirm',
                        exact: true,
                    });

                    await expect(deleteDialog).toBeVisible();
                    await expect(confirmButton).toBeEnabled();
                    await confirmButton.click();

                    await expect(deleteDialog).toBeHidden();
                    await expect(savedQueryButton).toHaveCount(0);
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
