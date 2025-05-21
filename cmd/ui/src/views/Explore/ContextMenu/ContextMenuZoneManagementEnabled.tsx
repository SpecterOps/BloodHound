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
import { FC } from 'react';
import { useMutation, useQuery } from 'react-query';
import AssetGroupMenuItem from './AssetGroupMenuItemZoneManagementEnabled';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
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
                primarySearch: selectedItemData?.objectId as string,
            });
        }
    };

    const handleSetEndingNode = () => {
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && isNode(selectedItemData)) {
            const searchType = primarySearch ? 'pathfinding' : 'node';
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                secondarySearch: selectedItemData?.objectId as string,
            });
        }
    };

    const handleAddNode = (assetGroupId: string | number) => {
        createAssetGroupTagSelectorMutation.mutate({
            assetGroupId,
            node: selectedItemQuery.data as NodeResponse,
        });
    };

    const tierZeroAssetGroupId = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.name === 'Tier Zero';
    })?.id;

    const ownedAssetGroupId = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.name === 'Owned';
    })?.id;

    if (getAssetGroupTagsQuery.isLoading || selectedItemQuery.isLoading) {
        return (
            <Menu
                open={contextMenu !== null}
                anchorPosition={{ left: contextMenu?.mouseX || 0 + 10, top: contextMenu?.mouseY || 0 }}
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
                anchorPosition={{ left: contextMenu?.mouseX || 0 + 10, top: contextMenu?.mouseY || 0 }}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem disabled>Unavailable</MenuItem>
            </Menu>
        );
    }

    return (
        <Menu
            open={contextMenu !== null}
            anchorPosition={{ left: contextMenu?.mouseX || 0 + 10, top: contextMenu?.mouseY || 0 }}
            anchorReference='anchorPosition'
            onClick={onClose}
            keepMounted>
            <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>
            {checkPermission(Permission.GRAPH_DB_WRITE) && [
                <AssetGroupMenuItem
                    key={tierZeroAssetGroupId}
                    assetGroupId={tierZeroAssetGroupId || Number.NaN}
                    assetGroupName='High Value'
                    onAddNode={handleAddNode}
                    isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isTierZero}
                    showConfirmationOnAdd
                    confirmationOnAddMessage={`Are you sure you want to add this node to High Value? This action will initiate an analysis run to update group membership.`}
                />,
                <AssetGroupMenuItem
                    key={ownedAssetGroupId}
                    assetGroupId={ownedAssetGroupId || Number.NaN}
                    assetGroupName='Owned'
                    onAddNode={handleAddNode}
                    isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isOwnedObject}
                />,
            ]}
            <CopyMenuItem />
        </Menu>
    );
};

export default ContextMenu;
