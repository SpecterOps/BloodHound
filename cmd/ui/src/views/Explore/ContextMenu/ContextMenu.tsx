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
    const { exploreSearchTab } = useExploreParams();

    const { checkPermission } = usePermissions();
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');
    const tierZeroId = useAppSelector(selectTierZeroAssetGroupId);
    const ownedId = useAppSelector(selectOwnedAssetGroupId);

    if (!contextMenu || !selectedItemQuery.data) {
        return null;
    }

    const isEdgeSelected =
        selectedItemType === 'edge' && exploreSearchTab === 'pathfinding' && selectedItemQuery.data.id?.includes('_');
    const isNodeSelected = selectedItemType === 'node';
    const isAssetGroupShown = !tierFlag?.enabled && checkPermission(Permission.GRAPH_DB_WRITE);

    return (
        <Menu open anchorPosition={getPosition(contextMenu)} anchorReference='anchorPosition' onClick={handleClose}>
            {isEdgeSelected && <EdgeMenuItems id={selectedItemQuery.data.id} pathfindingFilters={pathfindingFilters} />}

            {isNodeSelected && <NodeMenuItems objectId={(selectedItemQuery.data as NodeResponse).objectId} />}

            {isNodeSelected && isAssetGroupShown && (
                <>
                    <AssetGroupMenuItem assetGroupId={tierZeroId} assetGroupName='High Value' />
                    <AssetGroupMenuItem assetGroupId={ownedId} assetGroupName='Owned' />
                </>
            )}

            <CopyMenuItem />
        </Menu>
    );
};

export default ContextMenu;
