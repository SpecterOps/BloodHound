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

import { Dialog } from '@bloodhoundenterprise/doodleui';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useState } from 'react';
import { act } from 'react-dom/test-utils';
import { render, screen } from 'src/test-utils';
import AssetGroupMenuItem from './AssetGroupMenuItemZoneManagementEnabled';

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

const AssetGroupMenuItemWithDialog = (props: any) => {
    const [dialogOpen, setDialogOpen] = useState(false);

    return (
        <Dialog open={dialogOpen}>
            <AssetGroupMenuItem
                {...props}
                onShowConfirmation={() => setDialogOpen(true)}
                onCancelConfirmation={() => setDialogOpen(false)}
            />
        </Dialog>
    );
};

describe('AssetGroupMenuItem', () => {
    describe('adding to an asset group', () => {
        const server = setupServer(
            rest.get('/api/v2/graph-search', (req, res, ctx) => {
                return res(ctx.json({}));
            }),
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json({}));
            }),
            rest.post('/api/v2/asset-group-tags/:tagId/selectors', (req, res, ctx) => {
                return res(ctx.json({}));
            })
        );

        beforeAll(() => server.listen());

        afterEach(() => {
            server.resetHandlers();
        });

        afterAll(() => server.close());

        it('handles adding to tier zero asset group', async () => {
            const testOnAddNode = vi.fn();
            const testRemoveNodePath = '/meow';

            render(
                <AssetGroupMenuItemWithDialog
                    assetGroupId={tierZeroAssetGroup.id}
                    assetGroupName={tierZeroAssetGroup.name}
                    isCurrentMember={false}
                    showConfirmationOnAdd={true}
                    onAddNode={testOnAddNode}
                    removeNodePath={testRemoveNodePath}
                />,
                {
                    route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                }
            );

            const user = userEvent.setup();

            const addToHighValueButton = screen.getByRole('menuitem', { name: /add to high value/i });
            expect(addToHighValueButton).toBeInTheDocument();

            await user.click(addToHighValueButton);

            const confirmationDialog = screen.getByRole('dialog', { name: /confirm selection/i });
            expect(confirmationDialog).toBeVisible();

            const applyButton = screen.getByRole('button', { name: /ok/i });
            await user.click(applyButton);

            expect(testOnAddNode).toHaveBeenCalledTimes(1);
            expect(testOnAddNode).toHaveBeenCalledWith(tierZeroAssetGroup.id);
        });

        it('handles adding to non-tier-zero asset group', async () => {
            const testOnAddNode = vi.fn();
            const testRemoveNodePath = '/meow';

            render(
                <AssetGroupMenuItem
                    assetGroupId={ownedAssetGroup.id}
                    assetGroupName={ownedAssetGroup.name}
                    isCurrentMember={false}
                    onAddNode={testOnAddNode}
                    removeNodePath={testRemoveNodePath}
                />,
                {
                    route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                }
            );

            const user = userEvent.setup();

            const addButton = screen.getByRole('menuitem', { name: /add to owned/i });
            expect(addButton).toBeInTheDocument();

            await user.click(addButton);

            expect(testOnAddNode).toHaveBeenCalledTimes(1);
            expect(testOnAddNode).toHaveBeenCalledWith(ownedAssetGroup.id);
        });
    });

    describe('removing from an asset group', () => {
        const server = setupServer(
            rest.get('/api/v2/graph-search', (req, res, ctx) => {
                return res(ctx.json({}));
            }),
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json({}));
            }),
            rest.post('/api/v2/asset-group-tags/:tagId/selectors', (req, res, ctx) => {
                return res(ctx.json({}));
            })
        );

        beforeAll(() => server.listen());

        afterEach(() => {
            server.resetHandlers();
        });

        afterAll(() => server.close());

        it('handles removing from a tier zero asset group', async () => {
            const testRemoveNodePath = '/meow';
            await act(async () => {
                await render(
                    <AssetGroupMenuItem
                        assetGroupId={tierZeroAssetGroup.id}
                        assetGroupName={tierZeroAssetGroup.name}
                        isCurrentMember={true}
                        removeNodePath={testRemoveNodePath}
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

            const removeButton = screen.getByRole('menuitem', { name: /remove from high value/i });
            expect(removeButton).toBeInTheDocument();

            await user.click(removeButton);

            expect(window.location.pathname).toBe(testRemoveNodePath);
        });

        it('handles removing from a non-tier-zero asset group', async () => {
            const testRemoveNodePath = '/meow';
            await act(async () => {
                await render(
                    <AssetGroupMenuItem
                        assetGroupId={ownedAssetGroup.id}
                        assetGroupName={ownedAssetGroup.name}
                        isCurrentMember={true}
                        removeNodePath={testRemoveNodePath}
                    />,
                    {
                        initialState: {
                            ...getAssetGroupTestProps({ isTierZero: false }),
                        },
                        route: ROUTE_WITH_SELECTED_ITEM_PARAM,
                    }
                );
            });

            const user = userEvent.setup();

            const removeButton = screen.getByRole('menuitem', { name: /remove from owned/i });
            expect(removeButton).toBeInTheDocument();

            await user.click(removeButton);

            expect(window.location.pathname).toBe(testRemoveNodePath);
        });
    });
});
