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

import { faCaretRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { MenuItem, Tooltip, TooltipProps, styled, tooltipClasses } from '@mui/material';
import { useNotifications } from 'bh-shared-ui';
import { useAppSelector } from 'src/store';

const StyledTooltip = styled(({ className, ...props }: TooltipProps) => (
    <Tooltip {...props} classes={{ popper: className }} />
))(({ theme }) => ({
    [`& .${tooltipClasses.tooltip}`]: {
        color: 'black',
        backgroundColor: theme.palette.common.white,
        padding: 0,
        paddingTop: '0.5rem',
        paddingBottom: '0.5rem',
        boxShadow: theme.shadows[8],
        marginLeft: '2px !important',
    },
}));

const CopyMenuItem = () => {
    const { addNotification } = useNotifications();

    const selectedNode = useAppSelector((state) => state.entityinfo.selectedNode);

    const handleCopyDisplayName = () => {
        if (selectedNode) {
            navigator.clipboard.writeText(selectedNode.name);
            addNotification(`Display name copied to clipboard`, 'copyToClipboard');
        }
    };

    const handleCopyObjectId = () => {
        if (selectedNode) {
            navigator.clipboard.writeText(selectedNode.id);
            addNotification(`Object ID name copied to clipboard`, 'copyToClipboard');
        }
    };

    const handleCopyCypher = () => {
        if (selectedNode) {
            const cypher = `MATCH (n:${selectedNode.type}) WHERE n.objectid = '${selectedNode.id}' RETURN n`;
            navigator.clipboard.writeText(cypher);
            addNotification(`Cypher copied to clipboard`, 'copyToClipboard');
        }
    };

    return (
        <div>
            <StyledTooltip
                placement='right'
                title={
                    <>
                        <MenuItem onClick={handleCopyDisplayName}>Display Name</MenuItem>
                        <MenuItem onClick={handleCopyObjectId}>Object ID</MenuItem>
                        <MenuItem onClick={handleCopyCypher}>Cypher</MenuItem>
                    </>
                }>
                <MenuItem sx={{ justifyContent: 'space-between' }} onClick={(e) => e.stopPropagation()}>
                    Copy <FontAwesomeIcon icon={faCaretRight} />
                </MenuItem>
            </StyledTooltip>
        </div>
    );
};

export default CopyMenuItem;
