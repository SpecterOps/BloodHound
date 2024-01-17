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
import { Table, TableBody, TableCell, TableContainer, TableHead, TablePagination, TableRow } from '@mui/material';
import { Skeleton } from '@mui/material';

export interface Header {
    label: string;
    alignment?: 'inherit' | 'left' | 'center' | 'right' | 'justify';
}

export interface DataTableProps {
    headers: Header[];
    data?: any[][];
    isLoading?: boolean;
    showPaginationControls?: boolean;
    paginationProps?: {
        page: number;
        rowsPerPage: number;
        count: number;
        onPageChange: (event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null, page: number) => void;
        onRowsPerPageChange: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement>;
    };
}

const DataTable: React.FC<DataTableProps> = ({
    headers,
    data,
    isLoading = false,
    showPaginationControls = false,
    paginationProps = {
        page: 1,
        rowsPerPage: 10,
        count: 10,
        onPageChange: () => {},
        onRowsPerPageChange: () => {},
    },
}) => {
    return (
        <>
            <TableContainer>
                <Table>
                    <TableHead>
                        <TableRow>
                            {headers.map((header, index) => (
                                <TableCell key={index} align={header.alignment}>
                                    {header.label}
                                </TableCell>
                            ))}
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                {headers.map((cell, cellIndex) => (
                                    <TableCell key={cellIndex}>
                                        <Skeleton />
                                    </TableCell>
                                ))}
                            </TableRow>
                        ) : (
                            data?.map((row, rowIndex) => (
                                <TableRow key={rowIndex}>
                                    {row.map((cell, cellIndex) => (
                                        <TableCell key={cellIndex} align={headers[cellIndex]?.alignment}>
                                            {cell}
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </TableContainer>
            {showPaginationControls && (
                <TablePagination rowsPerPageOptions={[5, 10, 25, 100]} component='div' {...paginationProps} />
            )}
        </>
    );
};

export default DataTable;
