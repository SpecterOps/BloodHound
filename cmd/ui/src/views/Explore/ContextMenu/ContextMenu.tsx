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
    CopyMenuItems,
    EdgeMenuItems,
    NodeMenuItems,
    useContextMenuItems,
    type MousePosition,
    type PathfindingFilters,
} from 'bh-shared-ui';
import { type FC } from 'react';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';
import AssetGroupMenuItem from './AssetGroupMenuItem';

const ContextMenu: FC<{
    onClose: () => void;
    pathfindingFilters: PathfindingFilters;
    position: MousePosition | null;
}> = ({ onClose, pathfindingFilters, position }) => {
    const { asEdgeItem, asNodeItem, exploreParams, isAssetGroupEnabled, menuPosition, selectedItemQuery } =
        useContextMenuItems(position);

    const edgeItem = asEdgeItem(selectedItemQuery);
    const nodeItem = asNodeItem(selectedItemQuery);

    const tierZeroId = useAppSelector(selectTierZeroAssetGroupId);
    const ownedId = useAppSelector(selectOwnedAssetGroupId);

    if (!menuPosition || !(edgeItem || nodeItem)) {
        return null;
    }

    return (
        <Menu open anchorPosition={menuPosition} anchorReference='anchorPosition' onClick={onClose}>
            {edgeItem && <EdgeMenuItems id={edgeItem.id} pathfindingFilters={pathfindingFilters} />}

            {nodeItem && <NodeMenuItems exploreParams={exploreParams} objectId={nodeItem.objectId} />}

            {nodeItem && isAssetGroupEnabled && (
                <>
                    <AssetGroupMenuItem assetGroupId={tierZeroId} assetGroupName='High Value' />
                    <AssetGroupMenuItem assetGroupId={ownedId} assetGroupName='Owned' />
                </>
            )}

            <CopyMenuItems selectedItem={selectedItemQuery.data} />
        </Menu>
    );
};

export default ContextMenu;
