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

import { faCaretRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { MenuItem, MenuSub, MenuSubContent, MenuSubTrigger } from 'doodle-ui';
import { NodeResponse, useExploreSelectedItem } from '../../../hooks';
import { useNotifications } from '../../../providers';

const CopyMenuItem = () => {
    const { addNotification } = useNotifications();

    const { selectedItemQuery } = useExploreSelectedItem();

    const handleCopyName = () => {
        if (selectedItemQuery.data) {
            navigator.clipboard.writeText(selectedItemQuery.data.label);
            addNotification(`Name copied to clipboard`, 'copyToClipboard');
        }
    };

    const handleCopyObjectId = () => {
        if (selectedItemQuery.data) {
            navigator.clipboard.writeText((selectedItemQuery.data as NodeResponse).objectId);
            addNotification(`Object ID copied to clipboard`, 'copyToClipboard');
        }
    };

    const handleCopyCypher = () => {
        if (selectedItemQuery.data) {
            const cypher = `MATCH (n:${selectedItemQuery.data.kind}) WHERE n.objectid = '${(selectedItemQuery.data as NodeResponse).objectId}' RETURN n`;
            navigator.clipboard.writeText(cypher);
            addNotification(`Cypher copied to clipboard`, 'copyToClipboard');
        }
    };

    return (
        <MenuSub>
            <MenuSubTrigger>
                Copy <FontAwesomeIcon icon={faCaretRight} />
            </MenuSubTrigger>
            <MenuSubContent>
                <MenuItem onSelect={handleCopyName}>Name</MenuItem>
                <MenuItem onSelect={handleCopyObjectId}>Object ID</MenuItem>
                <MenuItem onSelect={handleCopyCypher}>Cypher</MenuItem>
            </MenuSubContent>
        </MenuSub>
    );
};

export default CopyMenuItem;
