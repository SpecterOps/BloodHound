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
import { NodeResponse, apiClient, useAppNavigate, useExploreGraph, useExploreSelectedItem, useNotifications } from 'bh-shared-ui';
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { selectTierZeroAssetGroupId, selectOwnedAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';

const AssetGroupMenuItem: FC<{ assetGroupId: number; assetGroupName: string }> = ({ assetGroupId, assetGroupName }) => {
    const { addNotification } = useNotifications();
    const { refetch } = useExploreGraph();
    const navigate = useAppNavigate();
    const queryClient = useQueryClient();

    const [open, setOpen] = useState(false);

    const selectedNode = useExploreSelectedItem().selectedItemQuery.data as NodeResponse;
    const tierZeroAssetGroupId = useAppSelector(selectTierZeroAssetGroupId);
    const ownedAssetGroupId = useAppSelector(selectOwnedAssetGroupId);

    const isMenuItemForTierZero = assetGroupId === tierZeroAssetGroupId;

    const assetGroupTag = new Map([
        [tierZeroAssetGroupId, 'Tag_Tier_Zero'],
        [ownedAssetGroupId, 'Tag_Owned'],
    ]).get(assetGroupId);

    const mutation = useMutation({
        mutationFn: (node: NodeResponse) => {
            return apiClient.createAssetGroupTagSelector(assetGroupId, {
                name: node.label ?? node.objectId,
                seeds: [{
                    type: SeedTypeObjectId,
                    value: node.objectId,
                }],
            });
        },
        onSuccess: (data: any, node: NodeResponse) => {
            queryClient.invalidateQueries(['check-tier-node', assetGroupTag, node.objectId]);
            refetch();
            addNotification('Update successful.', 'AssetGroupUpdateSuccess');
        },
        onError: (error: any) => {
            console.error(error);
            addNotification('Unknown error, group was not updated', 'AssetGroupUpdateError');
        },
    });

    const { data: taggedNode } = useQuery(['check-tier-node', assetGroupTag, selectedNode?.objectId], () =>
        apiClient
            .cypherSearch(`MATCH (n:${assetGroupTag}) WHERE n.objectid = '${selectedNode?.objectId}' RETURN n LIMIT 1`)
            .then((res) => { let r = res.data.data?.nodes; for (let n in r) return r[n]; })
    );

    const handleAddToAssetGroup = () => {
        if (selectedNode && 'objectId' in selectedNode) {
            mutation.mutate(selectedNode);
        }
    };

    const handleRemoveFromAssetGroup = () => {
        if (selectedNode && 'objectId' in selectedNode) {
            navigate(`/tier-management/details/tag/${assetGroupId}`);
        }
    };

    const handleOpenConfirmation = (e: React.MouseEvent<HTMLLIElement>) => {
        e.stopPropagation();
        setOpen(true);
    };

    const handleCloseConfirmation = () => {
        setOpen(false);
    };

    // unsupported type
    if (!assetGroupTag) {
        return null;
    }

    // selected node is not a member of the group
    if (!taggedNode) {
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
                        dialogContent={`Are you sure you want to add this node to ${assetGroupName}? This action will initiate an analysis run to update group membership.`}
                    />
                ) : null}
            </>
        );
    } else {
        return (
            <MenuItem onClick={handleRemoveFromAssetGroup}>
                Remove from {assetGroupName}
            </MenuItem>
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
