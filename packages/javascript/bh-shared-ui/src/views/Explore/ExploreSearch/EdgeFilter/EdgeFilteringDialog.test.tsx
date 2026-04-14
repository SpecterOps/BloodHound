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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useState } from 'react';
import { INITIAL_FILTERS } from '../../../../hooks/useExploreGraph/queries';
import { act, render, screen } from '../../../../test-utils';
import EdgeFilteringDialog from './EdgeFilteringDialog';
import { BUILTIN_EDGE_CATEGORIES, EdgeCheckboxType } from './edgeCategories';

const server = setupServer(
    rest.get('/api/v2/features', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const WrappedDialog = () => {
    const [selectedFilters, setSelectedFilters] = useState<EdgeCheckboxType[]>(INITIAL_FILTERS);
    return (
        <EdgeFilteringDialog
            isOpen
            selectedFilters={selectedFilters}
            handleCancel={vi.fn()}
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
        expect(screen.getByText(/select the edge types to include in your pathfinding/i)).toBeInTheDocument();

        const categoryAD = screen.getByRole('checkbox', { name: 'Active Directory' });
        expect(categoryAD).toBeInTheDocument();

        const categoryAzure = screen.getByRole('checkbox', { name: 'Azure' });
        expect(categoryAzure).toBeInTheDocument();
    });

    it('should render search input', async () => {
        const searchInput = screen.getByPlaceholderText('Search edges...');
        expect(searchInput).toBeInTheDocument();
    });

    it('should render expand all and collapse all buttons', async () => {
        expect(screen.getByRole('button', { name: /Expand All/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Collapse All/i })).toBeInTheDocument();
    });

    it('categories default to expanded', async () => {
        // categories start expanded, so minimize buttons should be visible
        expect(screen.getByRole('button', { name: 'minimize-Active Directory' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'minimize-Azure' })).toBeInTheDocument();

        // subcategories should also be visible since they default expanded
        const subcategoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory Structure' });
        expect(subcategoryCheckbox).toBeInTheDocument();
    });

    it('checking a category updates all children and grandchildren', async () => {
        const user = userEvent.setup();

        // all subcategories and edge types are already visible since sections default expanded
        const activeDirectorySubcategories = BUILTIN_EDGE_CATEGORIES[0].subcategories;
        activeDirectorySubcategories.forEach((subcategory) => {
            const matches = screen.getAllByRole('checkbox', { name: subcategory.name });
            expect(matches[0]).toBeChecked();
        });

        // click the top level checkbox
        const categoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory' });
        await user.click(categoryCheckbox);

        // assert all subcategories underneath `Active Directory` are now `UNCHECKED`
        expect(categoryCheckbox).not.toBeChecked();
        activeDirectorySubcategories.forEach((subcategory) => {
            const matches = screen.getAllByRole('checkbox', { name: subcategory.name });
            expect(matches[0]).not.toBeChecked();
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

        // assert that subcategory and all children are checked (already visible)
        const subcategoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory Structure' });
        expect(subcategoryCheckbox).toBeChecked();
        const edgeTypes = BUILTIN_EDGE_CATEGORIES[0].subcategories[0].edgeTypes;
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

        // assert initial state: parent and grandparent elements are checked (already visible)
        const categoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory' });
        const subcategoryCheckbox = screen.getByRole('checkbox', { name: 'Active Directory Structure' });
        expect(categoryCheckbox).toBeChecked();
        expect(subcategoryCheckbox).toBeChecked();

        // select an edge type and uncheck it
        const edgeTypeCheckbox = screen.getByRole('checkbox', { name: 'Contains' });
        await user.click(edgeTypeCheckbox);

        // assert that parent and grandparent elements (subcategory and category) are now indeterminate
        expect(edgeTypeCheckbox).not.toBeChecked();
        expect(subcategoryCheckbox).toHaveAttribute('data-state', 'indeterminate');
        expect(categoryCheckbox).toHaveAttribute('data-state', 'indeterminate');
    });

    it('collapse all button collapses all sections', async () => {
        const user = userEvent.setup();

        // subcategories should be visible initially
        expect(screen.getByRole('checkbox', { name: 'Active Directory Structure' })).toBeInTheDocument();

        // click Collapse All
        await user.click(screen.getByRole('button', { name: /Collapse All/i }));

        // subcategories should no longer be visible
        expect(screen.queryByRole('checkbox', { name: 'Active Directory Structure' })).not.toBeInTheDocument();

        // category expand buttons should now show expand instead of minimize
        expect(screen.getByRole('button', { name: 'expand-Active Directory' })).toBeInTheDocument();
    });

    it('expand all button expands all sections after collapsing', async () => {
        const user = userEvent.setup();

        // collapse all first
        await user.click(screen.getByRole('button', { name: /Collapse All/i }));
        expect(screen.queryByRole('checkbox', { name: 'Active Directory Structure' })).not.toBeInTheDocument();

        // click Expand All
        await user.click(screen.getByRole('button', { name: /Expand All/i }));

        // subcategories should be visible again
        expect(screen.getByRole('checkbox', { name: 'Active Directory Structure' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'minimize-Active Directory' })).toBeInTheDocument();
    });

    it('search filters edges by name', async () => {
        const user = userEvent.setup();

        const searchInput = screen.getByPlaceholderText('Search edges...');
        await user.type(searchInput, 'Contains');

        // matching edge should be visible
        expect(screen.getByRole('checkbox', { name: /^Contains$/i })).toBeInTheDocument();

        // non-matching edges should not be visible
        expect(screen.queryByRole('checkbox', { name: 'AdminTo' })).not.toBeInTheDocument();
    });

    it('search highlights matching substring with text-link class', async () => {
        const user = userEvent.setup();

        const searchInput = screen.getByPlaceholderText('Search edges...');
        await user.type(searchInput, 'Contains');

        const highlightedSpan = document.querySelector('.text-link');
        expect(highlightedSpan).toBeInTheDocument();
        expect(highlightedSpan?.textContent).toBe('Contains');
    });

    it('search forces sections to expand when there is a match', async () => {
        const user = userEvent.setup();

        // collapse all first
        await user.click(screen.getByRole('button', { name: /Collapse All/i }));
        expect(screen.queryByRole('checkbox', { name: 'Active Directory Structure' })).not.toBeInTheDocument();

        // search for an edge type
        const searchInput = screen.getByPlaceholderText('Search edges...');
        await user.type(searchInput, 'Contains');

        // matching category and subcategory should be force-expanded
        expect(screen.getByRole('checkbox', { name: /^Contains$/i })).toBeInTheDocument();
    });

    it('expand/collapse all buttons are disabled when searching', async () => {
        const user = userEvent.setup();

        const expandAllButton = screen.getByRole('button', { name: /Expand All/i });
        const collapseAllButton = screen.getByRole('button', { name: /Collapse All/i });

        expect(expandAllButton).not.toBeDisabled();
        expect(collapseAllButton).not.toBeDisabled();

        const searchInput = screen.getByPlaceholderText('Search edges...');
        await user.type(searchInput, 'test');

        expect(expandAllButton).toBeDisabled();
        expect(collapseAllButton).toBeDisabled();
    });

    it('edges are displayed in alphabetical order', async () => {
        // get all edge type checkboxes under the first subcategory
        const edgeTypes = BUILTIN_EDGE_CATEGORIES[0].subcategories[0].edgeTypes;
        const sortedEdgeTypes = [...edgeTypes].sort((a, b) => a.localeCompare(b));

        // get all checkboxes matching the edge types
        const edgeCheckboxes = sortedEdgeTypes.map((edgeType) => screen.getByRole('checkbox', { name: edgeType }));

        // verify they all exist and are in sorted order by checking DOM position
        for (let i = 0; i < edgeCheckboxes.length - 1; i++) {
            const position = edgeCheckboxes[i].compareDocumentPosition(edgeCheckboxes[i + 1]);
            // Node.DOCUMENT_POSITION_FOLLOWING = 4
            expect(position & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
        }
    });
});
