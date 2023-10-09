import { Box, Button, Grid, IconButton, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from "@mui/material"
import { AssetGroupMember } from "js-client-library";
import { FC, useState } from "react"
import AssetGroupAutocomplete, { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from "../AssetGroupAutocomplete";
import { SubHeader } from "../../views/Explore";
import NodeIcon from "../NodeIcon";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTimes } from "@fortawesome/free-solid-svg-icons";
import { useMutation } from "react-query";
import { apiClient } from "../..";

const AssetGroupEdit: FC<{
    assetGroupId: string,
    members: AssetGroupMember[],
}> = ({ assetGroupId, members }) => {

    console.log(members);

    const [changelog, setChangelog] = useState<AssetGroupChangelog>([]);
    const addRows = changelog.filter(entry => entry.action === ChangelogAction.ADD);
    const removeRows = changelog.filter(entry => entry.action === ChangelogAction.REMOVE);

    const handleUpdateAssetGroupChangelog = (_event: any, changelogEntry: AssetGroupChangelogEntry) => {
        if (changelogEntry.action === ChangelogAction.ADD || changelogEntry.action === ChangelogAction.REMOVE) {
            setChangelog([...changelog, changelogEntry]);
        }
        if (changelogEntry.action === ChangelogAction.UNDO) {
            removeEntryFromChangelog(changelogEntry)
        }
    }

    const mapChangelogToSelectors = (): { selector_name: string, sid: string, action: 'add' | 'remove' }[] => {
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

    const removeEntryFromChangelog = (entry: AssetGroupChangelogEntry) => {
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
                <>
                    <TableContainer>
                        <Table size='small'>
                            {addRows.length > 0 && (
                                <>
                                    <TableHead>
                                        <TableRow>
                                            <TableCell colSpan={2}>Add to Group</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {addRows.map((row) => (
                                            <TableRow key={row.objectid}>
                                                <TableCell padding='none'>
                                                    <IconButton
                                                        size='small'
                                                        onClick={() => removeEntryFromChangelog(row)}
                                                    >
                                                        <FontAwesomeIcon icon={faTimes} />
                                                    </IconButton>
                                                </TableCell>
                                                <TableCell
                                                    style={{
                                                        whiteSpace: 'nowrap',
                                                    }}>
                                                    <NodeIcon nodeType={row.type} />
                                                    {row.name}
                                                    <br />
                                                    <small>{row.objectid}</small>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </>
                            )}
                            {removeRows.length > 0 && (
                                <>
                                    <TableHead>
                                        <TableRow>
                                            <TableCell colSpan={2}>Remove from Tier Zero</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {removeRows.map((row) => (
                                            <TableRow key={row.objectid}>
                                                <TableCell padding='none'>
                                                    <IconButton
                                                        size='small'
                                                        onClick={() => removeEntryFromChangelog(row)}
                                                    >
                                                        <FontAwesomeIcon icon={faTimes} />
                                                    </IconButton>
                                                </TableCell>
                                                <TableCell
                                                    style={{
                                                        whiteSpace: 'nowrap',
                                                    }}>
                                                    <NodeIcon nodeType={row.type} />
                                                    {row.name}
                                                    <br />
                                                    <small>{row.objectid}</small>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </>
                            )}
                        </Table>
                    </TableContainer>
                    <Box mt={1}>
                        <Grid container direction='row' justifyContent='flex-end' spacing={1}>
                            <Grid item>
                                <Button color='inherit' size='small' onClick={() => setChangelog([])}>
                                    Cancel
                                </Button>
                            </Grid>
                            <Grid item>
                                <Button
                                    size='small'
                                    color='primary'
                                    variant='contained'
                                    disableElevation
                                    onClick={() => mutation.mutate()}>
                                    Confirm Changes
                                </Button>
                            </Grid>
                        </Grid>
                    </Box>
                </>
            )}
        </Box>
    )
}

export default AssetGroupEdit;