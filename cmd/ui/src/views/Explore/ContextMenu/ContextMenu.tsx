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
    Permission,
    isEdgeType,
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

type EdgeMenuItemsProps = {
    id: string;
    pathfindingFilters: PathfindingFilters;
};

type NodeMenuItemsProps = {
    objectId: string;
    pathfindingFilters: PathfindingFilters;
};

const NAV_MENU_WIDTH = 56;

const RX_EDGE_TYPE = /^[^_]+_([^_]+)_[^_]+$/;

/** Return position to show context menu, with nav menu offset */
const getPosition = (coordinates: Coordinates) => ({
    left: coordinates.x + NAV_MENU_WIDTH,
    top: coordinates.y,
});

const EdgeMenuItems: FC<EdgeMenuItemsProps> = ({ id, pathfindingFilters }) => {
    const { handleRemoveEdgeType } = pathfindingFilters;

    const edgeType = id.match(RX_EDGE_TYPE)?.[1];

    const filterEdge = () => {
        // edgeType will exist otherwise this method could't be executed
        handleRemoveEdgeType(edgeType!);
    };

    if (!edgeType) {
        return null;
    }

    // Prevent filtering for edge types not found in AllEdgeTypes array
    return isEdgeType(edgeType) ? (
        <MenuItem key='filter-edge' onClick={filterEdge}>
            Filter out Edge
        </MenuItem>
    ) : (
        <MenuItem key='non-filterable' disabled>
            Non-filterable Edge
        </MenuItem>
    );
};

const NodeMenuItems: FC<Omit<NodeMenuItemsProps, 'pathfindingFilters'>> = ({ objectId }) => {
    const { checkPermission } = usePermissions();
    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');

    const tierZeroId = useAppSelector(selectTierZeroAssetGroupId);
    const ownedId = useAppSelector(selectOwnedAssetGroupId);

    return (
        <>
            <MenuItem
                key='starting-node'
                onClick={() =>
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType: secondarySearch ? 'pathfinding' : 'node',
                        primarySearch: objectId,
                    })
                }>
                Set as starting node
            </MenuItem>

            <MenuItem
                key='ending-node'
                onClick={() =>
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType: primarySearch ? 'pathfinding' : 'node',
                        secondarySearch: objectId,
                    })
                }>
                Set as ending node
            </MenuItem>

            {!tierFlag?.enabled && checkPermission(Permission.GRAPH_DB_WRITE) && (
                <>
                    <AssetGroupMenuItem key='tier-zero' assetGroupId={tierZeroId} assetGroupName='High Value' />
                    <AssetGroupMenuItem key='owned' assetGroupId={ownedId} assetGroupName='Owned' />
                </>
            )}
        </>
    );
};

const ContextMenu: FC<{
    contextMenu: Coordinates | null;
    handleClose: () => void;
    pathfindingFilters: PathfindingFilters;
}> = ({ contextMenu, handleClose, pathfindingFilters }) => {
    const { selectedItemQuery, selectedItemType } = useExploreSelectedItem();
    const { exploreSearchTab } = useExploreParams();

    if (!contextMenu || !selectedItemQuery.data) {
        return null;
    }

    const isEdgeSelected =
        selectedItemType === 'edge' && exploreSearchTab === 'pathfinding' && selectedItemQuery.data.id?.includes('_');
    const isNodeSelected = selectedItemType === 'node';

    return (
        <Menu open anchorPosition={getPosition(contextMenu)} anchorReference='anchorPosition' onClick={handleClose}>
            {isEdgeSelected && <EdgeMenuItems id={selectedItemQuery.data.id} pathfindingFilters={pathfindingFilters} />}
            {isNodeSelected && <NodeMenuItems objectId={(selectedItemQuery.data as NodeResponse).objectId} />}
            <CopyMenuItem />
        </Menu>
    );
};

export default ContextMenu;
