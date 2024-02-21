// Copyright 2023 Specter Ops, Inc.
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

import { FC } from 'react';
import { destinationNodeSelected, sourceNodeSelected, tabChanged } from 'src/ducks/searchbar/actions';
import { useAppDispatch, useAppSelector } from 'src/store';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import AssetGroupMenuItem from './AssetGroupMenuItem';
import CopyMenuItem from './CopyMenuItem';

const ContextMenu: FC<{ contextMenu: { mouseX: number; mouseY: number } | null; handleClose: () => void }> = ({
    contextMenu,
    handleClose,
}) => {
    const dispatch = useAppDispatch();

    const selectedNode = useAppSelector((state) => state.entityinfo.selectedNode);

    const ownedAssetGroupId = useAppSelector(selectOwnedAssetGroupId);
    const tierZeroAssetGroupId = useAppSelector(selectTierZeroAssetGroupId);

    const handleSetStartingNode = () => {
        if (selectedNode) {
            dispatch(tabChanged('secondary'));
            dispatch(
                sourceNodeSelected(
                    {
                        name: selectedNode.name,
                        objectid: selectedNode.id,
                        type: selectedNode.type,
                    },
                    true
                )
            );
        }
    };

    const handleSetEndingNode = () => {
        if (selectedNode) {
            dispatch(tabChanged('secondary'));
            dispatch(
                destinationNodeSelected({
                    name: selectedNode.name,
                    objectid: selectedNode.id,
                    type: selectedNode.type,
                })
            );
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

            <AssetGroupMenuItem assetGroupId={tierZeroAssetGroupId} assetGroupName='High Value' />
            <AssetGroupMenuItem assetGroupId={ownedAssetGroupId} assetGroupName='Owned' />

            <CopyMenuItem />
        </Menu>
    );
};

export default ContextMenu;
