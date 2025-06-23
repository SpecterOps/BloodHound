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
    NodeResponse,
    Permission,
    apiClient,
    isNode,
    useExploreParams,
    useExploreSelectedItem,
    useNotifications,
    usePermissions,
} from 'bh-shared-ui';
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import { type Coordinates } from 'sigma/types';
import AssetGroupMenuItem from './AssetGroupMenuItemZoneManagementEnabled';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{
    contextMenu: Coordinates | null;
    onClose?: () => void;
}> = ({ contextMenu, onClose = () => {} }) => {
    const { addNotification } = useNotifications();

    const { checkPermission } = usePermissions();

    const { selectedItemQuery } = useExploreSelectedItem();

    const { setExploreParams, primarySearch, secondarySearch } = useExploreParams();

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

    const handleSetStartingNode = () => {
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && isNode(selectedItemData)) {
            const searchType = secondarySearch ? 'pathfinding' : 'node';
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                primarySearch: selectedItemData?.objectId,
            });
        }
    };

    const [dialogOpen, setDialogOpen] = useState(false);

    const handleSetEndingNode = () => {
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && isNode(selectedItemData)) {
            const searchType = primarySearch ? 'pathfinding' : 'node';
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                secondarySearch: selectedItemData?.objectId,
            });
        }
    };

    const handleAddNode = (assetGroupId: string | number) => {
        createAssetGroupTagSelectorMutation.mutate(
            {
                assetGroupId,
                node: selectedItemQuery.data as NodeResponse,
            },
            {
                onSettled: () => {
                    setDialogOpen(false);
                },
            }
        );
    };

    const tierZeroAssetGroup = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.position === 1;
    });

    const ownedAssetGroup = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.type === 3;
    });

    if (getAssetGroupTagsQuery.isLoading || selectedItemQuery.isLoading) {
        return (
            <Menu
                open={contextMenu !== null}
                anchorPosition={{ left: contextMenu?.x || 0, top: contextMenu?.y || 0 }}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem disabled>Loading</MenuItem>
            </Menu>
        );
    }

    if (getAssetGroupTagsQuery.isError || selectedItemQuery.isError) {
        return (
            <Menu
                open={contextMenu !== null}
                anchorPosition={{ left: contextMenu?.x || 0, top: contextMenu?.y || 0 }}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem disabled>Unavailable</MenuItem>
            </Menu>
        );
    }

    return (
        <Dialog open={dialogOpen}>
            <Menu
                open={contextMenu !== null}
                anchorPosition={{ left: contextMenu?.x || 0, top: contextMenu?.y || 0 }}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
                <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>
                {checkPermission(Permission.GRAPH_DB_WRITE) && [
                    <AssetGroupMenuItem
                        key={tierZeroAssetGroup!.id}
                        assetGroupId={tierZeroAssetGroup!.id}
                        assetGroupName={tierZeroAssetGroup!.name}
                        onAddNode={handleAddNode}
                        removeNodePath={`/zone-management/details/tier/${tierZeroAssetGroup!.id}`}
                        isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isTierZero}
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
                        isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isOwnedObject}
                    />,
                ]}
                <CopyMenuItem />
            </Menu>
        </Dialog>
    );
};

export default ContextMenu;
