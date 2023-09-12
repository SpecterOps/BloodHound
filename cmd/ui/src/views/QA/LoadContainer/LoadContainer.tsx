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

import React from 'react';
import { TableRow, TableCell, Box } from '@mui/material';
import withStyles from '@mui/styles/withStyles';
import { Skeleton } from '@mui/material';

interface LoadContainerProps {
    icon: any;
    value?: number;
    display: string;
    type?: 'percent' | 'number';
    loading?: boolean;
}

const StyledTableRow = withStyles({
    root: {
        '&:last-child': {
            borderBottom: 'none',
        },
    },
})(TableRow);

const LoadContainer: React.FC<LoadContainerProps> = ({ icon, value = 0, display, type = 'number', loading }) => {
    if (loading)
        return (
            <StyledTableRow>
                <TableCell>
                    <Box display='inline-block' width='32px' textAlign='center'>
                        {icon}
                    </Box>
                    {display}
                </TableCell>
                <TableCell align='right'>
                    <Skeleton variant='text' />
                </TableCell>
            </StyledTableRow>
        );

    return (
        <StyledTableRow>
            <TableCell>
                <Box display='inline-block' width='32px' textAlign='center'>
                    {icon}
                </Box>
                {display}
            </TableCell>
            <TableCell align='right'>
                {type === 'percent' ? `${Math.floor(value * 100)}%` : value.toLocaleString()}
            </TableCell>
        </StyledTableRow>
    );
};

export default LoadContainer;
