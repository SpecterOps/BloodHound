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

import { faCheck, faTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
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
} from '@mui/material';
import { AssetGroup, AssetGroupMember, AssetGroupMemberParams } from 'js-client-library';
import { FC, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { apiClient, cn } from '../../utils';
import NodeIcon from '../NodeIcon';

const AssetGroupMemberList: FC<{
    assetGroup: AssetGroup | null;
    filter: AssetGroupMemberParams;
    onSelectMember: (member: any) => void;
    canFilterToEmpty: boolean;
}> = ({ assetGroup, filter, onSelectMember, canFilterToEmpty }) => {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(25);
    const [count, setCount] = useState(0);

    const { data, isLoading, isPreviousData, isSuccess } = useQuery(
        ['listAssetGroupMembers', assetGroup, filter, page, rowsPerPage],
        ({ signal }) => {
            const paginatedFilter = {
                skip: page * rowsPerPage,
                limit: rowsPerPage,
                // we could make this user selected in the future
                sort_by: 'name',
                ...filter,
            };
            if (assetGroup && assetGroup.id) {
                return apiClient.listAssetGroupMembers(assetGroup.id, paginatedFilter, { signal }).then((res) => {
                    setCount(res.data.count);
                    return res.data.data.members;
                });
            }
        },
        {
            enabled: !!assetGroup,
            keepPreviousData: true,
        }
    );

    // Prevents an error that occurs if you try to query with a "skip" value greater than the member count of the current group
    useEffect(() => setPage(0), [assetGroup, filter]);

    const getLoadingRows = (count: number) => {
        const rows = [];
        for (let i = 0; i < count; i++) {
            rows.push(
                <TableRow key={i}>
                    <TableCell>
                        <Skeleton variant='text' />
                    </TableCell>
                    <TableCell>
                        <Skeleton variant='text' />
                    </TableCell>
                </TableRow>
            );
        }
        return rows;
    };

    return (
        <TableContainer className='max-h-full bg-neutral-2' component={Paper} elevation={0}>
            <Table stickyHeader className='h-full relative'>
                <colgroup>
                    <col width='80%' />
                    <col width='20%' />
                </colgroup>
                <TableHead>
                    <TableRow>
                        <TableCell className='bg-neutral-2'>Name</TableCell>
                        <TableCell className='bg-neutral-2 text-center' align='right'>
                            Custom Member
                        </TableCell>
                    </TableRow>
                </TableHead>
                <TableBody className='h-full overflow-auto'>
                    {isLoading && getLoadingRows(10)}
                    {isSuccess &&
                        !!data?.length &&
                        data.map((member) => (
                            <AssetGroupMemberRow
                                member={member}
                                onClick={onSelectMember}
                                key={member.object_id}
                                disabled={isPreviousData}
                            />
                        ))}
                    {isSuccess && data?.length === 0 && (
                        <TableRow>
                            <TableCell className='text-center h-24' colSpan={2}>
                                {canFilterToEmpty
                                    ? 'No members match that filter'
                                    : 'No members in selected Asset Group'}
                            </TableCell>
                        </TableRow>
                    )}
                </TableBody>
                {isSuccess && !!data?.length && (
                    <TableFooter>
                        <TableRow>
                            <TablePagination
                                className='sticky bottom-0 bg-neutral-2 border-t border-neutral-light-3'
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
};

const AssetGroupMemberRow: FC<{
    member: AssetGroupMember;
    disabled: boolean;
    onClick: (member: AssetGroupMember) => void;
}> = ({ member, disabled, onClick }) => {
    const handleClick = () => {
        if (!disabled) onClick(member);
    };

    return (
        <TableRow
            onClick={handleClick}
            className={cn({ 'hover:bg-neutral-4 cursor-pointer': !disabled, 'opacity-50': disabled })}>
            <TableCell>
                <Box className='flex items-center w-full'>
                    <NodeIcon nodeType={member.primary_kind} />
                    <Typography noWrap marginLeft={1} display={'inline-block'}>
                        {member.name}
                    </Typography>
                </Box>
            </TableCell>
            <TableCell align='right' className='p-0 flex justify-center items-center h-full'>
                {member.custom_member ? (
                    <FontAwesomeIcon icon={faCheck} color='green' size='lg' />
                ) : (
                    <FontAwesomeIcon icon={faTimes} color='red' size='lg' />
                )}
            </TableCell>
        </TableRow>
    );
};

export default AssetGroupMemberList;
