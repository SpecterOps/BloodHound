import {
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    useTheme
} from "@mui/material";
import { FC } from "react"
import NodeIcon from "../NodeIcon";

const AssetGroupMemberList: FC<{
    assetGroupMembers: any[],
    onSelectMember: (member: any) => void,
}> = ({ assetGroupMembers, onSelectMember }) => {

    const theme = useTheme();

    const hoverStyles = {
        "&:hover": {
            backgroundColor: theme.palette.action.hover,
            cursor: "pointer"
        }
    }

    return (
        <TableContainer sx={{ height: "100%" }} component={Paper} elevation={0}>
            <Table stickyHeader sx={{ height: "100%" }}>
                <TableHead>
                    <TableRow>
                        <TableCell sx={{ bgcolor: "white" }}>Name</TableCell>
                        <TableCell sx={{ bgcolor: "white" }} align="right">Custom Member</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody sx={{ height: "100%", overflow: "auto" }}>
                    {assetGroupMembers?.map(member => {
                        return (
                            <TableRow
                                onClick={() => onSelectMember(member)}
                                sx={{...hoverStyles }}
                                key={member.object_id}
                            >
                                <TableCell><NodeIcon nodeType={member.primary_kind} />{member.name}</TableCell>
                                <TableCell align="right">X</TableCell>
                            </TableRow>
                        )
                    })}
                </TableBody>
            </Table>
        </TableContainer>
    );
}

export default AssetGroupMemberList;