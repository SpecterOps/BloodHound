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

import { MenuItem, MenuSub, MenuSubContent, MenuSubTrigger } from 'doodle-ui';

import { NodeDetails } from 'js-client-library';
import { useExploreSelectedItem } from '../../../hooks';
import { usePrimaryKind } from '../../../hooks/usePrimaryKind';
import { useNotifications } from '../../../providers';
import { escapeCypherString } from '../../../utils/cypher';

const CopyMenuItem = () => {
    const { addNotification } = useNotifications();
    const { selectedItemQuery } = useExploreSelectedItem();
    const nodeInfo = selectedItemQuery.data as NodeDetails | undefined;

    const primaryKind = usePrimaryKind(nodeInfo?.kinds || []);

    const handleCopyName = () => {
        if (nodeInfo) {
            navigator.clipboard.writeText(nodeInfo.properties.name || nodeInfo.properties.objectid || '');
            addNotification(`Name copied to clipboard`, 'copyToClipboard');
        }
    };

    const handleCopyObjectId = () => {
        if (nodeInfo) {
            navigator.clipboard.writeText(nodeInfo.properties.objectid || '');
            addNotification(`Object ID copied to clipboard`, 'copyToClipboard');
        }
    };

    const handleCopyCypher = () => {
        if (nodeInfo) {
            const cypher = `MATCH (n:${primaryKind}) WHERE n.objectid = ${escapeCypherString(nodeInfo.properties.objectid || '')} RETURN n`;
            navigator.clipboard.writeText(cypher);
            addNotification(`Cypher copied to clipboard`, 'copyToClipboard');
        }
    };

    return (
        <MenuSub>
            <MenuSubTrigger>Copy</MenuSubTrigger>
            <MenuSubContent
                className='max-h-[var(--radix-dropdown-menu-content-available-height)] overflow-y-auto'
                aria-label='Copy options'>
                <MenuItem onSelect={handleCopyName}>Name</MenuItem>
                <MenuItem onSelect={handleCopyObjectId}>Object ID</MenuItem>
                <MenuItem onSelect={handleCopyCypher}>Cypher</MenuItem>
            </MenuSubContent>
        </MenuSub>
    );
};

export default CopyMenuItem;
