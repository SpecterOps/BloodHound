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
import { apiClient, isNode, useExploreGraph, useExploreSelectedItem, useNotifications } from 'bh-shared-ui';
import { FC, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import { selectOwnedAssetGroupId, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { useAppSelector } from 'src/store';

const useAssetGroupMembersList = (assetGroupId: number, objectId?: string) => {
    return useQuery({
        queryKey: ['listAssetGroupMembers', assetGroupId],
        queryFn: () =>
            apiClient
                .listAssetGroupMembers(assetGroupId, undefined, {
                    params: { object_id: `object_id=eq:${objectId}` },
                })
                .then((res: any) => res.data.data?.members),
        enabled: objectId !== undefined,
    });
};

const useUpdateAssetGroup = () => {
    const { addNotification } = useNotifications();
    const { refetch } = useExploreGraph();

    return useMutation({
        mutationFn: ({
            assetGroupId,
            nodeId,
            action,
        }: {
            assetGroupId: number;
            nodeId: string;
            action: 'add' | 'remove';
        }) => {
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
};

const getMessage = (name: string, opMessage: string) =>
    `Are you sure you want to ${opMessage} ${name}? This action will initiate an analysis run to update group membership.`;

const ConfirmNodeChangesDialog: FC<{
    assetGroupName: string;
    onApply: () => void;
    onCancel: () => void;
    operation: 'add' | 'remove';
}> = ({ assetGroupName, onApply, onCancel, operation }) => {
    return (
        <Dialog open>
            <DialogTitle>Confirm Selection</DialogTitle>
            <DialogContent>
                {getMessage(assetGroupName, operation == 'add' ? 'add this node to' : 'remove this node from')}
            </DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={onCancel}>
                    Cancel
                </Button>
                <Button variant='primary' onClick={onApply}>
                    Ok
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const AssetGroupMenuItem: FC<{ assetGroupId: number; assetGroupName: string; confirm?: boolean; objectId: string }> = ({
    assetGroupId,
    assetGroupName,
    confirm = false,
    objectId,
}) => {
    const { data: assetGroupMembers, isSuccess } = useAssetGroupMembersList(assetGroupId, objectId);
    const { mutate } = useUpdateAssetGroup();
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    if (!isSuccess) {
        return null;
    }

    const closeDialog = () => setIsDialogOpen(false);

    const openDialog = (e: React.MouseEvent<HTMLLIElement>) => {
        e.stopPropagation();
        setIsDialogOpen(true);
    };

    const isNotMember = assetGroupMembers.length === 0;
    const isCustomMember = assetGroupMembers.length === 1 && assetGroupMembers[0].custom_member;

    if (!isNotMember && !isCustomMember) {
        return null;
    }

    const operation = isNotMember ? 'add' : 'remove';

    const updateAssetGroup = () => {
        mutate({ assetGroupId, nodeId: objectId, action: operation });
    };

    // MenuItem can perform an 'add' or a 'remove' action with an optional 'Confirmation' step
    return (
        <MenuItem onClick={confirm ? openDialog : updateAssetGroup}>
            {operation === 'add' ? `Add to ${assetGroupName}` : `Remove from ${assetGroupName}`}
            {confirm && isDialogOpen && (
                <ConfirmNodeChangesDialog
                    assetGroupName={assetGroupName}
                    onApply={updateAssetGroup}
                    onCancel={closeDialog}
                    operation={operation}
                />
            )}
        </MenuItem>
    );
};

export const AssetGroupMenuItems: FC = () => {
    const { selectedItemQuery } = useExploreSelectedItem();
    const tierZeroId = useAppSelector(selectTierZeroAssetGroupId);
    const ownedId = useAppSelector(selectOwnedAssetGroupId);

    const node = isNode(selectedItemQuery.data) ? selectedItemQuery.data : undefined;

    if (node === undefined) {
        return null;
    }

    return [
        <AssetGroupMenuItem assetGroupId={tierZeroId} assetGroupName='High Value' confirm objectId={node.objectId} />,
        <AssetGroupMenuItem assetGroupId={ownedId} assetGroupName='Owned' objectId={node.objectId} />,
    ];
};
