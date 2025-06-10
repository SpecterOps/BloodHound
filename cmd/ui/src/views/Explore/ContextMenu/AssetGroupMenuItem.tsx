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
import { NodeResponse, apiClient, isNode, useExploreSelectedItem, useNotifications } from 'bh-shared-ui';
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useState } from 'react';
import { useMutation } from 'react-query';
import { Link } from 'react-router-dom';
import { useSigmaExploreGraph } from 'src/hooks/useSigmaExploreGraph';

const AssetGroupMenuItem: FC<{
    assetGroupId: number;
    assetGroupName: string;
    assetGroupTag: string;
    isCurrentMember: boolean;
    showConfirmationOnAdd?: boolean;
    confirmationOnAddMessage?: string;
}> = ({
    assetGroupId,
    assetGroupName,
    assetGroupTag,
    isCurrentMember,
    showConfirmationOnAdd = false,
    confirmationOnAddMessage = '',
}) => {
    const { addNotification } = useNotifications();

    const { refetch } = useSigmaExploreGraph();

    const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);

    const { selectedItemQuery } = useExploreSelectedItem();

    const createAssetGroupTagSelectorMutation = useMutation({
        mutationFn: (node: NodeResponse) => {
            return apiClient.createAssetGroupTagSelector(assetGroupId, {
                name: node.label ?? node.objectId,
                seeds: [
                    {
                        type: SeedTypeObjectId,
                        value: node.objectId,
                    },
                ],
            });
        },
        onSuccess: () => {
            addNotification('Update successful.', 'AssetGroupUpdateSuccess');
        },
        onError: (error: any) => {
            console.error(error);
            addNotification('Unknown error, group was not updated', 'AssetGroupUpdateError');
        },
        onSettled: () => {
            refetch();
            selectedItemQuery.refetch();
        },
    });

    const handleAddToAssetGroup = () => {
        if (isNode(selectedItemQuery.data)) {
            createAssetGroupTagSelectorMutation.mutate(selectedItemQuery.data);
        }
    };

    // unsupported type
    if (!assetGroupTag) {
        return null;
    }

    // selected node is not a member of the group
    if (!isCurrentMember) {
        return (
            <>
                <MenuItem onClick={showConfirmationOnAdd ? () => setConfirmDialogOpen(true) : handleAddToAssetGroup}>
                    Add to {assetGroupName}
                </MenuItem>
                {showConfirmationOnAdd && (
                    <ConfirmNodeChangesDialog
                        handleCancel={() => setConfirmDialogOpen(false)}
                        handleApply={handleAddToAssetGroup}
                        open={confirmDialogOpen}
                        dialogContent={confirmationOnAddMessage}
                    />
                )}
            </>
        );
    } else {
        return (
            <MenuItem component={Link} to={`/tier-management/details/tag/${assetGroupId}`}>
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
