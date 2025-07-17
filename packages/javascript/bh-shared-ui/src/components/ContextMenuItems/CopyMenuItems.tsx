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
import { MenuItem, Tooltip, TooltipProps, styled, tooltipClasses } from '@mui/material';
import { FC } from 'react';
import { ItemResponse, NodeResponse, isNode } from '../../hooks';
import { useNotifications } from '../../providers';

export const StyledTooltip = styled(({ className, ...props }: TooltipProps) => (
    <Tooltip {...props} classes={{ popper: className }} />
))(({ theme }) => ({
    [`& .${tooltipClasses.tooltip}`]: {
        color: theme.palette.text.primary,
        backgroundColor: theme.palette.background.paper,
        padding: 0,
        paddingTop: '0.5rem',
        paddingBottom: '0.5rem',
        boxShadow: theme.shadows[8],
        marginLeft: '2px !important',
    },
}));

export const CopyMenuItems: FC<{ selectedItem: ItemResponse }> = ({ selectedItem }) => {
    const { addNotification } = useNotifications();

    if (!selectedItem) {
        return null;
    }

    const isNodeSelected = isNode(selectedItem);

    const handleCopyName = () => {
        navigator.clipboard.writeText(selectedItem.label);
        addNotification(`Name copied to clipboard`, 'copyToClipboard');
    };

    const handleCopyObjectId = () => {
        navigator.clipboard.writeText((selectedItem as NodeResponse).objectId);
        addNotification(`Object ID copied to clipboard`, 'copyToClipboard');
    };

    const handleCopyCypher = () => {
        const cypher = `MATCH (n:${selectedItem.kind}) WHERE n.objectid = '${(selectedItem as NodeResponse).objectId}' RETURN n`;
        navigator.clipboard.writeText(cypher);
        addNotification('Cypher copied to clipboard', 'copyToClipboard');
    };

    return (
        <StyledTooltip
            placement='right'
            title={
                <>
                    <MenuItem onClick={handleCopyName}>Name</MenuItem>
                    {isNodeSelected && <MenuItem onClick={handleCopyObjectId}>Object ID</MenuItem>}
                    {isNodeSelected && <MenuItem onClick={handleCopyCypher}>Cypher</MenuItem>}
                </>
            }>
            <MenuItem sx={{ justifyContent: 'space-between', minWidth: '8rem' }} onClick={(e) => e.stopPropagation()}>
                Copy <FontAwesomeIcon icon={faCaretRight} />
            </MenuItem>
        </StyledTooltip>
    );
};
