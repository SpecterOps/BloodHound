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
        const activeDirectorySubcategories = BUILTIN_EDGE_CATEGORIES[0].subcategories;
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
});
