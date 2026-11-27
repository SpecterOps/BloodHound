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
import ExploreSearchCombobox from '.';
import { ActiveDirectoryNodeKind, AzureNodeKind, CommonNodeKind } from '../../graphSchema';
import { mockKindsHandler } from '../../mocks';
import { mockGetConfigurationHandler } from '../../mocks/handlers';
import { act, render, screen, within } from '../../test-utils';

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
    }),
    rest.get(`/api/v2/custom-nodes`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    mockKindsHandler(),
    mockGetConfigurationHandler()
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ExploreSearchCombobox', () => {
    it('can render', async () => {
        const labelText: string = 'test label';
        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue=''
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={null}
                />
            );
        });
        expect(screen.getByLabelText(labelText)).toBeInTheDocument();
    });

    it('when an `inputValue` is provided it eventually displays a list of search results', async () => {
        const user = userEvent.setup();
        const labelText: string = 'test label';

        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue='a'
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={null}
                />
            );
        });

        await user.click(screen.getByLabelText(labelText));
        const options = await screen.findAllByRole('option');

        expect(options).toHaveLength(testSearchResults.data.length);
        for (let i = 0; i < testSearchResults.data.length; i++) {
            expect(options[i]).toHaveTextContent(testSearchResults.data[i].name);
            within(options[i]).getByTitle(testSearchResults.data[i].type);
        }
    });

    it('when a search result is clicked it calls `handleNodeSelected`', async () => {
        const user = userEvent.setup();
        const labelText: string = 'test label';
        const handleNodeSelected = vi.fn();

        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue='a'
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={handleNodeSelected}
                    selectedItem={null}
                />
            );
        });

        await user.type(screen.getByLabelText(labelText), 'admin');
        const options = await screen.findAllByRole('option');
        await user.click(options[0]);

        expect(handleNodeSelected).toHaveBeenCalledTimes(1);
        expect(handleNodeSelected).toHaveBeenCalledWith(testSearchResults.data[0]);
    });
});

describe('icon rendering', () => {
    const labelText: string = 'test label';

    it('when `selectedItem` is provided, the combobox displays the icon', async () => {
        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue=''
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={{ type: ActiveDirectoryNodeKind.Computer, objectid: '1', name: 'Computer a' }}
                />
            );
        });

        const input = screen.getByLabelText(labelText);
        expect(input).toHaveClass('MuiInputBase-inputAdornedStart');
    });

    it('when `selectedItem` is null, the combobox does not display an icon', async () => {
        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue=''
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={null}
                />
            );
        });

        const input = screen.getByLabelText(labelText);
        expect(input).not.toHaveClass('MuiInputBase-inputAdornedStart');
    });
});

describe('ExploreSearchCombobox result metadata', () => {
    const searchResults = {
        data: [
            {
                name: 'ADMIN@TESTLAB.LOCAL',
                objectid: '1',
                type: 'User',
                distinguishedname: 'CN=Admin,OU=Users,DC=testlab,DC=local',
            },
            {
                name: 'ADMIN@TESTLAB.LOCAL',
                objectid: '2',
                type: 'Group',
                distinguishedname: 'CN=Admin,OU=Groups,DC=testlab,DC=local',
            },
            {
                name: 'UNIQUE@TESTLAB.LOCAL',
                objectid: '3',
                type: 'Computer',
                distinguishedname: 'CN=Unique,OU=Computers,DC=testlab,DC=local',
            },
            {
                name: 'SUPPORT@TESTLAB.LOCAL',
                objectid: '4',
                type: ActiveDirectoryNodeKind.OU,
                distinguishedname: 'OU=Support,OU=Corporate,DC=testlab,DC=local',
            },
            { name: 'LAB@TESTLAB.LOCAL', objectid: '5', type: ActiveDirectoryNodeKind.OU, distinguishedname: '' },
            { name: 'Cloud Admin', objectid: '6', type: AzureNodeKind.User },
            { name: 'Example Repo', objectid: '7', type: 'GitHubRepository' },
            { name: 'Tier Zero', objectid: '8', type: CommonNodeKind.Zone },
            { name: 'Unclassified node', objectid: '9', type: 'Unknown' },
        ],
    };

    beforeEach(() => {
        server.use(
            rest.get(`/api/v2/search`, (req, res, ctx) => {
                return res(ctx.json(searchResults));
            })
        );
    });

    it('shows platform and type for each result and appends available DNs for OUs and duplicate names', async () => {
        const user = userEvent.setup();
        const labelText: string = 'test label';

        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue='admin'
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={null}
                />
            );
        });

        await user.click(screen.getByLabelText(labelText));
        await screen.findAllByRole('option');

        expect(screen.getByText('Active Directory | User | CN=Admin,OU=Users,DC=testlab,DC=local')).toBeVisible();
        expect(screen.getByText('Active Directory | Group | CN=Admin,OU=Groups,DC=testlab,DC=local')).toBeVisible();
        expect(screen.getByText('Active Directory | OU | OU=Support,OU=Corporate,DC=testlab,DC=local')).toBeVisible();
        expect(
            within(screen.getByRole('option', { name: /LAB@TESTLAB\.LOCAL/ })).getByText('Active Directory | OU')
        ).toBeVisible();
        expect(
            within(screen.getByRole('option', { name: /UNIQUE@TESTLAB\.LOCAL/ })).getByText(
                'Active Directory | Computer'
            )
        ).toBeVisible();
        expect(within(screen.getByRole('option', { name: /Cloud Admin/ })).getByText('Azure | AZUser')).toBeVisible();
        expect(
            within(screen.getByRole('option', { name: /Example Repo/ })).getByText('OpenGraph | GitHubRepository')
        ).toBeVisible();
        expect(within(screen.getByRole('option', { name: /Tier Zero/ })).getByText('BloodHound | Zone')).toBeVisible();
        expect(
            within(screen.getByRole('option', { name: /Unclassified node/ })).getByText('Unknown platform | Unknown')
        ).toBeVisible();
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

        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue={searchText}
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={null}
                />
            );
        });

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

        await act(async () => {
            render(
                <ExploreSearchCombobox
                    labelText={labelText}
                    inputValue={searchText}
                    handleNodeEdited={vi.fn()}
                    handleNodeSelected={vi.fn()}
                    selectedItem={null}
                />
            );
        });

        await user.type(screen.getByLabelText(labelText), searchText);

        expect(await screen.findByText(`Search has timed out. Please try again.`)).toBeInTheDocument();
    });
});
