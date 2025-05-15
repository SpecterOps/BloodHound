// Copyright 2025 Specter Ops, Inc.
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
import { SeedTypeObjectId } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen, waitFor } from '../../test-utils';
import SelectorFormContext, { initialValue } from '../../views/TierManagement/Save/SelectorForm/SelectorFormContext';
import AssetGroupSelectorObjectSelect from './AssetGroupSelectorObjectSelect';

const testNodes = [
    {
        name: 'foo',
        objectid: '2',
        type: 'User',
    },
];
const testSearchResults = {
    data: testNodes,
};

const server = setupServer(
    rest.get(`/api/v2/search`, (_, res, ctx) => {
        return res(ctx.json(testSearchResults));
    }),
    rest.post(`/api/v2/asset-group-tags/preview-selectors`, (_, res, ctx) => {
        return res(ctx.json({ data: { members: testNodes } }));
    }),
    rest.post(`/api/v2/graphs/cypher`, (_, res, ctx) => {
        return res(ctx.json({ data: { nodes: testNodes } }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('AssetGroupSelectorObjectSelect', () => {
    const user = userEvent.setup();
    const seeds = [
        {
            value: '1',
            type: SeedTypeObjectId,
            selector_id: 1,
        },
    ];

    const setSeeds = vi.fn();

    beforeEach(async () => {
        await act(async () => {
            render(
                <SelectorFormContext.Provider value={initialValue}>
                    <AssetGroupSelectorObjectSelect seeds={seeds} />
                </SelectorFormContext.Provider>
            );
        });
    });

    it('should render', async () => {
        expect(await screen.findByTestId('explore_search_input-search')).toBeInTheDocument();
        expect(screen.getByText('Object Selector')).toBeInTheDocument();
        expect(screen.getByText('Use the input field to add objects to the list')).toBeInTheDocument();
    });

    it('invokes setSeeds when a current seed is deleted', async () => {
        const deleteBtn = await screen.findByText('trash-can');

        await user.click(deleteBtn);

        waitFor(() => {
            expect(setSeeds).toHaveBeenCalledWith([]);
        });
    });

    it.skip('invokes setSeeds when a new seed is selected', async () => {
        await screen.findByTestId('explore_search_input-search');

        const input = screen.getByLabelText('Search Objects To Add');

        user.type(input, 'foo');

        const options = await screen.findAllByRole('option');

        await user.click(options[0]);

        expect(await screen.findByText('user')).toBeInTheDocument();
        expect(await screen.findByText('foo')).toBeInTheDocument();

        waitFor(() => {
            expect(setSeeds).toHaveBeenCalledWith([...seeds, { type: SeedTypeObjectId, value: '2' }]);
        });
    });
});
