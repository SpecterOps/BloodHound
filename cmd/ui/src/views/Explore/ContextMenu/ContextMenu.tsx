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
    isNode,
    // useDeleteEdgeMutation,
    useDeleteNodeMutation,
    useExploreParams,
    useExploreSelectedItem,
    useFeatureFlag,
    usePermissions,
} from 'bh-shared-ui';
import { FC } from 'react';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';
import AssetGroupMenuItem from './AssetGroupMenuItem';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    handleClose: () => void;
}> = ({ contextMenu, handleClose }) => {
    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();
    const { mutateAsync: deleteNode } = useDeleteNodeMutation();
    // const { mutateAsync: deleteEdge } = useDeleteEdgeMutation();
    const { selectedItemQuery, selectedItemType } = useExploreSelectedItem();
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');

    const ownedAssetGroupId = useAppSelector(selectOwnedAssetGroupId);
    const tierZeroAssetGroupId = useAppSelector(selectTierZeroAssetGroupId);

    const { checkPermission } = usePermissions();

    const handleAddNode = () => {
        // TODO: Implement me
        console.log('Add node');
    };

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
        const searchType = primarySearch ? 'pathfinding' : 'node';
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && isNode(selectedItemData)) {
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                secondarySearch: selectedItemData?.objectId as string,
            });
        }
    };

    const handleDeleteNode = async () => {
        if (selectedItemQuery?.data?.id) {
            await deleteNode(selectedItemQuery.data.id);
        }

        console.log(selectedItemQuery.data);
    };

    const isStageClick = selectedItemType === undefined;

    const isNodeClick = selectedItemType === 'node';

    return (
        <Menu
            open={contextMenu !== null}
            anchorPosition={{ left: contextMenu?.mouseX || 0 + 10, top: contextMenu?.mouseY || 0 }}
            anchorReference='anchorPosition'
            onClick={handleClose}>
            {isStageClick && <MenuItem onClick={handleAddNode}>Add a node</MenuItem>}
            {isNodeClick && (
                <>
                    <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
                    <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>
                    <MenuItem onClick={handleDeleteNode}>Delete node</MenuItem>

                    {!tierFlag?.enabled &&
                        checkPermission(Permission.GRAPH_DB_WRITE) && [
                            <AssetGroupMenuItem
                                key={tierZeroAssetGroupId}
                                assetGroupId={tierZeroAssetGroupId}
                                assetGroupName='High Value'
                            />,
                            <AssetGroupMenuItem
                                key={ownedAssetGroupId}
                                assetGroupId={ownedAssetGroupId}
                                assetGroupName='Owned'
                            />,
                        ]}
                    <CopyMenuItem />
                </>
            )}
        </Menu>
    );
};

export default ContextMenu;
