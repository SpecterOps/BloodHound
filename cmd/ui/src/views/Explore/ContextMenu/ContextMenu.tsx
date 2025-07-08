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

import { Menu } from '@mui/material';

import {
    EdgeMenuItems,
    NodeMenuItems,
    Permission,
    useExploreParams,
    useExploreSelectedItem,
    useFeatureFlag,
    usePermissions,
    type NodeResponse,
    type PathfindingFilters,
} from 'bh-shared-ui';
import { type FC } from 'react';
import type { Coordinates } from 'sigma/types';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';
import AssetGroupMenuItem from './AssetGroupMenuItem';
import CopyMenuItem from './CopyMenuItem';

const NAV_MENU_WIDTH = 56;

/** Return position to show context menu, with nav menu offset */
const getPosition = (coordinates: Coordinates) => ({
    left: coordinates.x + NAV_MENU_WIDTH,
    top: coordinates.y,
});

const ContextMenu: FC<{
    contextMenu: Coordinates | null;
    handleClose: () => void;
    pathfindingFilters: PathfindingFilters;
}> = ({ contextMenu, handleClose, pathfindingFilters }) => {
    const { selectedItemQuery, selectedItemType } = useExploreSelectedItem();
    const exploreParams = useExploreParams();
    const { exploreSearchTab } = exploreParams;

    const { checkPermission } = usePermissions();
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');
    const tierZeroId = useAppSelector(selectTierZeroAssetGroupId);
    const ownedId = useAppSelector(selectOwnedAssetGroupId);

    if (!contextMenu || !selectedItemQuery.data) {
        return null;
    }

    const menuItems = [];

    // Add pathfinding edge filtering if edge selected on pathfinding tab
    if (selectedItemType === 'edge' && exploreSearchTab === 'pathfinding' && selectedItemQuery.data.id?.includes('_')) {
        menuItems.push(
            <EdgeMenuItems key='edge-items' id={selectedItemQuery.data.id} pathfindingFilters={pathfindingFilters} />
        );
    }

    // Add node options and asset group options
    if (selectedItemType === 'node') {
        menuItems.push(
            <NodeMenuItems
                key='node-items'
                exploreParams={exploreParams}
                objectId={(selectedItemQuery.data as NodeResponse).objectId}
            />
        );

        if (!tierFlag?.enabled && checkPermission(Permission.GRAPH_DB_WRITE)) {
            menuItems.push(
                <AssetGroupMenuItem key='node-tier-zero' assetGroupId={tierZeroId} assetGroupName='High Value' />
            );
            menuItems.push(<AssetGroupMenuItem key='node-owned' assetGroupId={ownedId} assetGroupName='Owned' />);
        }
    }

    menuItems.push(<CopyMenuItem key='copy-items' />);

    return (
        <Menu open anchorPosition={getPosition(contextMenu)} anchorReference='anchorPosition' onClick={handleClose}>
            {menuItems}
        </Menu>
    );
};

export default ContextMenu;
