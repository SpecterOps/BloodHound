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

import { Permission, useExploreParams, useExploreSelectedItem, usePermissions } from 'bh-shared-ui';
import { FC } from 'react';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';
import AssetGroupMenuItemV2 from './AssetGroupMenuItemV2';
import CopyMenuItemV2 from './CopyMenuItemV2';

const ContextMenuV2: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    handleClose: () => void;
}> = ({ contextMenu, handleClose }) => {
    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();

    const { selectedItemQuery } = useExploreSelectedItem();

    const ownedAssetGroupId = useAppSelector(selectOwnedAssetGroupId);
    const tierZeroAssetGroupId = useAppSelector(selectTierZeroAssetGroupId);

    const { checkPermission } = usePermissions();

    const handleSetStartingNode = () => {
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && 'object_id' in selectedItemData) {
            const searchType = secondarySearch ? 'pathfinding' : 'node';
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                primarySearch: selectedItemData.object_id as string,
            });
        }
    };

    const handleSetEndingNode = () => {
        const searchType = primarySearch ? 'pathfinding' : 'node';
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && 'object_id' in selectedItemData) {
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                secondarySearch: selectedItemData.object_id as string,
            });
        }
    };

    return (
        <Menu
            open={contextMenu !== null}
            anchorPosition={{ left: contextMenu?.mouseX || 0 + 10, top: contextMenu?.mouseY || 0 }}
            anchorReference='anchorPosition'
            onClick={handleClose}>
            <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>

            {checkPermission(Permission.GRAPH_DB_WRITE) && [
                <AssetGroupMenuItemV2
                    key={tierZeroAssetGroupId}
                    assetGroupId={tierZeroAssetGroupId}
                    assetGroupName='High Value'
                />,
                <AssetGroupMenuItemV2
                    key={ownedAssetGroupId}
                    assetGroupId={ownedAssetGroupId}
                    assetGroupName='Owned'
                />,
            ]}
            <CopyMenuItemV2 />
        </Menu>
    );
};

export default ContextMenuV2;
