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
        const testOnInputValueChange: any = vi.fn();
        const testOnSelectedItemChange: any = vi.fn();
        const testLabelText: string = 'test label';
        await act(async () => {
            render(
                <ExploreSearchCombobox
                    inputValue=''
                    onInputValueChange={testOnInputValueChange}
                    selectedItem={null}
                    onSelectedItemChange={testOnSelectedItemChange}
                    labelText={testLabelText}
                />
            );
        });
        expect(screen.getByLabelText(testLabelText)).toBeInTheDocument();
    });

    it('typing a new search query calls onInputValueChange', async () => {
        const user = userEvent.setup();
        const testOnInputValueChange: any = vi.fn();
        const testOnSelectedItemChange: any = vi.fn();
        const testLabelText: string = 'test label';
        render(
            <ExploreSearchCombobox
                inputValue=''
                onInputValueChange={testOnInputValueChange}
                selectedItem={null}
                onSelectedItemChange={testOnSelectedItemChange}
                labelText={testLabelText}
            />
        );

        const testQuery = 'admin';
        await user.type(screen.getByLabelText(testLabelText), testQuery);
        expect(testOnInputValueChange).toHaveBeenLastCalledWith({
            highlightedIndex: -1,
            inputValue: testQuery,
            isOpen: true,
            selectedItem: null,
            type: '__input_change__',
        });
    });

    it('when a search query is provided it eventually displays a list of search results', async () => {
        const user = userEvent.setup();
        const testOnInputValueChange: any = vi.fn();
        const testOnSelectedItemChange: any = vi.fn();
        const testLabelText: string = 'test label';
        const testInputValue = 'admin';
        render(
            <ExploreSearchCombobox
                inputValue={testInputValue}
                onInputValueChange={testOnInputValueChange}
                selectedItem={null}
                onSelectedItemChange={testOnSelectedItemChange}
                labelText={testLabelText}
            />
        );

        await user.click(screen.getByLabelText(testLabelText));
        const options = await screen.findAllByRole('option');
        expect(options).toHaveLength(testSearchResults.data.length);
        for (let i = 0; i < testSearchResults.data.length; i++) {
            expect(options[i]).toHaveTextContent(testSearchResults.data[i].name);
            within(options[i]).getByTitle(testSearchResults.data[i].type);
        }
    });

    it('when a search result is clicked it calls onSelectedItemChange', async () => {
        const user = userEvent.setup();
        const testOnInputValueChange: any = vi.fn();
        const testOnSelectedItemChange: any = vi.fn();
        const testLabelText: string = 'test label';
        const testInputValue = 'admin';
        render(
            <ExploreSearchCombobox
                inputValue={testInputValue}
                onInputValueChange={testOnInputValueChange}
                selectedItem={null}
                onSelectedItemChange={testOnSelectedItemChange}
                labelText={testLabelText}
            />
        );

        await user.click(screen.getByLabelText(testLabelText));
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        expect(testOnSelectedItemChange).toHaveBeenCalledWith({
            highlightedIndex: -1,
            inputValue: 'admin1@testlab.local',
            isOpen: false,
            selectedItem: testSearchResults.data[0],
            type: '__item_click__',
        });
    });
});

