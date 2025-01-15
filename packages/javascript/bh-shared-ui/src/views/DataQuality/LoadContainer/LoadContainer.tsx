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
    icon: JSX.Element;
    loading: boolean;
    display: string;
    value?: number;
    type?: 'percent' | 'number';
}

const StyledTableRow = withStyles({
    root: {
        '&:last-child': {
            borderBottom: 'none',
        },
    },
})(TableRow);

const LoadContainer: React.FC<LoadContainerProps> = ({ icon, loading, display, value = 0, type = 'number' }) => {
    return (
        <StyledTableRow>
            <TableCell>
                <Box display='inline-block' width='32px' textAlign='center'>
                    {icon}
                </Box>
                {display}
            </TableCell>
            <TableCell align='right'>
                {loading ? (
                    <Skeleton variant='text' />
                ) : type === 'percent' ? (
                    `${Math.floor(value * 100)}%`
                ) : (
                    value.toLocaleString()
                )}
            </TableCell>
        </StyledTableRow>
    );
};

export default LoadContainer;
