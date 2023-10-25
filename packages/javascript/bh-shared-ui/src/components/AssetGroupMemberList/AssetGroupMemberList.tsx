import {
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Typography,
    useTheme
} from "@mui/material";
import { FC } from "react"
import NodeIcon from "../NodeIcon";
import { AssetGroup, AssetGroupMemberParams } from "js-client-library";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCheck, faTimes } from "@fortawesome/free-solid-svg-icons";
import { useQuery } from "react-query";
import { apiClient } from "../../utils";

const AssetGroupMemberList: FC<{
    assetGroup: AssetGroup | null,
    filter: AssetGroupMemberParams,
    onSelectMember: (member: any) => void,
}> = ({ assetGroup, filter, onSelectMember }) => {

    const theme = useTheme();

    const listAssetGroupMembersQuery = useQuery(
        ["listAssetGroupMembers", assetGroup, filter],
        ({ signal }) => apiClient.listAssetGroupMembers(`${assetGroup?.id}`, filter, { signal }).then(res => res.data.data.members),
    );

    const hoverStyles = {
        "&:hover": {
            backgroundColor: theme.palette.action.hover,
            cursor: "pointer"
        }
    }

    return (
        <TableContainer sx={{ maxHeight: "100%" }} component={Paper} elevation={0}>
            <Table stickyHeader sx={{ height: "100%" }}>
                <TableHead>
                    <TableRow>
                        <TableCell sx={{ bgcolor: "white" }}>Name</TableCell>
                        <TableCell sx={{ bgcolor: "white", textAlign: "center" }} align="right">Custom Member</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody sx={{ height: "100%", overflow: "auto" }}>
                    {listAssetGroupMembersQuery.data?.map(member => {
                        return (
                            <TableRow
                                onClick={() => onSelectMember(member)}
                                sx={{...hoverStyles }}
                                key={member.object_id}
                            >
                                <TableCell><NodeIcon nodeType={member.primary_kind} />
                                    <Typography marginLeft={1} display={"inline-block"}>{member.name}</Typography>
                                </TableCell>
                                <TableCell align="right" sx={{
                                    padding: "0",
                                    display: "flex",
                                    justifyContent: "center",
                                    alignItems: "center",
                                    height: "100%"
                                }}>
                                    {member.custom_member ?
                                        <FontAwesomeIcon icon={faCheck} color="green" size="lg" /> :
                                        <FontAwesomeIcon icon={faTimes} color="red" size="lg" />
                                    }
                                </TableCell>
                            </TableRow>
                        )
                    })}
                </TableBody>
            </Table>
        </TableContainer>
    );
}

export default AssetGroupMemberList;