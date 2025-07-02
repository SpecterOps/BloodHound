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
import * as bhSharedUi from 'bh-shared-ui';
import { DeepPartial, PathfindingFilters, Permission, createAuthStateWithPermissions } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act } from 'react-dom/test-utils';
import { AppState } from 'src/store';
import { render, screen, waitFor } from 'src/test-utils';
import ContextMenu from './ContextMenu';

const mockUseExploreParams = vi.spyOn(bhSharedUi, 'useExploreParams');
const mockSelectedItemQuery = vi.spyOn(bhSharedUi, 'useExploreSelectedItem');

const fakeSelectedItemId = 'abc';
const fakeSelectedNode = {
    selectedItemQuery: {
        data: {
            objectId: fakeSelectedItemId,
        },
    },
    selectedItemType: 'node',
} as any;

const fakeSelectedEdge = {
    selectedItemQuery: {
        data: {
            objectId: fakeSelectedItemId,
            id: '123_MemberOf_456',
        },
    },
    selectedItemType: 'edge',
} as any;

const fakeSelectedNonFilterable = JSON.parse(JSON.stringify(fakeSelectedEdge));
fakeSelectedNonFilterable.selectedItemQuery.data.id = '123_CrossForestTrust_456';

const fakeSelectedNonEdge = JSON.parse(JSON.stringify(fakeSelectedEdge));
fakeSelectedNonEdge.selectedItemQuery.data.id = '345';

const server = setupServer(
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions([Permission.GRAPH_DB_WRITE]).user,
            })
        );
    }),
    rest.get('/api/v2/asset-groups/:assetGroupId/members', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    members: [],
                },
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
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

interface SetupOptions {
    exploreSearchTab?: string;
    permissions?: Permission[];
    primarySearch?: string;
    secondarySearch?: string;
}

const setup = async ({ exploreSearchTab, permissions, primarySearch, secondarySearch }: SetupOptions) => {
    const initialState: DeepPartial<AppState> = {
        assetgroups: {
            assetGroups: [
                { tag: 'owned', id: 1 },
                { tag: 'admin_tier_0', id: 2 },
            ],
        },
    };

    if (permissions) {
        initialState.auth = createAuthStateWithPermissions(permissions);
    }

    const mockSetExploreParams = vi.fn();
    mockUseExploreParams.mockReturnValue({
        setExploreParams: mockSetExploreParams,
        primarySearch,
        secondarySearch,
        exploreSearchTab,
    } as any);

    const mockPathfindingFilters: PathfindingFilters = {
        handleApplyFilters: vi.fn(),
        handleRemoveEdgeType: vi.fn(),
        handleUpdateFilters: vi.fn(),
        initialize: vi.fn(),
        selectedFilters: [],
    };

    const screen = await act(async () => {
        render(
            <ContextMenu
                contextMenu={{ x: 0, y: 0 }}
                handleClose={vi.fn()}
                pathfindingFilters={mockPathfindingFilters}
            />,
            {
                initialState,
            }
        );
    });

    const user = userEvent.setup();
    return { screen, user, mockSetExploreParams, mockPathfindingFilters };
};

