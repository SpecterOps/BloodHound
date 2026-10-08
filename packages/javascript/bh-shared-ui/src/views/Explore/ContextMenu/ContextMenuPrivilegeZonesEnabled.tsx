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

import { MenuItem } from 'doodle-ui';
import {
    AssetGroupTag,
    AssetGroupTagSelectorAutoCertifySeedsOnly,
    CreateSelectorRequest,
    NodeDetails,
    SeedTypeObjectId,
} from 'js-client-library';
import { FC } from 'react';
import {
    getOwnedTag,
    getTierZeroTag,
    isNodeResponse,
    useExploreParams,
    useExploreSelectedItem,
    usePZPathParams,
} from '../../../hooks';
import { AssetGroupMenuItem } from './AssetGroupMenuItemPrivilegeZonesEnabled';
import CopyMenuItem from './CopyMenuItem';
import GraphContextMenu from './GraphContextMenu';

const ContextMenu: FC<{
    contextMenu: { mouseX: number; mouseY: number } | null;
    onClose?: () => void;
}> = ({ contextMenu, onClose = () => {} }) => {
    const { selectedItemQuery } = useExploreSelectedItem();
    const { setExploreParams, primarySearch, secondarySearch } = useExploreParams();
    const { tagDetailsLink } = usePZPathParams();

    const node = selectedItemQuery.data ? (selectedItemQuery.data as NodeDetails) : undefined;

    const ownedPayload: CreateSelectorRequest = {
        name: node?.properties.name ?? node?.properties.objectid ?? '',
        seeds: [
            {
                type: SeedTypeObjectId,
                value: node?.properties.objectid ?? '',
            },
        ],
    };

    const tierZeroPayload: CreateSelectorRequest = {
        ...ownedPayload,
        auto_certify: AssetGroupTagSelectorAutoCertifySeedsOnly,
    };

    const handleSetStartingNode = () => {
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && isNodeResponse(selectedItemData)) {
            const searchType = secondarySearch ? 'pathfinding' : 'node';
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                primarySearch: selectedItemData?.properties.objectid,
            });
        }
    };

    const handleSetEndingNode = () => {
        const selectedItemData = selectedItemQuery.data;
        if (selectedItemData && isNodeResponse(selectedItemData)) {
            const searchType = primarySearch ? 'pathfinding' : 'node';
            setExploreParams({
                exploreSearchTab: 'pathfinding',
                searchType,
                secondarySearch: selectedItemData?.properties.objectid,
            });
        }
    };

    return (
        <GraphContextMenu contextMenu={contextMenu} onClose={onClose}>
            <MenuItem onSelect={handleSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onSelect={handleSetEndingNode}>Set as ending node</MenuItem>

            <AssetGroupMenuItem
                onClose={onClose}
                addNodePayload={tierZeroPayload}
                removeNodePathFn={(tag: AssetGroupTag) => tagDetailsLink(tag.id, 'zones')}
                showConfirmationOnAdd
                tagIdentifierFn={getTierZeroTag}
            />

            <AssetGroupMenuItem
                onClose={onClose}
                addNodePayload={ownedPayload}
                removeNodePathFn={(tag: AssetGroupTag) => tagDetailsLink(tag.id, 'labels')}
                tagIdentifierFn={getOwnedTag}
            />

            <CopyMenuItem />
        </GraphContextMenu>
    );
};

export default ContextMenu;
