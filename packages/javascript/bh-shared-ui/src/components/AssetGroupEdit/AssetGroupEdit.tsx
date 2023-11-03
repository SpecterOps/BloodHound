import { Box, Paper } from '@mui/material';
import { AssetGroup, AssetGroupMemberParams, UpdateAssetGroupSelectorRequest } from 'js-client-library';
import { FC, useState } from 'react';
import { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from './types';
import AssetGroupAutocomplete from './AssetGroupAutocomplete';
import { SubHeader } from '../../views/Explore';
import { useMutation, useQuery } from 'react-query';
import { apiClient } from '../../utils';
import AssetGroupChangelogTable from './AssetGroupChangelogTable';
import {
    ActiveDirectoryNodeKind,
    ActiveDirectoryNodeKindToDisplay,
    AzureNodeKind,
    AzureNodeKindToDisplay,
} from '../../graphSchema';
import { useNotifications } from '../../providers';

const AssetGroupEdit: FC<{
    assetGroup: AssetGroup;
    filter: AssetGroupMemberParams;
}> = ({ assetGroup, filter }) => {
    const [changelog, setChangelog] = useState<AssetGroupChangelog>([]);
    const addRows = changelog.filter((entry) => entry.action === ChangelogAction.ADD);
    const removeRows = changelog.filter((entry) => entry.action === ChangelogAction.REMOVE);
    const { addNotification } = useNotifications();

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

    const mutation = useMutation({
        mutationFn: () => {
            const selectors = mapChangelogToSelectors();
            return apiClient.updateAssetGroupSelector(assetGroup.id.toString(), selectors);
        },
        onSuccess: () => {
            setChangelog([]);
            addNotification(
                'Update successful. Please check back later to view updated Asset Group.',
                'AssetGroupUpdateSuccess'
            );
        },
        onError: (error) => {
            console.error(error);
            setChangelog([]);
            addNotification('Unknown error, group was not updated', 'AssetGroupUpdateError');
        },
    });

    const handleRemoveEntryFromChangelog = (entry: AssetGroupChangelogEntry) => {
        setChangelog(changelog.filter((item) => item.objectid !== entry.objectid));
    };

    return (
        <Box component={Paper} elevation={0} padding={1}>
            <FilteredMemberCountDisplay assetGroupId={assetGroup.id} label='Total Count' filter={filter} />
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
            {Object.values(ActiveDirectoryNodeKind).map((kind) => {
                const filterByKind = { ...filter, primary_kind: `eq:${kind}` };
                const label = ActiveDirectoryNodeKindToDisplay(kind) || '';
                return (
                    <FilteredMemberCountDisplay
                        key={label}
                        assetGroupId={assetGroup.id}
                        label={label}
                        filter={filterByKind}
                    />
                );
            })}
            {Object.values(AzureNodeKind).map((kind) => {
                const filterByKind = { ...filter, primary_kind: `eq:${kind}` };
                const label = AzureNodeKindToDisplay(kind) || '';
                return (
                    <FilteredMemberCountDisplay
                        key={label}
                        assetGroupId={assetGroup.id}
                        label={label}
                        filter={filterByKind}
                    />
                );
            })}
        </Box>
    );
};

const FilteredMemberCountDisplay: FC<{
    assetGroupId: number;
    label: string;
    filter: AssetGroupMemberParams;
}> = ({ assetGroupId, label, filter }) => {
    const {
        data: count,
        isError,
        isLoading,
    } = useQuery(['countAssetGroupMembers', assetGroupId, filter], ({ signal }) =>
        apiClient.listAssetGroupMembers(assetGroupId.toString(), filter, { signal }).then((res) => res.data.count)
    );

    const hasValidCount = !isLoading && !isError && count && count > 0;

    if (hasValidCount) {
        return <SubHeader label={label} count={count} />;
    } else {
        return null;
    }
};

export default AssetGroupEdit;
