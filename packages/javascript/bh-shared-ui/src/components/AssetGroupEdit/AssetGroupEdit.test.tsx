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

import { setupServer } from 'msw/node';
import { createMockAssetGroup, createMockAssetGroupMembers, createMockSearchResults } from '../../mocks/factories';
import { act, render, waitFor } from '../../test-utils';
import { AUTOCOMPLETE_PLACEHOLDER } from './AssetGroupAutocomplete';
import AssetGroupEdit from './AssetGroupEdit';
import { rest } from 'msw';
import userEvent from '@testing-library/user-event';

const assetGroup = createMockAssetGroup();
const assetGroupMembers = createMockAssetGroupMembers();
const searchResults = createMockSearchResults();

const server = setupServer(
    rest.get('/api/v2/asset-groups/1/members', (req, res, ctx) => {
        return res(
            ctx.json({
                count: assetGroupMembers.members.length,
                limit: 100,
                skip: 0,
                data: assetGroupMembers,
            })
        );
    }),
    rest.get('/api/v2/search', (req, res, ctx) => {
        return res(
            ctx.json({
                data: searchResults,
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('AssetGroupEdit', () => {
    const setup = async () => {
        const user = userEvent.setup();
        const screen = await act(async () => {
            return render(<AssetGroupEdit assetGroup={assetGroup} filter={{}} makeNodeFilterable={() => ({})} />);
        });
        return { user, screen };
    };

    it('should display a searchbox with a placeholder when rendered', async () => {
        const { screen } = await setup();
        const input = screen.getByPlaceholderText(AUTOCOMPLETE_PLACEHOLDER);
        expect(input).toBeInTheDocument();
    });

    it('should display a total count of asset group members', async () => {
        const { screen } = await setup();
        const count = screen.getByText('Total Count').nextSibling.textContent;
        expect(count).toBe(assetGroupMembers.members.length.toString());
    });

    it('should display search results when the user enters text', async () => {
        const { screen, user } = await setup();
        const input = screen.getByRole('combobox');

        await user.type(input, 'test');
        expect(input.value).toEqual('test');

        const result = await waitFor(() => screen.getByText('00001.TESTLAB.LOCAL'));
        expect(result).toBeInTheDocument();
    });

    it('should add an option and display the changelog when it is clicked', async () => {
        const { screen, user } = await setup();
        const selection = searchResults[0];

        const input = screen.getByRole('combobox');
        await user.type(input, 'test');
        expect(input.value).toEqual('test');

        const result = await waitFor(() => screen.getByText(selection.name));
        await user.click(result);
        await user.keyboard('{Escape}');

        expect(screen.getByText(selection.name)).toBeInTheDocument();
        expect(screen.getByText(selection.objectid)).toBeInTheDocument();
        expect(screen.getByText('Cancel')).toBeInTheDocument();
        expect(screen.getByText('Confirm Changes')).toBeInTheDocument();
    });

    it('should remove the option from the changelog when the corresponding button is clicked', async () => {
        const { screen, user } = await setup();
        const selection = searchResults[0];

        const input = screen.getByRole('combobox');
        await user.type(input, 'test');
        expect(input.value).toEqual('test');

        const result = await waitFor(() => screen.getByText(selection.name));
        await user.click(result);
        await user.keyboard('{Escape}');

        const removeButton = screen.getByText('xmark');
        await user.click(removeButton);

        expect(screen.queryByText(selection.name)).toBeNull();
        expect(screen.queryByText(selection.objectid)).toBeNull();
    });
});
