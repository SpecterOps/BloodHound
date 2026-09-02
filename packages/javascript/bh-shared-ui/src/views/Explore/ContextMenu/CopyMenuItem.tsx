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
import { Menu, MenuItem } from '@mui/material';
import { NodeDetails } from 'js-client-library';
import { KeyboardEvent, MouseEvent, useRef, useState } from 'react';
import { useExploreSelectedItem } from '../../../hooks';
import { usePrimaryKind } from '../../../hooks/usePrimaryKind';
import { useNotifications } from '../../../providers';
import { escapeCypherString } from '../../../utils/cypher';

const CopyMenuItem = () => {
    const { addNotification } = useNotifications();
    const triggerRef = useRef<HTMLLIElement>(null);
    const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

    const { selectedItemQuery } = useExploreSelectedItem();
    const nodeInfo = selectedItemQuery.data as NodeDetails | undefined;

    const primaryKind = usePrimaryKind(nodeInfo?.kinds || []);

    const closeMenu = () => {
        setAnchorEl(null);
        requestAnimationFrame(() => triggerRef.current?.focus());
    };

    const handleOpen = (event: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>) => {
        event.preventDefault();
        event.stopPropagation();
        setAnchorEl(event.currentTarget);
    };

    const handleTriggerKeyDown = (event: KeyboardEvent<HTMLElement>) => {
        if (event.key === 'ArrowRight' || event.key === 'Enter' || event.key === ' ') handleOpen(event);
    };

    const handleCopyName = () => {
        if (nodeInfo) {
            navigator.clipboard.writeText(nodeInfo.properties.name || nodeInfo.properties.objectid || '');
            addNotification(`Name copied to clipboard`, 'copyToClipboard');
        }
        closeMenu();
    };

    const handleCopyObjectId = () => {
        if (nodeInfo) {
            navigator.clipboard.writeText(nodeInfo.properties.objectid || '');
            addNotification(`Object ID copied to clipboard`, 'copyToClipboard');
        }
        closeMenu();
    };

    const handleCopyCypher = () => {
        if (nodeInfo) {
            const cypher = `MATCH (n:${primaryKind}) WHERE n.objectid = ${escapeCypherString(nodeInfo.properties.objectid || '')} RETURN n`;
            navigator.clipboard.writeText(cypher);
            addNotification(`Cypher copied to clipboard`, 'copyToClipboard');
        }
        closeMenu();
    };

    return (
        <>
            <MenuItem
                ref={triggerRef}
                className='justify-between'
                aria-haspopup='menu'
                aria-expanded={Boolean(anchorEl)}
                onClick={handleOpen}
                onKeyDown={handleTriggerKeyDown}>
                Copy <FontAwesomeIcon icon={faCaretRight} />
            </MenuItem>
            <Menu
                anchorEl={anchorEl}
                open={Boolean(anchorEl)}
                onClose={closeMenu}
                anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
                transformOrigin={{ vertical: 'top', horizontal: 'left' }}
                MenuListProps={{
                    'aria-label': 'Copy options',
                    onKeyDown: (event) => {
                        if (event.key === 'Tab' || event.key === 'Escape') return;
                        event.stopPropagation();
                    },
                }}>
                <MenuItem onClick={handleCopyName}>Name</MenuItem>
                <MenuItem onClick={handleCopyObjectId}>Object ID</MenuItem>
                <MenuItem onClick={handleCopyCypher}>Cypher</MenuItem>
            </Menu>
        </>
    );
};

export default CopyMenuItem;
