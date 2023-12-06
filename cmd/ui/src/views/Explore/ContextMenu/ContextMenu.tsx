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

import { Menu, MenuItem, Tooltip, TooltipProps, styled, tooltipClasses } from '@mui/material';
import { apiClient, useNotifications } from 'bh-shared-ui';
import { FC, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { destinationNodeSelected, sourceNodeSelected, tabChanged } from 'src/ducks/searchbar/actions';
import { AppState, useAppDispatch } from 'src/store';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCaretRight } from '@fortawesome/free-solid-svg-icons';
import { useMutation, useQuery } from 'react-query';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';

const ContextMenu: FC<{ anchorPosition?: { x: number; y: number } }> = ({ anchorPosition }) => {
    const dispatch = useAppDispatch();

    const [open, setOpen] = useState(false);

    const selectedNode = useSelector((state: AppState) => state.entityinfo.selectedNode);
    const ownedAssetGroupId = useSelector(selectOwnedAssetGroupId);
    const tierZeroAssetGroupId = useSelector(selectTierZeroAssetGroupId);

    useEffect(() => {
        if (anchorPosition) {
            setOpen(true);
        } else {
            setOpen(false);
        }
    }, [anchorPosition]);

    const handleClose = () => {
        setOpen(false);
    };

    const handleSetStartingNode = () => {
        if (selectedNode) {
            dispatch(tabChanged('secondary'));
            dispatch(
                sourceNodeSelected({
                    name: selectedNode.name,
                    objectid: selectedNode.id,
                    type: selectedNode.type,
                })
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
            open={open}
            anchorPosition={{ left: anchorPosition?.x || 0 + 10, top: anchorPosition?.y || 0 }}
            anchorReference='anchorPosition'
            onClick={handleClose}>
            <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>

            <AssetGroupMenuItem assetGroupId={tierZeroAssetGroupId} assetGroupName='tier zero' />
            <AssetGroupMenuItem assetGroupId={ownedAssetGroupId} assetGroupName='owned' />

            <CopyMenuItem />
        </Menu>
    );
};

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

    const selectedNode = useSelector((state: AppState) => state.entityinfo.selectedNode);

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

const AssetGroupMenuItem: FC<{ assetGroupId: string; assetGroupName: string }> = ({ assetGroupId, assetGroupName }) => {
    const { addNotification } = useNotifications();

    const selectedNode = useSelector((state: AppState) => state.entityinfo.selectedNode);

    const mutation = useMutation({
        mutationFn: ({ nodeId, action }: { nodeId: string; action: 'add' | 'remove' }) => {
            return apiClient.updateAssetGroupSelector(assetGroupId, [
                {
                    selector_name: nodeId,
                    sid: nodeId,
                    action,
                },
            ]);
        },
        onSuccess: () => {
            addNotification(
                'Update successful. Please check back later to view updated Asset Group.',
                'AssetGroupUpdateSuccess'
            );
        },
        onError: (error: any) => {
            console.error(error);
            addNotification('Unknown error, group was not updated', 'AssetGroupUpdateError');
        },
    });

    const { data: assetGroupMembers } = useQuery(['listAssetGroupMembers', assetGroupId], () =>
        apiClient
            .listAssetGroupMembers(assetGroupId, undefined, {
                params: {
                    object_id: `object_id=eq:${selectedNode?.id}`,
                },
            })
            .then((res) => res.data.data.members)
    );

    const handleAddToAssetGroup = () => {
        if (selectedNode) {
            mutation.mutate({ nodeId: selectedNode.id, action: 'add' });
        }
    };

    const handleRemoveFromAssetGroup = () => {
        if (selectedNode) {
            mutation.mutate({ nodeId: selectedNode.id, action: 'remove' });
        }
    };

    // error state, data didn't load
    if (!assetGroupMembers) {
        return null;
    }

    // selected node is not a member of the group
    if (assetGroupMembers.length === 0) {
        return <MenuItem onClick={handleAddToAssetGroup}>Add to {assetGroupName}</MenuItem>;
    }

    // selected node is a custom member of the group
    if (assetGroupMembers.length === 1 && assetGroupMembers[0].custom_member) {
        return <MenuItem onClick={handleRemoveFromAssetGroup}>Remove from {assetGroupName}</MenuItem>;
    }
};

export default ContextMenu;
