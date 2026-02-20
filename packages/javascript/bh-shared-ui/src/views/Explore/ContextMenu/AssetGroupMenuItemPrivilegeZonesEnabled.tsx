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

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
} from '@bloodhoundenterprise/doodleui';
import { MenuItem } from '@mui/material';
import { FC, useState } from 'react';
import { useMutation } from 'react-query';
import { Link } from 'react-router-dom';

import type { AssetGroupTag, CreateSelectorRequest } from 'js-client-library';
import { useExploreSelectedItem, usePermissions, useTagsQuery, type ItemResponse } from '../../../hooks';
import { useNotifications } from '../../../providers';
import { Permission, apiClient } from '../../../utils';

const ConfirmNodeChangesDialog: FC<{
    dialogContent: string;
    disableAccept: boolean;
    onAccept: () => void;
    onCancel: () => void;
    open: boolean;
}> = ({ dialogContent, disableAccept, onAccept, onCancel, open }) => {
    return (
        <Dialog open={open}>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle>Confirm Selection</DialogTitle>
                    <DialogDescription>{dialogContent}</DialogDescription>
                    <DialogActions>
                        <Button variant='tertiary' onClick={onCancel}>
                            Cancel
                        </Button>
                        <Button variant='primary' onClick={onAccept} disabled={disableAccept}>
                            Ok
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export const AssetGroupMenuItem: FC<{
    addNodePayload: CreateSelectorRequest;
    isCurrentMemberFn: (node: ItemResponse) => boolean;
    removeNodePathFn: (tag: AssetGroupTag) => string;
    showConfirmationOnAdd?: boolean;
    tagIdentifierFn: (tags: AssetGroupTag[]) => AssetGroupTag | undefined;
}> = ({ addNodePayload, isCurrentMemberFn, removeNodePathFn, showConfirmationOnAdd = false, tagIdentifierFn }) => {
    const [dialogOpen, setDialogOpen] = useState(false);
    const { addNotification } = useNotifications();
    const { selectedItemQuery } = useExploreSelectedItem();
    const { checkPermission } = usePermissions();
    const { data: tags, isLoading, isError } = useTagsQuery();
    const assetGroupTag = tags ? tagIdentifierFn(tags) : undefined;

    const closeDialog = () => setDialogOpen(false);
    const openDialog = () => setDialogOpen(true);

    const { mutate: createRule, isLoading: isMutationLoading } = useMutation({
        mutationFn: () => {
            if (!assetGroupTag) {
                return Promise.reject(new Error('Asset group tag not found'));
            }
            return apiClient.createAssetGroupTagSelector(assetGroupTag?.id ?? '', addNodePayload);
        },
        // Query cache is not invalidated because API call resolution does not guarantee database has been updated
        onSuccess: () => addNotification('Node successfully added.', 'AssetGroupUpdateSuccess'),
        onError: (error: any) => {
            console.error(error);
            addNotification('An error occurred when adding node', 'AssetGroupUpdateError');
        },
        onSettled: closeDialog,
    });

    const createRuleAction = showConfirmationOnAdd ? () => openDialog() : () => createRule();
    const hasPermission = checkPermission(Permission.GRAPH_DB_WRITE);

    // Is the selected node already a member of tier zero or owned?
    const isCurrentMember = isCurrentMemberFn(selectedItemQuery.data);

    // Don't render anything if the user doesn't have permission or the asset group doesn't exist
    if (!hasPermission) {
        return null;
    }

    // Show a loading state until the query is resolved
    if (isLoading) {
        return <MenuItem disabled>Loading</MenuItem>;
    }

    // Show an error state if the query failed
    if (isError) {
        return <MenuItem disabled>Unavailable</MenuItem>;
    }

    // If everything is loaded but there are no tags, there is nothing to render
    if (!assetGroupTag) {
        return null;
    }

    // If selected node is already a member, navigate to the asset group details page for removal
    if (isCurrentMember) {
        return (
            <MenuItem component={Link} to={removeNodePathFn(assetGroupTag)}>
                Remove from {assetGroupTag.name}
            </MenuItem>
        );
    }

    return (
        <>
            <MenuItem onClick={createRuleAction}>Add to {assetGroupTag.name}</MenuItem>

            {showConfirmationOnAdd && (
                <ConfirmNodeChangesDialog
                    dialogContent={`Are you sure you want to add this node to ${assetGroupTag.name}? This action will initiate an analysis run to update zone membership.`}
                    disableAccept={isMutationLoading}
                    onAccept={createRule}
                    onCancel={closeDialog}
                    open={dialogOpen}
                />
            )}
        </>
    );
};