describe('ContextMenu - Nodes', async () => {
    beforeEach(() => mockSelectedItemQuery.mockReturnValue(fakeSelectedNode));

    it('renders asset group edit options with graph write permissions', async () => {
        await setup({ permissions: [Permission.GRAPH_DB_WRITE] });

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        const addToHighValueOption = screen.getByRole('menuitem', { name: /add to high value/i });
        const addToOwnedOption = screen.getByRole('menuitem', { name: /add to owned/i });

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
        expect(addToHighValueOption).toBeInTheDocument();
        expect(addToOwnedOption).toBeInTheDocument();
    });

    it('renders no asset group edit options without graph write permissions', async () => {
        await setup({});

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        const addToHighValueOption = screen.queryByText(/add to tier zero/i);

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
        expect(addToHighValueOption).toBeNull();
    });

    it('sets a primarySearch=id and searchType=node when secondarySearch is falsey', async () => {
        const { user, mockSetExploreParams } = await setup({});

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            primarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'node',
        });
    });

    it('sets a primarySearch=id and searchType=pathfinding when secondarySearch is truthy', async () => {
        const secondarySearch = 'cdf';
        const { user, mockSetExploreParams } = await setup({ secondarySearch });

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            primarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'pathfinding',
        });
    });

    it('sets secondarySearch=id and searchType=node when primarySearch is falsey', async () => {
        const { user, mockSetExploreParams } = await setup({});

        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        await user.click(endNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            secondarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'node',
        });
    });

    it('sets a secondary=id and searchType=pathfinding when primary is truthy', async () => {
        const secondarySearch = 'cdf';
        const { user, mockSetExploreParams } = await setup({ primarySearch: secondarySearch });

        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        await user.click(endNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            secondarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'pathfinding',
        });
    });

    it('opens a submenu when user hovers over `Copy`', async () => {
        await setup({});

        const user = userEvent.setup();

        const copyOption = screen.getByRole('menuitem', { name: /copy/i });
        await user.hover(copyOption);

        const tip = await screen.findByRole('tooltip');
        expect(tip).toBeInTheDocument();

        const displayNameOption = screen.getByLabelText(/display name/i);
        expect(displayNameOption).toBeInTheDocument();

        const objectIdOption = screen.getByLabelText(/object id/i);
        expect(objectIdOption).toBeInTheDocument();

        const cypherOption = screen.getByLabelText(/cypher/i);
        expect(cypherOption).toBeInTheDocument();

        // hover off the `Copy` option in order to close the tooltip
        await userEvent.unhover(copyOption);

        await waitFor(() => {
            expect(screen.queryByText(/display name/i)).not.toBeInTheDocument();
            expect(screen.queryByText(/object id/i)).not.toBeInTheDocument();
            expect(screen.queryByText(/cypher/i)).not.toBeInTheDocument();
        });
    });
});

describe('ContextMenu - Edges', () => {
    it('shows edge filtering options on pathfinding tab', async () => {
        mockSelectedItemQuery.mockReturnValue(fakeSelectedEdge);
        await setup({ exploreSearchTab: 'pathfinding' });

        const filterEdgeOption = screen.getByRole('menuitem', { name: /Filter out Edge/i });
        expect(filterEdgeOption).toBeInTheDocument();
    });

    it('does not show edge filtering on non-pathfinding tab', async () => {
        mockSelectedItemQuery.mockReturnValue(fakeSelectedEdge);
        await setup({ exploreSearchTab: 'cypher' });

        const filterEdgeOption = screen.queryByText('Filter out Edge');
        expect(filterEdgeOption).not.toBeInTheDocument();
    });

    it('does not show edge filtering when bad edge id', async () => {
        mockSelectedItemQuery.mockReturnValue(fakeSelectedNonEdge);
        await setup({ exploreSearchTab: 'pathfinding' });

        const filterEdgeOption = screen.queryByText('Filter out Edge');
        expect(filterEdgeOption).not.toBeInTheDocument();
    });

    it('shows edge as non-filterable where there is no edge id', async () => {
        mockSelectedItemQuery.mockReturnValue(fakeSelectedNonFilterable);
        await setup({ exploreSearchTab: 'pathfinding' });

        const filterEdgeOption = screen.queryByText('Filter out Edge');
        const nonFilterEdgeOption = screen.getByRole('menuitem', { name: /Non-filterable Edge/i });
        expect(filterEdgeOption).not.toBeInTheDocument();
        expect(nonFilterEdgeOption).toBeInTheDocument();
    });

    it('filters out the selected edge', async () => {
        mockSelectedItemQuery.mockReturnValue(fakeSelectedEdge);
        const { user, mockPathfindingFilters } = await setup({ exploreSearchTab: 'pathfinding' });

        const filterEdgeOption = screen.getByRole('menuitem', { name: /Filter out Edge/i });
        await user.click(filterEdgeOption);

        expect(mockPathfindingFilters.handleRemoveEdgeType).toBeCalled();
    });
});
