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
import { getIsOwnedTag, getIsTierZeroTag, isOwnedObject, isTierZero } from '../../../hooks';
import { createAuthStateWithPermissions } from '../../../mocks';
import { withoutErrorLogging } from '../../../mocks/stderr';
import { render, screen } from '../../../test-utils';
import { Permission, apiClient } from '../../../utils';
import { AssetGroupMenuItem } from './AssetGroupMenuItemPrivilegeZonesEnabled';

const assetGroupTags = {
    data: {
        tags: [
            {
                id: 2,
                type: 3,
                kind_id: 2,
                name: 'Owned',
                description: 'Owned',
            },
            {
                id: 1,
                type: 1,
                kind_id: 1,
                position: 1,
                name: 'Tier Zero',
                description: 'Tier Zero',
            },
        ],
    },
};

const cypherSearchResponse = {
    data: {
        nodes: {
            '1234': {
                label: 'TEST@DOMAIN.CORP',
                kind: 'GPO',
                kinds: ['Base', 'GPO'],
                objectId: 'A1B2C3D4-1111-2222-3333-0123456789AB',
                isTierZero: false,
                isOwnedObject: false,
                lastSeen: '2025-12-04T20:16:49.209Z',
            },
        },
        edges: [],
    },
};

const cypherSearchOwnedResponse = {
    data: {
        nodes: {
            '1234': {
                ...cypherSearchResponse.data.nodes['1234'],
                isOwnedObject: true,
            },
        },
        edges: [],
    },
};

const getEntityInfoTestProps = () => ({
    entityinfo: {
        selectedNode: {
            name: 'foo',
            id: '1234',
            type: 'User',
        },
    },
});

const server = setupServer(
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 1, key: 'tier_management_engine', enabled: true }],
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags', (req, res, ctx) => {
        return res(ctx.json(assetGroupTags));
    }),
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.json(cypherSearchResponse));
    }),
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions([Permission.GRAPH_DB_WRITE]).user,
            })
        );
    }),
    rest.post('/api/v2/asset-group-tags/:tagId/selectors', (req, res, ctx) => {
        return res(ctx.json({}));
    })
);

const ROUTE_WITH_SELECTED_ITEM_PARAM = `?selectedItem=${getEntityInfoTestProps().entityinfo.selectedNode.id}&searchType=node&primarySearch=${getEntityInfoTestProps().entityinfo.selectedNode.id}`;

describe('AssetGroupMenuItem', () => {
    beforeAll(() => server.listen());

    afterEach(() => {
        server.resetHandlers();
        vi.restoreAllMocks();
    });

    afterAll(() => server.close());

    it('shows a loading state', async () => {
        render(
            <AssetGroupMenuItem
                addNodePayload={{} as any}
                isCurrentMemberFn={isOwnedObject}
                removeNodePathFn={() => '/privilege-zones/labels/1/details'}
                tagIdentifierFn={getIsOwnedTag}
            />,
            {
                route: ROUTE_WITH_SELECTED_ITEM_PARAM,
            }
        );

        const loadingState = await screen.findByRole('menuitem', { name: /Loading/i });
        expect(loadingState).toBeInTheDocument();
    });

    it('shows an error state', async () => {
        server.use(
            rest.get('/api/v2/asset-group-tags', (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );

        await withoutErrorLogging(async () => {
            render(
                <AssetGroupMenuItem
                    addNodePayload={{} as any}
                    isCurrentMemberFn={isOwnedObject}
                    removeNodePathFn={() => '/privilege-zones/labels/1/details'}
                    tagIdentifierFn={getIsOwnedTag}
                />,
                {
                    route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                }
            );

            const errorState = await screen.findByRole('menuitem', { name: /Unavailable/i });
            expect(errorState).toBeInTheDocument();
        });
    });

    it('adds node to asset group tag with confirmation', async () => {
        const mutateSpy = vi.spyOn(apiClient, 'createAssetGroupTagSelector');

        render(
            <AssetGroupMenuItem
                addNodePayload={{} as any}
                isCurrentMemberFn={isTierZero}
                removeNodePathFn={() => '/privilege-zones/zones/1/details'}
                tagIdentifierFn={getIsTierZeroTag}
                showConfirmationOnAdd={true}
            />,
            {
                route: ROUTE_WITH_SELECTED_ITEM_PARAM,
            }
        );

        const user = userEvent.setup();

        const addToHighValueButton = await screen.findByRole('menuitem', { name: /Add to Tier Zero/i });
        expect(addToHighValueButton).toBeInTheDocument();

        await user.click(addToHighValueButton);

        const confirmationDialog = screen.getByRole('dialog', { name: /Confirm Selection/i });
        expect(confirmationDialog).toBeVisible();

        const applyButton = screen.getByRole('button', { name: /Ok/i });

        await user.click(applyButton);

        expect(mutateSpy).toHaveBeenCalled();
    });

    it('adds node to asset group tag without confirmation', async () => {
        const mutateSpy = vi.spyOn(apiClient, 'createAssetGroupTagSelector');

        render(
            <AssetGroupMenuItem
                addNodePayload={{} as any}
                isCurrentMemberFn={isOwnedObject}
                removeNodePathFn={() => '/privilege-zones/labels/1/details'}
                tagIdentifierFn={getIsOwnedTag}
            />,
            {
                route: ROUTE_WITH_SELECTED_ITEM_PARAM,
            }
        );

        const user = userEvent.setup();

        const addToOwnedButton = await screen.findByRole('menuitem', { name: /Add to Owned/i });
        expect(addToOwnedButton).toBeInTheDocument();

        await user.click(addToOwnedButton);

        expect(mutateSpy).toHaveBeenCalled();
    });

    it('removes a node from an asset group tag', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json(cypherSearchOwnedResponse));
            })
        );

        render(
            <AssetGroupMenuItem
                addNodePayload={{} as any}
                isCurrentMemberFn={isOwnedObject}
                removeNodePathFn={() => '/privilege-zones/labels/1/details'}
                tagIdentifierFn={getIsOwnedTag}
            />,
            { route: ROUTE_WITH_SELECTED_ITEM_PARAM }
        );

        const user = userEvent.setup();

        const removeButton = await screen.findByRole('menuitem', { name: /Remove from Owned/i });
        expect(removeButton).toBeInTheDocument();

        await user.click(removeButton);
        expect(window.location.pathname).toBe('/privilege-zones/labels/1/details');
    });
});
