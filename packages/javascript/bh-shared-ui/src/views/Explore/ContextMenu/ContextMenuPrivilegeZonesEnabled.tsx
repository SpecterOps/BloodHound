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
import { FC, useState } from 'react';
import { isNode, useExploreParams, useExploreSelectedItem, useTagsQuery } from '../../../hooks';
import useAssetGroupMenuItems from '../../../hooks/useAssetGroupMenuItems';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    onClose?: () => void;
}> = ({ contextMenu, onClose = () => {} }) => {
    const [dialogOpen, setDialogOpen] = useState(false);

    const { selectedItemQuery } = useExploreSelectedItem();
    const { setExploreParams, primarySearch, secondarySearch } = useExploreParams();
    const getAssetGroupTagsQuery = useTagsQuery();
    const assetGroupMenuItems = useAssetGroupMenuItems(setDialogOpen);

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
                {assetGroupMenuItems.length ? assetGroupMenuItems.map((MenuItem) => MenuItem) : null}
                <CopyMenuItem />
            </Menu>
        </Dialog>
    );
};

export default ContextMenu;
