// Copyright 2023 Specter Ops, Inc.
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

import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { AllEdgeTypes, EdgeCheckboxType } from '../../../edgeTypes';
import { getInitialPathFilters } from '../../../hooks';
import { act, render, screen } from '../../../test-utils';
import EdgeFilteringDialog from './EdgeFilteringDialog';

const INITIAL_FILTERS = getInitialPathFilters();
const WrappedDialog = () => {
    const [selectedFilters, setSelectedFilters] = useState<EdgeCheckboxType[]>(INITIAL_FILTERS);
    return (
        <EdgeFilteringDialog
            isOpen
            selectedFilters={selectedFilters}
            handleCancel={vi.fn}
            handleApply={vi.fn()}
            handleUpdate={(filters) => setSelectedFilters(filters)}
        />
    );
};

describe('Pathfinding', () => {
    beforeEach(async () => {
        await act(async () => {
            render(<WrappedDialog />);
        });
    });

    it('should render', async () => {
        expect(await screen.findByRole('dialog', { name: /Path Edge Filtering/i })).toBeInTheDocument();
        expect(
            screen.getByRole('heading', { name: /select the edge types to include in your pathfinding/i })
        ).toBeInTheDocument();

        const categoryAD = screen.getByRole('checkbox', { name: /active directory/i });
        expect(categoryAD).toBeInTheDocument();

        const categoryAzure = screen.getByRole('checkbox', { name: /azure/i });
        expect(categoryAzure).toBeInTheDocument();
    });

    it('checking a category updates all children and grandchildren', async () => {
        const user = userEvent.setup();

        // expand the category
        const expandCategoryButton = screen.getByRole('button', { name: 'expand-Active Directory' });
        await user.click(expandCategoryButton);

        // expand the subcategory
        const expandSubcategoryButton = screen.getByRole('button', {
            name: 'expand-Active Directory Structure',
        });
        await user.click(expandSubcategoryButton);

        // assert all subcategories underneath `Active Directory` category are `CHECKED`
        const activeDirectorySubcategories = AllEdgeTypes[0].subcategories;
        activeDirectorySubcategories.forEach((subcategory) => {
            const subcategoryElement = screen.getByRole('checkbox', { name: subcategory.name });
            expect(subcategoryElement).toBeChecked();
        });

        // BOOM! click the top level checkbox
        const categoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory' });
        await user.click(categoryCheckbox);

        // assert all subcategories underneath `Active Directory` are now `UNCHECKED`
        expect(categoryCheckbox).not.toBeChecked();
        activeDirectorySubcategories.forEach((subcategory) => {
            const subcategoryElement = screen.getByRole('checkbox', { name: subcategory.name });
            expect(subcategoryElement).not.toBeChecked();
        });

        // assert all edge types underneath first subcategory are now `UNCHECKED`.
        const edgeTypes = activeDirectorySubcategories[0].edgeTypes;
        edgeTypes.forEach((edgeType) => {
            const edgeTypeCheckbox = screen.getByRole('checkbox', { name: edgeType });
            expect(edgeTypeCheckbox).not.toBeChecked();
        });
    });

    it('checking a subcategory updates all children (edge types) in its group', async () => {
        const user = userEvent.setup();

        // expand category
        const expandCategoryButton = screen.getByRole('button', { name: 'expand-Active Directory' });
        await user.click(expandCategoryButton);

        // expand subcategory
        const expandSubcategoryButton = screen.getByRole('button', {
            name: 'expand-Active Directory Structure',
        });
        await user.click(expandSubcategoryButton);

        // assert that subcategory and all children are checked
        const subcategoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory Structure' });
        expect(subcategoryCheckbox).toBeChecked();
        const edgeTypes = AllEdgeTypes[0].subcategories[0].edgeTypes;
        edgeTypes.forEach((edgeType) => {
            const edgeTypeCheckbox = screen.getByRole('checkbox', { name: edgeType });
            expect(edgeTypeCheckbox).toBeChecked();
        });

        // uncheck subcategory
        await user.click(subcategoryCheckbox);

        // assert all edge types in subcategory are unchecked
        edgeTypes.forEach((edgeType) => {
            const edgeTypeCheckbox = screen.getByRole('checkbox', { name: edgeType });
            expect(edgeTypeCheckbox).not.toBeChecked();
        });
    });

    it('checking a single edge type updates its parent and grandparent', async () => {
        const user = userEvent.setup();

        // expand category
        const expandCategoryButton = screen.getByRole('button', { name: 'expand-Active Directory' });
        await user.click(expandCategoryButton);

        // expand subcategory
        const expandSubcategoryButton = screen.getByRole('button', {
            name: 'expand-Active Directory Structure',
        });
        await user.click(expandSubcategoryButton);

        // assert initial state: parent and grandparent elements are checked
        const categoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory' });
        const subcategoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory Structure' });
        expect(categoryCheckbox).toBeChecked();
        expect(subcategoryCheckbox).toBeChecked();

        // select an edge type and uncheck it
        const edgeTypeCheckbox = screen.getByRole('checkbox', { name: /contains/i });
        await user.click(edgeTypeCheckbox);

        // assert that parent and grandparent elements (subcategory and category) are now indeterminate
        expect(edgeTypeCheckbox).not.toBeChecked();
        expect(subcategoryCheckbox).toHaveAttribute('data-indeterminate', 'true');
        expect(categoryCheckbox).toHaveAttribute('data-indeterminate', 'true');
    });

    it('should render search field', async () => {
        const searchField = screen.getByPlaceholderText(/search edges/i);
        expect(searchField).toBeInTheDocument();
    });

    it('searching filters edge types and expands accordions automatically', async () => {
        const user = userEvent.setup();

        // get search field
        const searchField = screen.getByPlaceholderText(/search edges/i);

        // type in search query
        await user.type(searchField, 'Contains');

        // assert that `Contains` edge type is visible
        const containsEdgeCheckbox = screen.getByRole('checkbox', { name: 'Contains' });
        expect(containsEdgeCheckbox).toBeInTheDocument();

        // assert that accordions are expanded (minimize buttons are present)
        const minimizeCategoryButton = screen.getByRole('button', { name: 'minimize-Active Directory' });
        expect(minimizeCategoryButton).toBeInTheDocument();

        const minimizeSubcategoryButton = screen.getByRole('button', {
            name: 'minimize-Active Directory Structure',
        });
        expect(minimizeSubcategoryButton).toBeInTheDocument();
    });

    it('clearing search collapses accordions', async () => {
        const user = userEvent.setup();

        // get search field
        const searchField = screen.getByPlaceholderText(/search edges/i);

        // type in search query
        await user.type(searchField, 'Contains');

        // assert accordions are expanded
        expect(screen.getByRole('button', { name: 'minimize-Active Directory' })).toBeInTheDocument();

        // clear search field
        await user.clear(searchField);

        // assert accordions are collapsed (expand buttons are present)
        const expandCategoryButton = screen.getByRole('button', { name: 'expand-Active Directory' });
        expect(expandCategoryButton).toBeInTheDocument();
    });

    it('search only filters edge types, not categories or subcategories', async () => {
        const user = userEvent.setup();

        // get search field
        const searchField = screen.getByPlaceholderText(/search edges/i);

        // search for a category name that should not match
        await user.type(searchField, 'Active Directory');

        // assert that categories are still visible but only if they contain matching edge types
        // since "Active Directory" is not an edge type name, no categories should show
        const categoryCheckbox = screen.queryByRole('checkbox', { name: /active directory/i });
        expect(categoryCheckbox).not.toBeInTheDocument();
    });

    it('search is case insensitive', async () => {
        const user = userEvent.setup();

        // get search field
        const searchField = screen.getByPlaceholderText(/search edges/i);

        // type in lowercase search query
        await user.type(searchField, 'contains');

        // assert that edge type is found
        const containsEdgeCheckbox = screen.getByRole('checkbox', { name: 'Contains' });
        expect(containsEdgeCheckbox).toBeInTheDocument();

        // clear and try uppercase
        await user.clear(searchField);
        await user.type(searchField, 'CONTAINS');

        // assert that edge type is still found
        expect(screen.getByRole('checkbox', { name: 'Contains' })).toBeInTheDocument();
    });
});

