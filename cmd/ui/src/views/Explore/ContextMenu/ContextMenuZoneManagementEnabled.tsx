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

import { Dialog } from '@bloodhoundenterprise/doodleui';
import { Menu, MenuItem } from '@mui/material';
import {
    CopyMenuItems,
    EdgeMenuItems,
    NodeMenuItems,
    apiClient,
    useContextMenuItems,
    useNotifications,
    type MousePosition,
    type NodeResponse,
    type PathfindingFilters,
} from 'bh-shared-ui';
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import AssetGroupMenuItem from './AssetGroupMenuItemZoneManagementEnabled';

const ContextMenu: FC<{
    onClose: () => void;
    pathfindingFilters: PathfindingFilters;
    position: MousePosition | null;
}> = ({ onClose, pathfindingFilters, position }) => {
    const { asEdgeItem, asNodeItem, exploreParams, isAssetGroupEnabled, menuPosition, selectedItemQuery } =
        useContextMenuItems(position);

    const edgeItem = asEdgeItem(selectedItemQuery);
    const nodeItem = asNodeItem(selectedItemQuery);

    const { addNotification } = useNotifications();

    const getAssetGroupTagsQuery = useQuery(['getAssetGroupTags'], ({ signal }) =>
        apiClient.getAssetGroupTags({ signal }).then((res) => res.data.data)
    );

    const createAssetGroupTagSelectorMutation = useMutation({
        mutationFn: ({ assetGroupId, node }: { assetGroupId: string | number; node: NodeResponse }) => {
            return apiClient.createAssetGroupTagSelector(assetGroupId, {
                name: node.label ?? node.objectId,
                seeds: [
                    {
                        type: SeedTypeObjectId,
                        value: node.objectId,
                    },
                ],
            });
        },
        onSuccess: () => {
            addNotification('Node successfully added.', 'AssetGroupUpdateSuccess');
        },
        onError: (error: any) => {
            console.error(error);
            addNotification('An error occurred when adding node', 'AssetGroupUpdateError');
        },
    });

    const [dialogOpen, setDialogOpen] = useState(false);

    // Null position is used to indicated the menu is closed
    if (menuPosition === null && !dialogOpen) {
        return null;
    }

    const handleAddNode = (assetGroupId: string | number) => {
        if (nodeItem) {
            createAssetGroupTagSelectorMutation.mutate(
                {
                    assetGroupId,
                    node: nodeItem,
                },
                {
                    onSettled: () => {
                        setDialogOpen(false);
                    },
                }
            );
        }
    };

    const tierZeroAssetGroup = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.position === 1;
    });

    const ownedAssetGroup = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.type === 3;
    });

    const isLoading = getAssetGroupTagsQuery.isLoading || selectedItemQuery.isLoading;
    const isError = getAssetGroupTagsQuery.isError || selectedItemQuery.isError;

    if (isLoading || isError) {
        return (
            <Menu
                open={menuPosition !== null}
                anchorPosition={menuPosition!}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem disabled>{isLoading ? 'Loading' : 'Unavailable'}</MenuItem>
            </Menu>
        );
    }

    return (
        <Dialog open={dialogOpen}>
            <Menu
                open={menuPosition !== null}
                anchorPosition={menuPosition!}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                {edgeItem && <EdgeMenuItems id={edgeItem.id} pathfindingFilters={pathfindingFilters} />}

                {nodeItem && <NodeMenuItems exploreParams={exploreParams} objectId={nodeItem.objectId} />}

                {nodeItem &&
                    isAssetGroupEnabled && [
                        <AssetGroupMenuItem
                            key={tierZeroAssetGroup!.id}
                            assetGroupId={tierZeroAssetGroup!.id}
                            assetGroupName={tierZeroAssetGroup!.name}
                            onAddNode={handleAddNode}
                            removeNodePath={`/zone-management/details/tier/${tierZeroAssetGroup!.id}`}
                            isCurrentMember={nodeItem!.isTierZero}
                            onShowConfirmation={() => {
                                setDialogOpen(true);
                            }}
                            onCancelConfirmation={() => {
                                setDialogOpen(false);
                            }}
                            showConfirmationOnAdd
                            confirmationOnAddMessage={`Are you sure you want to add this node to ${tierZeroAssetGroup!.name}? This action will initiate an analysis run to update group membership.`}
                        />,
                        <AssetGroupMenuItem
                            key={ownedAssetGroup!.id}
                            assetGroupId={ownedAssetGroup!.id}
                            assetGroupName={ownedAssetGroup!.name}
                            onAddNode={handleAddNode}
                            removeNodePath={`/zone-management/details/label/${ownedAssetGroup!.id}`}
                            isCurrentMember={nodeItem!.isOwnedObject}
                        />,
                    ]}
                <CopyMenuItems selectedItem={selectedItemQuery.data} />
            </Menu>
        </Dialog>
    );
};

export default ContextMenu;
