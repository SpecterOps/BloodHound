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
import { Permission, createAuthStateWithPermissions, type DeepPartial, type PathfindingFilters } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act } from 'react-dom/test-utils';
import { type UseQueryResult } from 'react-query';
import { type AppState } from 'src/store';
import { render, screen, waitFor } from 'src/test-utils';
import ContextMenu from './ContextMenu';

type QueryResponse = UseQueryResult<bhSharedUi.ItemResponse, unknown>;

const mockUseContextMenuItems = vi.spyOn(bhSharedUi, 'useContextMenuItems');
const mockSetExploreParams = vi.fn();

const fakeSelectedItemId = 'abc';
const fakeSelectedNode = {
    objectId: fakeSelectedItemId,
} as bhSharedUi.ItemResponse;

const fakeSelectedEdge = {
    id: '123_MemberOf_456',
    source: 'edge_source',
} as bhSharedUi.ItemResponse;

const fakeSelectedNonFilterable = {
    id: '123_CrossForestTrust_456',
    source: 'edge_source',
} as bhSharedUi.ItemResponse;

const fakeSelectedNonEdge = {
    id: '345',
    source: 'edge_source',
} as bhSharedUi.ItemResponse;

const menuPosition = { mouseX: 0, mouseY: 0 };

const asEdgeItem = (query: QueryResponse) => (bhSharedUi.isEdge(query.data) ? query.data : undefined);
const asNodeItem = (query: QueryResponse) => (bhSharedUi.isNode(query.data) ? query.data : undefined);

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

const mockPathfindingFilters: PathfindingFilters = {
    handleApplyFilters: vi.fn(),
    handleUpdateAndApplyFilter: vi.fn(),
    handleUpdateFilters: vi.fn(),
    initialize: vi.fn(),
    selectedFilters: [],
};

const setUseContextMenuItems = ({
    isAssetGroupEnabled = false,
    primarySearch = fakeSelectedItemId,
    secondarySearch = '',
    selectedItem,
}: {
    isAssetGroupEnabled?: boolean;
    primarySearch?: string;
    secondarySearch?: string;
    selectedItem: bhSharedUi.ItemResponse;
}) => {
    mockUseContextMenuItems.mockReturnValue({
        asEdgeItem,
        asNodeItem,
        exploreParams: {
            setExploreParams: mockSetExploreParams,
            primarySearch,
            secondarySearch,
        },
        isAssetGroupEnabled,
        menuPosition: { left: menuPosition.mouseX, top: menuPosition.mouseY },
        selectedItemQuery: {
            data: selectedItem as bhSharedUi.ItemResponse,
        },
    } as any);
};

const setup = async () => {
    const initialState: DeepPartial<AppState> = {
        assetgroups: {
            assetGroups: [
                { tag: 'owned', id: 1 },
                { tag: 'admin_tier_0', id: 2 },
            ],
        },
    };

    const screen = await act(async () => {
        render(<ContextMenu pathfindingFilters={mockPathfindingFilters} position={menuPosition} onClose={vi.fn()} />, {
            initialState,
        });
    });

    const user = userEvent.setup();
    return { screen, user, mockPathfindingFilters };
};

describe('ContextMenu - Nodes', async () => {
    it('renders asset group edit options with graph write permissions', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode, isAssetGroupEnabled: true });
        await setup();

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
        setUseContextMenuItems({ selectedItem: fakeSelectedNode });
        await setup();

        const filterEdgeOption = screen.queryByText(/Filter out Edge/i);
        expect(filterEdgeOption).toBeNull();
    });

    it('renders no edge path filters', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode });
        await setup();

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        const addToHighValueOption = screen.queryByText(/add to tier zero/i);

        expect(startNodeOption).toBeInTheDocument();
        expect(endNodeOption).toBeInTheDocument();
        expect(addToHighValueOption).toBeNull();
    });

    it('sets a primarySearch=id and searchType=node when secondarySearch is falsey', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode, primarySearch: fakeSelectedItemId });
        const { user } = await setup();

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            primarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'node',
        });
    });

    it('sets a primarySearch=id and searchType=pathfinding when secondarySearch is truthy', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode, secondarySearch: fakeSelectedItemId });
        const { user } = await setup();

        const startNodeOption = screen.getByRole('menuitem', { name: /set as starting node/i });
        await user.click(startNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            primarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'pathfinding',
        });
    });

    it('sets secondarySearch=id and searchType=node when primarySearch is falsey', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode, primarySearch: '' });
        const { user } = await setup();

        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        await user.click(endNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            secondarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'node',
        });
    });

    it('sets a secondary=id and searchType=pathfinding when primary is truthy', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode, primarySearch: fakeSelectedItemId });
        const { user } = await setup();

        const endNodeOption = screen.getByRole('menuitem', { name: /set as ending node/i });
        await user.click(endNodeOption);

        expect(mockSetExploreParams).toBeCalledWith({
            secondarySearch: fakeSelectedItemId,
            exploreSearchTab: 'pathfinding',
            searchType: 'pathfinding',
        });
    });

    it('opens a submenu when user hovers over `Copy`', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNode, isAssetGroupEnabled: true });
        await setup();

        const user = userEvent.setup();

        const copyOption = screen.getByRole('menuitem', { name: /copy/i });
        await user.hover(copyOption);

        const tip = await screen.findByRole('tooltip');
        expect(tip).toBeInTheDocument();

        const nameOption = screen.getByLabelText(/name/i);
        expect(nameOption).toBeInTheDocument();

        const objectIdOption = screen.getByLabelText(/object id/i);
        expect(objectIdOption).toBeInTheDocument();

        const cypherOption = screen.getByLabelText(/cypher/i);
        expect(cypherOption).toBeInTheDocument();

        // hover off the `Copy` option in order to close the tooltip
        await userEvent.unhover(copyOption);

        await waitFor(() => {
            expect(screen.queryByText(/name/i)).not.toBeInTheDocument();
            expect(screen.queryByText(/object id/i)).not.toBeInTheDocument();
            expect(screen.queryByText(/cypher/i)).not.toBeInTheDocument();
        });
    });
});

describe('ContextMenu - Edges', () => {
    it('shows edge filtering options on pathfinding tab', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedEdge });
        await setup();

        const filterEdgeOption = screen.getByRole('menuitem', { name: /Filter out Edge/i });
        expect(filterEdgeOption).toBeInTheDocument();
    });

    it('shows edge as non-filterable where there is no edge id', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNonFilterable });
        await setup();

        const filterEdgeOption = screen.queryByText('Filter out Edge');
        const nonFilterEdgeOption = screen.getByRole('menuitem', { name: /Non-filterable Edge/i });
        expect(filterEdgeOption).not.toBeInTheDocument();
        expect(nonFilterEdgeOption).toBeInTheDocument();
    });

    it('filters out the selected edge', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedEdge });
        const { user } = await setup();

        const filterEdgeOption = screen.getByRole('menuitem', { name: /Filter out Edge/i });
        await user.click(filterEdgeOption);

        expect(mockPathfindingFilters.handleUpdateAndApplyFilter).toBeCalled();
    });

    it('does not show edge filtering when bad edge id', async () => {
        setUseContextMenuItems({ selectedItem: fakeSelectedNonEdge });
        await setup();

        const filterEdgeOption = screen.queryByText('Filter out Edge');
        expect(filterEdgeOption).not.toBeInTheDocument();
    });
});
