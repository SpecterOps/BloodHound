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
import { FC } from 'react';

import {
    AssetGroupTag,
    AssetGroupTagSelectorAutoCertifySeedsOnly,
    CreateSelectorRequest,
    SeedTypeObjectId,
} from 'js-client-library';
import {
    getIsOwnedTag,
    getIsTierZeroTag,
    isNode,
    isOwnedObject,
    isTierZero,
    useExploreParams,
    useExploreSelectedItem,
    usePZPathParams,
    type NodeResponse,
} from '../../../hooks';
import { AssetGroupMenuItem } from './AssetGroupMenuItemPrivilegeZonesEnabled';
import CopyMenuItem from './CopyMenuItem';

// TODO: Selector -> Rule
// TODO: TagSeleter to

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    onClose?: () => void;
}> = ({ contextMenu, onClose = () => {} }) => {
    const { selectedItemQuery } = useExploreSelectedItem();
    const { setExploreParams, primarySearch, secondarySearch } = useExploreParams();
    const { tagDetailsLink } = usePZPathParams();

    const node = selectedItemQuery.data ? (selectedItemQuery.data as NodeResponse) : undefined;

    const ownedPayload: CreateSelectorRequest = {
        name: node?.label ?? node?.objectId ?? '',
        seeds: [
            {
                type: SeedTypeObjectId,
                value: node?.objectId ?? '',
            },
        ],
    };

    const tierZeroPayload: CreateSelectorRequest = {
        ...ownedPayload,
        auto_certify: AssetGroupTagSelectorAutoCertifySeedsOnly,
    };

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

    return (
        <Menu
            open={contextMenu !== null}
            anchorPosition={{ left: contextMenu?.mouseX || 0, top: contextMenu?.mouseY || 0 }}
            anchorReference='anchorPosition'
            onClick={onClose}
            keepMounted>
            <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>

            <AssetGroupMenuItem
                addNodePayload={tierZeroPayload}
                isCurrentMemberFn={isTierZero}
                removeNodePathFn={(tag: AssetGroupTag) => tagDetailsLink(tag.id, 'zones')}
                showConfirmationOnAdd
                tagIdentifierFn={getIsTierZeroTag}
            />

            <AssetGroupMenuItem
                addNodePayload={ownedPayload}
                isCurrentMemberFn={isOwnedObject}
                removeNodePathFn={(tag: AssetGroupTag) => tagDetailsLink(tag.id, 'labels')}
                tagIdentifierFn={getIsOwnedTag}
            />

            <CopyMenuItem />
        </Menu>
    );
};

export default ContextMenu;
