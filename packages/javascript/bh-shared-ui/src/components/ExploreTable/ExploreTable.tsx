// Copyright 2025 Specter Ops, Inc.
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

import React, { useMemo, useState } from 'react';

// needed for table body level scope DnD setup

// needed for table body level scope DnD setup

// needed for row & cell level scope DnD setup
// needed for table body level scope DnD setup
import { ColumnDef, DataTable } from '@bloodhoundenterprise/doodleui';
import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { capitalize } from 'lodash';
import { makeFormattedObjectInfoFieldsMap } from '../../utils';
import NodeIcon from '../NodeIcon';

interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    onRowClick?: (data: any) => void;
    selectedRow: string;
    items?: any;
}

const ExploreTable: React.FC<ExploreTableProps> = ({ items, open, onClose, onRowClick = () => {}, selectedRow }) => {
    const [searchInput, setSearchInput] = useState('');
    const mungedData = useMemo(
        () =>
            items &&
            Object.keys(items)
                .map((id) => ({ ...items[id]?.data, id }))
                .slice(0, 40),
        [items]
    );

    const firstItem = mungedData?.[0];

    const labelsMap = makeFormattedObjectInfoFieldsMap(firstItem);

    const initialColumns: ColumnDef<any, any>[] = [
        {
            accessorKey: '',
            id: 'action-menu',
            cell: () => (
                <button className='pl-4'>
                    <FontAwesomeIcon icon={faEllipsis} className='rotate-90 dark:text-neutral-light-1' />
                </button>
            ),
        },
        {
            accessorKey: 'nonTierZeroPrincipal',
            header: () => {
                return <span className='dark:text-neutral-light-1'>Non Tier Zero Principal</span>;
            },
            cell: ({ row }) => {
                return (
                    <div className='flex justify-center items-center relative'>
                        <NodeIcon nodeType={row?.original?.nodetype || 'N/A'} />
                    </div>
                );
            },
        },
    ];

    const columns: ColumnDef<any, any>[] = useMemo(
        () =>
            firstItem &&
            // If column order exists in redux/localStorage, use that
            Object.keys(firstItem)
                .slice(0, 10)
                .map((key: any) => {
                    return {
                        accessorKey: key,
                        header: labelsMap?.[key]?.label || capitalize(key),
                        cell: (info: any) => String(info.getValue()),
                        id: key,
                        size: 150,
                    } as ColumnDef<any, any>;
                }),
        [labelsMap, firstItem]
    );

    if (!open || !items) return null;

    const finalColumns = [...initialColumns, ...columns];
    return (
        <div
            className={`border-2 overflow-hidden absolute bottom-16 left-4 right-4 ${selectedRow ? 'w-[calc(100%-450px)]' : 'w-90'} max-h-1/2 h-[475px] bg-neutral-light-2`}>
            <div className='explore-table-container w-full h-full'>
                <DataTable
                    className='h-full'
                    onRowClick={onRowClick}
                    selectedRow={selectedRow}
                    TableProps={{
                        containerClassName: 'h-full bg-cyan',
                    }}
                    TableHeaderProps={{
                        // TODO: icons were visible over header on scroll, find solution without z-index?
                        className: 'sticky top-0 z-10',
                    }}
                    TableControlsProps={{
                        onDownloadClick: () => alert('download icon clicked'),
                        onExpandClick: () => alert('expand icon clicked'),
                        onManageColumnsClick: () => alert('manage columns button clicked'),
                        onCloseClick: onClose,
                        tableName: 'Results',
                        resultsCount: mungedData.length,
                        SearchInputProps: {
                            onChange: (e) => setSearchInput(e.target.value),
                            value: searchInput,
                            placeholder: 'Search',
                        },
                    }}
                    tableOptions={{
                        getRowId: (row) => row?.id,
                    }}
                    data={mungedData}
                    columns={finalColumns}
                />
            </div>
        </div>
    );
};

export default ExploreTable;
