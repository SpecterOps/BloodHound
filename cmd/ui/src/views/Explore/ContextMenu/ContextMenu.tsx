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

import { CopyMenuItems, EdgeMenuItems, NodeMenuItems, type MousePosition, type PathfindingFilters } from 'bh-shared-ui';
import { type FC } from 'react';
import { AssetGroupMenuItems } from './AssetGroupMenuItems';

const NAV_MENU_WIDTH = 56;

const getMenuPosition = (position: MousePosition | null) =>
    position
        ? {
              left: position.mouseX + NAV_MENU_WIDTH,
              top: position.mouseY,
          }
        : null;

const ContextMenu: FC<{
    onClose: () => void;
    pathfindingFilters: PathfindingFilters;
    position: MousePosition | null;
}> = ({ onClose, pathfindingFilters, position }) => {
    const menuPosition = getMenuPosition(position);

    if (menuPosition === null) {
        return null;
    }

    return (
        <Menu open anchorPosition={menuPosition} anchorReference='anchorPosition' onClick={onClose}>
            <EdgeMenuItems pathfindingFilters={pathfindingFilters} />

            <NodeMenuItems />

            <AssetGroupMenuItems />

            <CopyMenuItems />
        </Menu>
    );
};

export default ContextMenu;
