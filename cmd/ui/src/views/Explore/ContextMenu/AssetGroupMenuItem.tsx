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

import { MenuItem } from '@mui/material';
import { apiClient, useNotifications } from 'bh-shared-ui';
import { FC } from 'react';
import { useMutation, useQuery } from 'react-query';
import { useSelector } from 'react-redux';
import { selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { toggleTierZeroNode } from 'src/ducks/explore/actions';
import { AppState, useAppDispatch } from 'src/store';

const AssetGroupMenuItem: FC<{ assetGroupId: string; assetGroupName: string }> = ({ assetGroupId, assetGroupName }) => {
    const { addNotification } = useNotifications();
    const dispatch = useAppDispatch();

    const selectedNode = useSelector((state: AppState) => state.entityinfo.selectedNode);
    const tierZeroAssetGroupId = useSelector(selectTierZeroAssetGroupId);

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
            if (selectedNode?.graphId && assetGroupId === tierZeroAssetGroupId) {
                dispatch(toggleTierZeroNode(selectedNode.graphId));
            }

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
            .then((res) => res.data.data?.members)
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

export default AssetGroupMenuItem;
