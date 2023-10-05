import { Box, Paper } from "@mui/material"
import { AssetGroupMember } from "js-client-library";
import { FC, useState } from "react"
import AssetGroupAutocomplete, { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from "../AssetGroupAutocomplete";
import { SubHeader } from "../../views/Explore";

const AssetGroupEdit: FC<{
    assetGroupId: number,
    domainId: string,
    members: AssetGroupMember[],
}> = ({ assetGroupId, domainId, members }) => {

    const [changelog, setChangelog] = useState<AssetGroupChangelog>([]);

    const handleUpdateAssetGroupChangelog = (_event: any, changelogEntry: AssetGroupChangelogEntry) => {
        if (changelogEntry.action === ChangelogAction.ADD || changelogEntry.action === ChangelogAction.REMOVE) {
            setChangelog([...changelog, changelogEntry]);
        }
        if (changelogEntry.action === ChangelogAction.UNDO) {
            setChangelog(changelog.filter(item => item))
        }
    }

    return (
        <Box component={Paper} elevation={0} padding={1}>
            <SubHeader label={'Total Members'} count={members.length} />
            <AssetGroupAutocomplete
                assetGroupMembers={members}
                changelog={changelog}
                onChange={handleUpdateAssetGroupChangelog}
            />
        </Box>
    )
}

export default AssetGroupEdit;