describe('icon rendering', () => {
    const labelText: string = 'test label';
    const userProvidedInput = 'admin';

    it('when a search result is clicked, the combobox displays the icon', async () => {
        const user = userEvent.setup();

        let selectedItem: any = null;
        const onSelectedItemChange: any = vi.fn((item) => {
            selectedItem = item.selectedItem;
        });

        const onInputValueChange: any = vi.fn();

        const { rerender } = render(
            <ExploreSearchCombobox
                inputValue={userProvidedInput}
                onInputValueChange={onInputValueChange}
                selectedItem={selectedItem}
                onSelectedItemChange={onSelectedItemChange}
                labelText={labelText}
            />
        );

        await user.click(screen.getByLabelText(labelText));
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        const input = screen.getByLabelText(labelText);
        expect(input).not.toHaveClass('MuiInputBase-inputAdornedStart');

        rerender(
            <ExploreSearchCombobox
                inputValue={userProvidedInput}
                onInputValueChange={onInputValueChange}
                selectedItem={selectedItem}
                onSelectedItemChange={onSelectedItemChange}
                labelText={labelText}
            />
        );

        expect(input).toHaveClass('MuiInputBase-inputAdornedStart');
    });

    it('when a search result is selected with enter key, the combobox displays the icon', async () => {
        const user = userEvent.setup();

        let selectedItem: any = null;
        const onSelectedItemChange: any = vi.fn((item) => {
            selectedItem = item.selectedItem;
        });

        const onInputValueChange: any = vi.fn();

        const { rerender } = render(
            <ExploreSearchCombobox
                inputValue={userProvidedInput}
                onInputValueChange={onInputValueChange}
                selectedItem={selectedItem}
                onSelectedItemChange={onSelectedItemChange}
                labelText={labelText}
            />
        );

        await user.click(screen.getByLabelText(labelText));
        const options = await screen.findAllByRole('option');
        await user.type(options[0], '{enter}');

        const input = screen.getByLabelText(labelText);
        expect(input).not.toHaveClass('MuiInputBase-inputAdornedStart');

        rerender(
            <ExploreSearchCombobox
                inputValue={userProvidedInput}
                onInputValueChange={onInputValueChange}
                selectedItem={selectedItem}
                onSelectedItemChange={onSelectedItemChange}
                labelText={labelText}
            />
        );

        expect(input).toHaveClass('MuiInputBase-inputAdornedStart');
    });

    it('when a search result is cleared, the combobox removes the icon', async () => {
        const user = userEvent.setup();

        let selectedItem: any = null;
        const onSelectedItemChange: any = vi.fn((item) => {
            selectedItem = item.selectedItem;
        });

        const onInputValueChange: any = vi.fn();

        const { rerender } = render(
            <ExploreSearchCombobox
                inputValue={userProvidedInput}
                onInputValueChange={onInputValueChange}
                selectedItem={selectedItem}
                onSelectedItemChange={onSelectedItemChange}
                labelText={labelText}
            />
        );

        await user.click(screen.getByLabelText(labelText));
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        const input = screen.getByLabelText(labelText);
        expect(input).not.toHaveClass('MuiInputBase-inputAdornedStart');

        rerender(
            <ExploreSearchCombobox
                inputValue={userProvidedInput}
                onInputValueChange={onInputValueChange}
                selectedItem={selectedItem}
                onSelectedItemChange={onSelectedItemChange}
                labelText={labelText}
            />
        );

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
        const testOnInputValueChange: any = vi.fn();
        const testOnSelectedItemChange: any = vi.fn();
        const testLabelText: string = 'test label';
        const nullResultSearchText: string = 'the search result is null';
        render(
            <ExploreSearchCombobox
                inputValue={nullResultSearchText}
                onInputValueChange={testOnInputValueChange}
                selectedItem={null}
                onSelectedItemChange={testOnSelectedItemChange}
                labelText={testLabelText}
            />
        );

        expect(await screen.findByText(`No results found for "the search result is null"`)).toBeInTheDocument();
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
        const testOnInputValueChange: any = vi.fn();
        const testOnSelectedItemChange: any = vi.fn();
        const testLabelText: string = 'test label';
        const inputValue: string = 'takesLongToFind';
        render(
            <ExploreSearchCombobox
                inputValue={inputValue}
                onInputValueChange={testOnInputValueChange}
                selectedItem={null}
                onSelectedItemChange={testOnSelectedItemChange}
                labelText={testLabelText}
            />
        );

        expect(await screen.findByText(`Search has timed out. Please try again.`)).toBeInTheDocument();
    });
});
