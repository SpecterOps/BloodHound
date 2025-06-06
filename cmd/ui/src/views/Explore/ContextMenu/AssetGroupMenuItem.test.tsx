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
import { apiClient } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act } from 'react-dom/test-utils';
import { render, screen } from 'src/test-utils';
import AssetGroupMenuItem from './AssetGroupMenuItem';

const tierZeroAssetGroup = { id: 1, name: 'high value' };
const ownedAssetGroup = { id: 2, name: 'owned' };

const getEntityInfoTestProps = () => ({
    entityinfo: {
        selectedNode: {
            name: 'foo',
            id: '1234',
            type: 'User',
        },
    },
});

const ROUTE_WITH_SELECTED_ITEM_PARAM = `?selectedItem=${getEntityInfoTestProps().entityinfo.selectedNode.id}&searchType=node&primarySearch=${getEntityInfoTestProps().entityinfo.selectedNode.id}`;

const getAssetGroupTestProps = ({ isTierZero }: { isTierZero: boolean }) => ({
    assetgroups: {
        assetGroups: isTierZero
            ? [{ tag: 'admin_tier_0', id: tierZeroAssetGroup.id }]
            : [{ tag: 'owned', id: ownedAssetGroup.id }],
    },
});

describe('AssetGroupMenuItem', async () => {
    describe('adding to an asset group', () => {
        const server = setupServer(
            rest.get('/api/v2/asset-groups/:assetGroupId/members', (req, res, ctx) => {
                // handle `tier zero` requests
                if (req.params.assetGroupId === tierZeroAssetGroup.id.toString()) {
                    return res(
                        ctx.json({
                            data: {
                                members: [],
                            },
                        })
                    );
                } else if (req.params.assetGroupId === ownedAssetGroup.id.toString()) {
                    // handle `owned` requests
                    return res(
                        ctx.json({
                            data: {
                                // members: [{ custom_member: true }],
                                members: [],
                            },
                        })
                    );
                } else {
                    return res(ctx.json({}));
                }
            }),
            rest.put('/api/v2/asset-groups/:assetGroupId/selectors', (req, res, ctx) => {
                return res(ctx.json({}));
            }),
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: { nodes: [{ objectId: getEntityInfoTestProps().entityinfo.selectedNode.id }] },
                    })
                );
            }),
            rest.get('/api/v2/graph-search', (req, res, ctx) => {
                return res(ctx.json({}));
            })
        );

        beforeAll(() => server.listen());

        afterEach(() => {
            server.resetHandlers();
        });

        afterAll(() => server.close());

        it('handles adding to tier zero asset group', async () => {
            await act(async () => {
                render(
                    <AssetGroupMenuItem
                        assetGroupId={tierZeroAssetGroup.id}
                        assetGroupName={tierZeroAssetGroup.name}
                    />,
                    {
                        initialState: {
                            ...getAssetGroupTestProps({ isTierZero: true }),
                        },
                        route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                    }
                );
            });

            const user = userEvent.setup();
            const addToHighValueSpy = vi.spyOn(apiClient, 'updateAssetGroupSelector');

            const addToHighValueButton = screen.getByRole('menuitem', { name: /add to high value/i });
            expect(addToHighValueButton).toBeInTheDocument();

            await user.click(addToHighValueButton);

            const confirmationDialog = screen.getByRole('dialog', { name: /confirm selection/i });
            expect(confirmationDialog).toBeInTheDocument();

            const applyButton = screen.getByRole('button', { name: /ok/i });
            await user.click(applyButton);

            expect(addToHighValueSpy).toHaveBeenCalledTimes(1);
            expect(addToHighValueSpy).toHaveBeenCalledWith(tierZeroAssetGroup.id, [
                {
                    action: 'add',
                    selector_name: '1234',
                    sid: '1234',
                },
            ]);
        });

        it('handles adding to non-tier-zero asset group', async () => {
            await act(async () => {
                render(<AssetGroupMenuItem assetGroupId={ownedAssetGroup.id} assetGroupName={ownedAssetGroup.name} />, {
                    initialState: {
                        ...getAssetGroupTestProps({ isTierZero: false }),
                    },
                    route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                });
            });

            const user = userEvent.setup();
            const addToAssetGroupSpy = vi.spyOn(apiClient, 'updateAssetGroupSelector');

            const addButton = screen.getByRole('menuitem', { name: /add to owned/i });
            expect(addButton).toBeInTheDocument();

            await user.click(addButton);

            expect(addToAssetGroupSpy).toHaveBeenCalledTimes(1);
            expect(addToAssetGroupSpy).toHaveBeenCalledWith(ownedAssetGroup.id, [
                {
                    action: 'add',
                    selector_name: '1234',
                    sid: '1234',
                },
            ]);
        });

        it('renders null if network fails to return valid asset group membership list', async () => {
            render(<AssetGroupMenuItem assetGroupId={3} assetGroupName={'blah'} />, {});

            expect(document.body.firstChild).toBeEmptyDOMElement();
        });
    });

    describe('removing from an asset group', () => {
        const server = setupServer(
            rest.get('/api/v2/asset-groups/:assetGroupId/members', (req, res, ctx) => {
                // handle `tier zero` requests
                if (req.params.assetGroupId === tierZeroAssetGroup.id.toString()) {
                    return res(
                        ctx.json({
                            data: {
                                members: [{ custom_member: true }],
                            },
                        })
                    );
                } else if (req.params.assetGroupId === ownedAssetGroup.id.toString()) {
                    // handle `owned` requests
                    return res(
                        ctx.json({
                            data: {
                                members: [{ custom_member: true }],
                            },
                        })
                    );
                } else {
                    return res(ctx.json({}));
                }
            }),
            rest.put('/api/v2/asset-groups/:assetGroupId/selectors', (req, res, ctx) => {
                return res(ctx.json({}));
            }),
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: { nodes: [{ objectId: getEntityInfoTestProps().entityinfo.selectedNode.id }] },
                    })
                );
            }),
            rest.get('/api/v2/graph-search', (req, res, ctx) => {
                return res(ctx.json({}));
            })
        );

        beforeAll(() => server.listen());

        afterEach(() => {
            server.resetHandlers();
        });

        afterAll(() => server.close());

        it('handles removing from a tier zero asset group', async () => {
            await act(async () => {
                await render(
                    <AssetGroupMenuItem
                        assetGroupId={tierZeroAssetGroup.id}
                        assetGroupName={tierZeroAssetGroup.name}
                    />,
                    {
                        initialState: {
                            ...getAssetGroupTestProps({ isTierZero: true }),
                        },
                        route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                    }
                );
            });

            const user = userEvent.setup();
            const removeFromAssetGroupSpy = vi.spyOn(apiClient, 'updateAssetGroupSelector');

            const removeButton = screen.getByRole('menuitem', { name: /remove from high value/i });
            expect(removeButton).toBeInTheDocument();

            await user.click(removeButton);

            const confirmationDialog = screen.getByRole('dialog', { name: /confirm selection/i });
            expect(confirmationDialog).toBeInTheDocument();

            const applyButton = screen.getByRole('button', { name: /ok/i });
            await user.click(applyButton);

            expect(removeFromAssetGroupSpy).toHaveBeenCalledTimes(1);
            expect(removeFromAssetGroupSpy).toHaveBeenCalledWith(tierZeroAssetGroup.id, [
                {
                    action: 'remove',
                    selector_name: '1234',
                    sid: '1234',
                },
            ]);
        });

        it('handles removing from a non-tier-zero asset group', async () => {
            await act(async () => {
                await render(
                    <AssetGroupMenuItem assetGroupId={ownedAssetGroup.id} assetGroupName={ownedAssetGroup.name} />,
                    {
                        initialState: {
                            ...getAssetGroupTestProps({ isTierZero: false }),
                        },
                        route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                    }
                );
            });

            const user = userEvent.setup();
            const removeFromAssetGroupSpy = vi.spyOn(apiClient, 'updateAssetGroupSelector');

            const removeButton = screen.getByRole('menuitem', { name: /remove from owned/i });
            expect(removeButton).toBeInTheDocument();

            await user.click(removeButton);

            expect(removeFromAssetGroupSpy).toHaveBeenCalledTimes(1);
            expect(removeFromAssetGroupSpy).toHaveBeenCalledWith(ownedAssetGroup.id, [
                {
                    action: 'remove',
                    selector_name: '1234',
                    sid: '1234',
                },
            ]);
        });
    });
});
