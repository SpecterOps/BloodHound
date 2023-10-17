import { Box, Paper } from "@mui/material"
import { AssetGroupMember, UpdateAssetGroupSelectorRequest } from "js-client-library";
import { FC, useState } from "react"
import AssetGroupAutocomplete, { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from "../AssetGroupAutocomplete";
import { SubHeader } from "../../views/Explore";
import { useMutation } from "react-query";
import { apiClient } from "../../utils";
import AssetGroupChangelogTable from "./AssetGroupChangelogTable";

const AssetGroupEdit: FC<{
    assetGroupId: string,
    members: AssetGroupMember[],
}> = ({ assetGroupId, members }) => {
    const [changelog, setChangelog] = useState<AssetGroupChangelog>([]);
    const addRows = changelog.filter(entry => entry.action === ChangelogAction.ADD);
    const removeRows = changelog.filter(entry => entry.action === ChangelogAction.REMOVE);

    const handleUpdateAssetGroupChangelog = (_event: any, changelogEntry: AssetGroupChangelogEntry) => {
        if (changelogEntry.action === ChangelogAction.ADD || changelogEntry.action === ChangelogAction.REMOVE) {
            setChangelog([...changelog, changelogEntry]);
        }
        if (changelogEntry.action === ChangelogAction.UNDO) {
            handleRemoveEntryFromChangelog(changelogEntry)
        }
    }

    const mapChangelogToSelectors = (): UpdateAssetGroupSelectorRequest[] => {
        return changelog.map(item => {
            return {
                selector_name: item.objectid,
                sid: item.objectid,
                action: item.action === ChangelogAction.ADD ? 'add' : 'remove',
            }
        })
    }

    const mutation = useMutation({
        mutationFn: () => {
            const selectors = mapChangelogToSelectors();
            return apiClient.updateAssetGroupSelector(assetGroupId, selectors);
        },
        onSuccess: () => {
            setChangelog([]);
        }
    })

    const handleRemoveEntryFromChangelog = (entry: AssetGroupChangelogEntry) => {
        setChangelog(changelog.filter(item => item.objectid !== entry.objectid))
    }

    return (
        <Box component={Paper} elevation={0} padding={1}>
            <SubHeader label={'Total Members'} count={members.length} />
            <AssetGroupAutocomplete
                assetGroupMembers={members}
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
        </Box>
    )
}

export default AssetGroupEdit;