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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Dialog, DialogActions, DialogContent, DialogTitle, MenuItem } from '@mui/material';
import { NodeResponse, apiClient, useExploreGraph, useExploreSelectedItem, useNotifications } from 'bh-shared-ui';
import { FC, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import { selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';

const AssetGroupMenuItem: FC<{ assetGroupId: number; assetGroupName: string }> = ({ assetGroupId, assetGroupName }) => {
    const { addNotification } = useNotifications();
    const { refetch } = useExploreGraph();

    const [open, setOpen] = useState(false);

    const { selectedItemQuery } = useExploreSelectedItem();
    const tierZeroAssetGroupId = useAppSelector(selectTierZeroAssetGroupId);

    const isMenuItemForTierZero = assetGroupId === tierZeroAssetGroupId;

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
            refetch();
            addNotification('Update successful.', 'AssetGroupUpdateSuccess');
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
                    object_id: `object_id=eq:${(selectedItemQuery.data as NodeResponse)?.objectId}`,
                },
            })
            .then((res) => res.data.data?.members)
    );

    const handleAddToAssetGroup = () => {
        if (selectedItemQuery.data && 'objectId' in selectedItemQuery.data) {
            mutation.mutate({ nodeId: selectedItemQuery.data.objectId, action: 'add' });
        }
    };

    const handleRemoveFromAssetGroup = () => {
        if (selectedItemQuery.data && 'objectId' in selectedItemQuery.data) {
            mutation.mutate({ nodeId: selectedItemQuery.data.objectId, action: 'remove' });
        }
    };

    const handleOpenConfirmation = (e: React.MouseEvent<HTMLLIElement>) => {
        e.stopPropagation();
        setOpen(true);
    };

    const handleCloseConfirmation = () => {
        setOpen(false);
    };

    // error state, data didn't load
    if (!assetGroupMembers) {
        return null;
    }

    // selected node is not a member of the group
    if (assetGroupMembers.length === 0) {
        return (
            <>
                <MenuItem onClick={isMenuItemForTierZero ? handleOpenConfirmation : handleAddToAssetGroup}>
                    Add to {assetGroupName}
                </MenuItem>
                {isMenuItemForTierZero ? (
                    <ConfirmNodeChangesDialog
                        handleCancel={handleCloseConfirmation}
                        handleApply={handleAddToAssetGroup}
                        open={open}
                        dialogContent={`Are you sure you want to add this node to ${assetGroupName}? This action will initiate an analysis run to update zone membership.`}
                    />
                ) : null}
            </>
        );
    }

    // selected node is a custom member of the group
    if (assetGroupMembers.length === 1 && assetGroupMembers[0].custom_member) {
        return (
            <>
                <MenuItem onClick={isMenuItemForTierZero ? handleOpenConfirmation : handleRemoveFromAssetGroup}>
                    Remove from {assetGroupName}
                </MenuItem>
                {isMenuItemForTierZero ? (
                    <ConfirmNodeChangesDialog
                        handleCancel={() => setOpen(false)}
                        handleApply={handleRemoveFromAssetGroup}
                        open={open}
                        dialogContent={`Are you sure you want to remove this node from ${assetGroupName}? This action will initiate an analysis run to update zone membership.`}
                    />
                ) : null}
            </>
        );
    }
};

const ConfirmNodeChangesDialog: FC<{
    open: boolean;
    handleCancel: () => void;
    handleApply: () => void;
    dialogContent: string;
}> = ({ open, handleApply, handleCancel, dialogContent }) => {
    return (
        <Dialog open={open}>
            <DialogTitle>Confirm Selection</DialogTitle>
            <DialogContent>{dialogContent}</DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={handleCancel}>
                    Cancel
                </Button>
                <Button variant='primary' onClick={handleApply}>
                    Ok
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default AssetGroupMenuItem;