describe('Pathfinding Search Persistence', () => {
    it('search field clears when dialog closes via cancel button', async () => {
        const user = userEvent.setup();
        const handleCancel = vi.fn();
        const handleApply = vi.fn();

        const { rerender } = render(
            <EdgeFilteringDialog
                isOpen={true}
                selectedFilters={INITIAL_FILTERS}
                handleCancel={handleCancel}
                handleApply={handleApply}
                handleUpdate={vi.fn()}
            />
        );

        // type in search field
        const searchField = screen.getByPlaceholderText(/search edges/i);
        await user.type(searchField, 'Contains');
        expect(searchField).toHaveValue('Contains');

        // close dialog
        const cancelButton = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancelButton);

        // reopen dialog
        rerender(
            <EdgeFilteringDialog
                isOpen={false}
                selectedFilters={INITIAL_FILTERS}
                handleCancel={handleCancel}
                handleApply={handleApply}
                handleUpdate={vi.fn()}
            />
        );

        await act(async () => {
            rerender(
                <EdgeFilteringDialog
                    isOpen={true}
                    selectedFilters={INITIAL_FILTERS}
                    handleCancel={handleCancel}
                    handleApply={handleApply}
                    handleUpdate={vi.fn()}
                />
            );
        });

        // assert search field is empty
        const newSearchField = screen.getByPlaceholderText(/search edges/i);
        expect(newSearchField).toHaveValue('');
    });

    it('search field clears when dialog closes via apply button', async () => {
        const user = userEvent.setup();
        const handleCancel = vi.fn();
        const handleApply = vi.fn();

        const { rerender } = render(
            <EdgeFilteringDialog
                isOpen={true}
                selectedFilters={INITIAL_FILTERS}
                handleCancel={handleCancel}
                handleApply={handleApply}
                handleUpdate={vi.fn()}
            />
        );

        // type in search field
        const searchField = screen.getByPlaceholderText(/search edges/i);
        await user.type(searchField, 'Contains');
        expect(searchField).toHaveValue('Contains');

        // close dialog via apply
        const applyButton = screen.getByRole('button', { name: /apply/i });
        await user.click(applyButton);

        // reopen dialog
        rerender(
            <EdgeFilteringDialog
                isOpen={false}
                selectedFilters={INITIAL_FILTERS}
                handleCancel={handleCancel}
                handleApply={handleApply}
                handleUpdate={vi.fn()}
            />
        );

        await act(async () => {
            rerender(
                <EdgeFilteringDialog
                    isOpen={true}
                    selectedFilters={INITIAL_FILTERS}
                    handleCancel={handleCancel}
                    handleApply={handleApply}
                    handleUpdate={vi.fn()}
                />
            );
        });

        // assert search field is empty
        const newSearchField = screen.getByPlaceholderText(/search edges/i);
        expect(newSearchField).toHaveValue('');
    });
});
