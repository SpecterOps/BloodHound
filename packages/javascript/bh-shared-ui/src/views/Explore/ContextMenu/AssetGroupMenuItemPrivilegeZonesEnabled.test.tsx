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
import { render, screen } from '../../../test-utils';
import { apiClient } from '../../../utils';
import { AssetGroupMenuItem } from './AssetGroupMenuItemPrivilegeZonesEnabled';

const selfResponse = {
    data: {
        roles: [
            {
                permissions: [
                    {
                        authority: 'graphdb',
                        name: 'Write',
                        id: 14,
                    },
                ],
            },
        ],
    },
};

const assetGroupTagsResponse = {
    data: {
        tags: [
            {
                id: 2,
                type: 3,
                name: 'Owned',
                position: null,
            },
            {
                id: 1,
                type: 1,
                name: 'Tier Zero',
                position: 1,
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

const cypherSearchTZResponse = {
    data: {
        nodes: {
            '1234': {
                ...cypherSearchResponse.data.nodes['1234'],
                isTierZero: true,
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
    rest.get('/api/v2/graph-search', (req, res, ctx) => {
        return res(ctx.json({}));
    }),
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.json(cypherSearchResponse));
    }),
    rest.get('/api/v2/asset-group-tags', (req, res, ctx) => {
        return res(ctx.json(assetGroupTagsResponse));
    }),
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(ctx.json(selfResponse));
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
    });

    afterAll(() => server.close());

    describe('adding to an asset group', () => {
        it('adds node to tier zero asset group', async () => {
            const mutateSpy = vi.spyOn(apiClient, 'createAssetGroupTagSelector');

            render(<AssetGroupMenuItem assetGroupType='tierZero' showConfirmationOnAdd={true} />, {
                route: ROUTE_WITH_SELECTED_ITEM_PARAM,
            });

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

        it('adds node to owned asset group', async () => {
            const mutateSpy = vi.spyOn(apiClient, 'createAssetGroupTagSelector');

            render(<AssetGroupMenuItem assetGroupType='owned' />, {
                route: ROUTE_WITH_SELECTED_ITEM_PARAM,
            });

            const user = userEvent.setup();

            const addToOwnedButton = await screen.findByRole('menuitem', { name: /Add to Owned/i });
            expect(addToOwnedButton).toBeInTheDocument();

            await user.click(addToOwnedButton);

            expect(mutateSpy).toHaveBeenCalled();
        });
    });

    describe('removing from an asset group', () => {
        it('removes a node from a tier zero asset group', async () => {
            server.use(
                rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                    return res(ctx.json(cypherSearchTZResponse));
                })
            );

            render(<AssetGroupMenuItem assetGroupType='tierZero' />, { route: ROUTE_WITH_SELECTED_ITEM_PARAM });

            const user = userEvent.setup();

            const removeButton = await screen.findByRole('menuitem', { name: /Remove from Tier Zero/i });
            expect(removeButton).toBeInTheDocument();

            await user.click(removeButton);
            expect(window.location.pathname).toBe('/privilege-zones/zones/1/details');
        });

        it('removes a node from an owned asset group', async () => {
            server.use(
                rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                    return res(ctx.json(cypherSearchOwnedResponse));
                })
            );

            render(<AssetGroupMenuItem assetGroupType='owned' />, { route: ROUTE_WITH_SELECTED_ITEM_PARAM });

            const user = userEvent.setup();

            const removeButton = await screen.findByRole('menuitem', { name: /Remove from Owned/i });
            expect(removeButton).toBeInTheDocument();

            await user.click(removeButton);
            expect(window.location.pathname).toBe('/privilege-zones/labels/2/details');
        });
    });
});
