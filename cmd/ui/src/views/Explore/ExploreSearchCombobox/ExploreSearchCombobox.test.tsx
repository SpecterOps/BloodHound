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
import { act, render, screen, within } from 'src/test-utils';
import ExploreSearchCombobox from '.';
import * as actions from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH } from 'src/ducks/searchbar/types';

const testSearchResults = {
    data: [
        {
            name: 'admin1@testlab.local',
            objectid: '1',
            type: 'User',
        },
        {
            name: 'admin2@testlab.local',
            objectid: '2',
            type: 'Group',
        },
        {
            name: 'admin3@testlab.local',
            objectid: '3',
            type: 'Computer',
        },
    ],
};

const server = setupServer(
    rest.get(`/api/v2/search`, (req, res, ctx) => {
        return res(ctx.json(testSearchResults));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ExploreSearchCombobox', () => {
    it('can render', async () => {
        const labelText: string = 'test label';
        await act(async () => {
            render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);
        });
        expect(screen.getByLabelText(labelText)).toBeInTheDocument();
    });

    it('typing a new search query calls onInputValueChange', async () => {
        const user = userEvent.setup();
        const spy = jest.spyOn(actions, 'sourceNodeEdited');
        const labelText: string = 'test label';

        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        const testQuery = 'admin';
        await user.type(screen.getByLabelText(labelText), testQuery);

        expect(spy).toHaveBeenLastCalledWith(testQuery);
    });

    it('when a search query is provided it eventually displays a list of search results', async () => {
        const user = userEvent.setup();
        const labelText: string = 'test label';

        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />, {
            initialState: {
                search: {
                    primary: {
                        searchTerm: ' ',
                    },
                },
            },
        });

        await user.click(screen.getByLabelText(labelText));
        const options = await screen.findAllByRole('option');

        expect(options).toHaveLength(testSearchResults.data.length);
        for (let i = 0; i < testSearchResults.data.length; i++) {
            expect(options[i]).toHaveTextContent(testSearchResults.data[i].name);
            within(options[i]).getByTitle(testSearchResults.data[i].type);
        }
    });

    it('when a search result is clicked it calls onSelectedItemChange', async () => {
        const user = userEvent.setup();
        const spy = vi.spyOn(actions, 'sourceNodeSelected');
        const labelText: string = 'test label';

        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        await user.type(screen.getByLabelText(labelText), 'admin');
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        expect(spy).toHaveBeenCalledTimes(1);
        expect(spy).toHaveBeenCalledWith(testSearchResults.data[0]);
    });
});

describe('icon rendering', () => {
    const labelText: string = 'test label';
    const userProvidedInput = 'admin';

    it('when a search result is clicked, the combobox displays the icon', async () => {
        const user = userEvent.setup();

        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        await user.type(screen.getByLabelText(labelText), userProvidedInput);
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        const input = screen.getByLabelText(labelText);
        expect(input).toHaveClass('MuiInputBase-inputAdornedStart');
    });

    it('when a search result is selected with enter key, the combobox displays the icon', async () => {
        const user = userEvent.setup();
        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        await user.type(screen.getByLabelText(labelText), userProvidedInput);
        const options = await screen.findAllByRole('option');
        await user.type(options[0], '{enter}');

        const input = screen.getByLabelText(labelText);
        expect(input).toHaveClass('MuiInputBase-inputAdornedStart');
    });

    it('when a search result is cleared, the combobox removes the icon', async () => {
        const user = userEvent.setup();
        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        await user.type(screen.getByLabelText(labelText), userProvidedInput);
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        const input = screen.getByLabelText(labelText);
        expect(input).toHaveClass('MuiInputBase-inputAdornedStart');

        await user.clear(input);
        expect(input).not.toHaveClass('MuiInputBase-inputAdornedStart');
    });
});

describe('ExploreSearchCombobox with null response', () => {
    beforeEach(() => {
        server.use(
            rest.get(`/api/v2/search`, (req, res, ctx) => {
                return res(ctx.json({ data: null }));
            })
        );
    });

    it('a null response from the server is handled', async () => {
        const user = userEvent.setup();
        const labelText: string = 'test label';
        const searchText: string = 'blah';

        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        await user.type(screen.getByLabelText(labelText), searchText);

        expect(await screen.findByText(`No results found for "blah"`)).toBeInTheDocument();
    });
});

describe('ExploreSearchCombobox with search timeout', () => {
    beforeEach(() => {
        console.error = vi.fn();
        server.use(
            rest.get(`/api/v2/search`, (req, res, ctx) => {
                return res(ctx.status(504));
            })
        );
    });

    it('a timeout response from the server is handled', async () => {
        const user = userEvent.setup();
        const labelText: string = 'test label';
        const searchText: string = 'blah';

        render(<ExploreSearchCombobox labelText={labelText} searchType={PRIMARY_SEARCH} />);

        await user.type(screen.getByLabelText(labelText), searchText);

        expect(await screen.findByText(`Search has timed out. Please try again.`)).toBeInTheDocument();
    });
});
