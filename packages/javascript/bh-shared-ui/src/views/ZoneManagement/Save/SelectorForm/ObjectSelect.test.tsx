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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen, waitFor } from '../../../../test-utils';
import ObjectSelect from './ObjectSelect';
import SelectorFormContext, { initialValue } from './SelectorFormContext';

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
    }),
    rest.get(`/api/v2/customnode`, async (_req, res, ctx) => {
        return res(ctx.json({ data: [] }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('AssetGroupTagsSelectorObjectSelect', () => {
    const user = userEvent.setup();

    const dispatch = vi.fn();

    beforeEach(async () => {
        await act(async () => {
            render(
                <SelectorFormContext.Provider
                    value={{
                        ...initialValue,
                        selectedObjects: [
                            {
                                objectid: '1',
                                type: 'User',
                                name: 'Bob',
                            },
                        ],
                        dispatch,
                    }}>
                    <ObjectSelect />
                </SelectorFormContext.Provider>
            );
        });
    });

    it('should render', async () => {
        expect(await screen.findByTestId('explore_search_input-search')).toBeInTheDocument();
        expect(screen.getByText('Object Selector')).toBeInTheDocument();
        expect(screen.getByText('Use the input field to add objects to the list')).toBeInTheDocument();
    });

    it('dispatches an action to the delete the associated node', async () => {
        const deleteBtn = await screen.findByText('trash-can');

        await user.click(deleteBtn);

        waitFor(() => {
            expect(dispatch).toHaveBeenCalledWith({
                type: 'remove-selected-object',
                node: { objectid: '1', type: 'User', name: 'Bob' },
            });
        });
    });

    it('dispatches an action to add the associated node', async () => {
        await screen.findByTestId('explore_search_input-search');

        const input = screen.getByLabelText('Search Objects To Add');

        user.type(input, 'foo');

        const options = await screen.findAllByRole('option');

        await user.click(options[0]);

        waitFor(() => {
            expect(dispatch).toHaveBeenCalledWith({
                type: 'add-selected-object',
                node: { objectid: '2', name: 'foo', type: 'User' },
            });

            expect(screen.getByText('user')).toBeInTheDocument();
            expect(screen.getByText('foo')).toBeInTheDocument();
        });
    });
});
