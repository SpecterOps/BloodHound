import {
    Avatar,
    Box,
    Paper,
    Skeleton,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableFooter,
    TableHead,
    TablePagination,
    TableRow,
    Typography,
    useTheme
} from "@mui/material";
import { FC, useState } from "react"
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

    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(25);
    const [count, setCount] = useState(0);

    const { data, isLoading, isSuccess } = useQuery(
        ["listAssetGroupMembers", assetGroup, filter, page, rowsPerPage],
        ({ signal }) => {
            const paginatedFilter = {
                skip: page * rowsPerPage,
                limit: rowsPerPage,
                // we could make this user selected in the future
                sort_by: "name",
                ...filter,
            }
            return apiClient.listAssetGroupMembers(`${assetGroup?.id}`, paginatedFilter, { signal })
                .then(res => {
                    setCount(res.data.count);
                    return res.data.data.members;
                })
        },
        {
            enabled: !!assetGroup,
            keepPreviousData: true
        }
    );

    const hoverStyles = {
        "&:hover": {
            backgroundColor: theme.palette.action.hover,
            cursor: "pointer"
        }
    }

    const getLoadingRows = (count: number) => {
        const rows = [];
        for (let i = 0; i < count; i++) {
            rows.push(
                <TableRow key={i}>
                    <TableCell><Skeleton variant="text" /></TableCell>
                    <TableCell><Skeleton variant="text" /></TableCell>
                </TableRow> 
            )
        }
        return rows;
    }

    return (
        <TableContainer sx={{ maxHeight: "100%" }} component={Paper} elevation={0}>
            <Table stickyHeader sx={{ height: "100%", position: "relative" }}>
                <colgroup>
                    <col width="80%" />
                    <col width="20%" />
                </colgroup>
                <TableHead>
                    <TableRow>
                        <TableCell sx={{ bgcolor: "white" }}>Name</TableCell>
                        <TableCell sx={{ bgcolor: "white", textAlign: "center" }} align="right">Custom Member</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody sx={{ height: "100%", overflow: "auto" }}>
                    {isLoading && getLoadingRows(5)}
                    {isSuccess && data?.map(member => {
                        return (
                            <TableRow
                                onClick={() => onSelectMember(member)}
                                sx={{...hoverStyles }}
                                key={member.object_id}
                            >
                                <TableCell>
                                    <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                                        <NodeIcon nodeType={member.primary_kind} />
                                        <Typography noWrap marginLeft={1} display={"inline-block"}>{member.name}</Typography>
                                    </Box>    
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
                {count > 0 && (
                    <TableFooter>
                        <TableRow>
                            <TablePagination
                                sx={{ position: "sticky", bottom: 0, bgcolor: "white", borderTop: "1px solid #E0E0E0" }}
                                colSpan={2}
                                rowsPerPageOptions={[10, 25, 100, 250]}
                                page={page}
                                rowsPerPage={rowsPerPage}
                                count={count}
                                onPageChange={(_event, page) => setPage(page)}
                                onRowsPerPageChange={(event) => setRowsPerPage(parseInt(event.target.value))}
                            />
                        </TableRow>
                    </TableFooter>
                )}
            </Table>
        </TableContainer>
    );
}

export default AssetGroupMemberList;