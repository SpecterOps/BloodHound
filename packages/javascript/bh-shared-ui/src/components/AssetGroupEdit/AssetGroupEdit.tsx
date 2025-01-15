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

import { Box, Paper, useTheme } from '@mui/material';
import {
    AssetGroup,
    AssetGroupMemberCounts,
    AssetGroupMemberParams,
    UpdateAssetGroupSelectorRequest,
} from 'js-client-library';
import { FC, useEffect, useState } from 'react';
import { useMutation, useQueryClient } from 'react-query';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import { SubHeader } from '../../views/Explore';
import AssetGroupAutocomplete from './AssetGroupAutocomplete';
import AssetGroupChangelogTable from './AssetGroupChangelogTable';
import { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from './types';

const AssetGroupEdit: FC<{
    assetGroup: AssetGroup;
    filter: AssetGroupMemberParams;
    memberCounts: AssetGroupMemberCounts | undefined;
    isEditable: boolean;
}> = ({ assetGroup, filter, memberCounts, isEditable }) => {
    const [changelog, setChangelog] = useState<AssetGroupChangelog>([]);
    const addRows = changelog.filter((entry) => entry.action === ChangelogAction.ADD);
    const removeRows = changelog.filter((entry) => entry.action === ChangelogAction.REMOVE);
    const { addNotification } = useNotifications();
    const theme = useTheme();
    const queryClient = useQueryClient();

    const handleUpdateAssetGroupChangelog = (_event: any, changelogEntry: AssetGroupChangelogEntry) => {
        if (changelogEntry.action === ChangelogAction.ADD || changelogEntry.action === ChangelogAction.REMOVE) {
            setChangelog([...changelog, changelogEntry]);
        }
        if (changelogEntry.action === ChangelogAction.UNDO) {
            handleRemoveEntryFromChangelog(changelogEntry);
        }
    };

    const mapChangelogToSelectors = (): UpdateAssetGroupSelectorRequest[] => {
        return changelog.map((item) => {
            return {
                selector_name: item.objectid,
                sid: item.objectid,
                action: item.action === ChangelogAction.ADD ? 'add' : 'remove',
            };
        });
    };

    // Clear out changelog when group/domain changes
    useEffect(() => setChangelog([]), [filter.environment_id, filter.environment_kind]);

    const mutation = useMutation({
        mutationFn: () => {
            const selectors = mapChangelogToSelectors();
            return apiClient.updateAssetGroupSelector(assetGroup.id, selectors);
        },
        onSuccess: () => {
            setChangelog([]);

            // refetch all page data after updating group membership
            queryClient.invalidateQueries({ queryKey: ['listAssetGroups'] });
            queryClient.invalidateQueries({ queryKey: ['listAssetGroupMembers'] });
            queryClient.invalidateQueries({ queryKey: ['countAssetGroupMembers'] });
            queryClient.resetQueries({ queryKey: ['search'] });

            addNotification('Update successful.', 'AssetGroupUpdateSuccess');
        },
        onError: (error) => {
            console.error(error);
            setChangelog([]);
            addNotification('Unknown error, group was not updated', 'AssetGroupUpdateError');
        },
    });

    const handleRemoveEntryFromChangelog = (entry: AssetGroupChangelogEntry) => {
        setChangelog((prev) => prev.filter((item) => item.objectid !== entry.objectid));
    };

    return (
        <Box component={Paper} elevation={0} padding={1} bgcolor={theme.palette.neutral.secondary}>
            <SubHeader label='Total Count' count={memberCounts?.total_count} />
            {isEditable && (
                <>
                    <AssetGroupAutocomplete
                        assetGroup={assetGroup}
                        changelog={changelog}
                        onChange={handleUpdateAssetGroupChangelog}
                    />
                    {changelog.length > 0 && (
                        <AssetGroupChangelogTable
                            addRows={addRows}
                            removeRows={removeRows}
                            onRemove={handleRemoveEntryFromChangelog}
                            onCancel={() => setChangelog([])}
                            onSubmit={() => mutation.mutate()}
                        />
                    )}
                </>
            )}
            {Object.entries(memberCounts?.counts ?? {}).map(([kind, count]) => {
                return <SubHeader key={kind} label={kind} count={count} />;
            })}
        </Box>
    );
};

export default AssetGroupEdit;
