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
import { AssetGroupTagSelectorAutoCertifySeedsOnly, SeedTypeObjectId } from 'js-client-library';
import { FC, useState } from 'react';
import { useMutation } from 'react-query';
import {
    NodeResponse,
    isNode,
    useExploreParams,
    useExploreSelectedItem,
    usePermissions,
    useTagsQuery,
} from '../../../hooks';
import { useNotifications } from '../../../providers';
import { detailsPath, labelsPath, privilegeZonesPath, zonesPath } from '../../../routes';
import { Permission, apiClient } from '../../../utils';
import AssetGroupMenuItem from './AssetGroupMenuItemPrivilegeZonesEnabled';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    onClose?: () => void;
}> = ({ contextMenu, onClose = () => {} }) => {
    const [dialogOpen, setDialogOpen] = useState(false);

    const { selectedItemQuery } = useExploreSelectedItem();
    const { setExploreParams, primarySearch, secondarySearch } = useExploreParams();
    const getAssetGroupTagsQuery = useTagsQuery();
    const { checkPermission } = usePermissions();
    const { addNotification } = useNotifications();

    const createAssetGroupTagSelectorMutation = useMutation({
        mutationFn: ({ assetGroupId, node }: { assetGroupId: string | number; node: NodeResponse }) => {
            return apiClient.createAssetGroupTagSelector(assetGroupId, {
                name: node.label ?? node.objectId,
                auto_certify: AssetGroupTagSelectorAutoCertifySeedsOnly,
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
        if (!createAssetGroupTagSelectorMutation.isLoading) {
            createAssetGroupTagSelectorMutation.mutate({
                assetGroupId,
                node: selectedItemQuery.data as NodeResponse,
            });
        }
    };

    if (getAssetGroupTagsQuery.isLoading || selectedItemQuery.isLoading) {
        return (
            <Menu
                open={contextMenu !== null}
                anchorPosition={{ left: contextMenu?.mouseX || 0, top: contextMenu?.mouseY || 0 }}
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
                anchorPosition={{ left: contextMenu?.mouseX || 0, top: contextMenu?.mouseY || 0 }}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem disabled>Unavailable</MenuItem>
            </Menu>
        );
    }

    const assetGroupMenuItems: JSX.Element[] = [];
    if (checkPermission(Permission.GRAPH_DB_WRITE)) {
        const tierZeroAssetGroup = getAssetGroupTagsQuery.data?.find((value) => {
            return value.position === 1;
        });

        if (tierZeroAssetGroup) {
            assetGroupMenuItems.push(
                <AssetGroupMenuItem
                    key={tierZeroAssetGroup.id}
                    disableAddNode={createAssetGroupTagSelectorMutation.isLoading}
                    assetGroupId={tierZeroAssetGroup.id}
                    assetGroupName={tierZeroAssetGroup.name}
                    onAddNode={handleAddNode}
                    removeNodePath={`/${privilegeZonesPath}/${zonesPath}/${tierZeroAssetGroup.id}/${detailsPath}`}
                    isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isTierZero}
                    onShowConfirmation={() => {
                        setDialogOpen(true);
                    }}
                    onCancelConfirmation={() => {
                        setDialogOpen(false);
                    }}
                    showConfirmationOnAdd
                    confirmationOnAddMessage={`Are you sure you want to add this node to ${tierZeroAssetGroup.name}? This action will initiate an analysis run to update group membership.`}
                />
            );
        }

        const ownedAssetGroup = getAssetGroupTagsQuery.data?.find((value) => {
            return value.type === 3;
        });
        if (ownedAssetGroup) {
            assetGroupMenuItems.push(
                <AssetGroupMenuItem
                    key={ownedAssetGroup.id}
                    disableAddNode={createAssetGroupTagSelectorMutation.isLoading}
                    assetGroupId={ownedAssetGroup.id}
                    assetGroupName={ownedAssetGroup.name}
                    onAddNode={handleAddNode}
                    removeNodePath={`/${privilegeZonesPath}/${labelsPath}/${ownedAssetGroup.id}/${detailsPath}`}
                    isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isOwnedObject}
                />
            );
        }
    }

    return (
        <Dialog open={dialogOpen}>
            <Menu
                open={contextMenu !== null}
                anchorPosition={{ left: contextMenu?.mouseX || 0, top: contextMenu?.mouseY || 0 }}
                anchorReference='anchorPosition'
                onClick={onClose}
                keepMounted>
                <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
                <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>
                {assetGroupMenuItems}
                <CopyMenuItem />
            </Menu>
        </Dialog>
    );
};

export default ContextMenu;
