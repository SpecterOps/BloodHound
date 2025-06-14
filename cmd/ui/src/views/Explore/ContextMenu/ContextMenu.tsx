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

import { Permission, apiClient, isNode, useExploreSelectedItem, usePermissions } from 'bh-shared-ui';
import { FC } from 'react';
import { useQuery } from 'react-query';
import { OWNED_ASSET_GROUP_TAG, TIER_ZERO_ASSET_GROUP_TAG } from '../utils';
import AssetGroupMenuItem from './AssetGroupMenuItem';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    onSetStartingNode?: () => void;
    onSetEndingNode?: () => void;
    onAddNode?: (assetGroupId: string | number) => void;
    onRemoveNode?: () => void;
    onClose?: () => void;
}> = ({
    contextMenu,
    onSetStartingNode = () => {},
    onSetEndingNode = () => {},
    onAddNode = () => {},
    onRemoveNode = () => {},
    onClose = () => {},
}) => {
    const { selectedItemQuery } = useExploreSelectedItem();

    const getAssetGroupTagsQuery = useQuery(['getAssetGroupTags'], ({ signal }) =>
        apiClient.getAssetGroupTags({ signal }).then((res) => res.data.data)
    );

    const { checkPermission } = usePermissions();

    const tierZeroAssetGroupId = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.name === 'Tier Zero';
    })?.id;

    const ownedAssetGroupId = getAssetGroupTagsQuery.data?.tags.find((value) => {
        return value.name === 'Owned';
    })?.id;

    return (
        <Menu
            open={contextMenu !== null}
            anchorPosition={{ left: contextMenu?.mouseX || 0 + 10, top: contextMenu?.mouseY || 0 }}
            anchorReference='anchorPosition'
            onClick={onClose}>
            <MenuItem onClick={onSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onClick={onSetEndingNode}>Set as ending node</MenuItem>
            {checkPermission(Permission.GRAPH_DB_WRITE) && [
                <AssetGroupMenuItem
                    key={tierZeroAssetGroupId}
                    assetGroupId={tierZeroAssetGroupId || Number.NaN}
                    assetGroupName='High Value'
                    assetGroupTag={TIER_ZERO_ASSET_GROUP_TAG}
                    onAddNode={onAddNode}
                    onRemoveNode={onRemoveNode}
                    isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isTierZero}
                    showConfirmationOnAdd
                    confirmationOnAddMessage={`Are you sure you want to add this node to High Value? This action will initiate an analysis run to update group membership.`}
                />,
                <AssetGroupMenuItem
                    key={ownedAssetGroupId}
                    assetGroupId={ownedAssetGroupId || Number.NaN}
                    assetGroupName='Owned'
                    assetGroupTag={OWNED_ASSET_GROUP_TAG}
                    onAddNode={onAddNode}
                    onRemoveNode={onRemoveNode}
                    isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isOwnedObject}
                />,
            ]}
            <CopyMenuItem />
        </Menu>
    );
};

export default ContextMenu;
