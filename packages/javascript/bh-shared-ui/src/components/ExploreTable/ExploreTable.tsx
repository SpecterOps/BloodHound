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

import { ColumnDef, DataTable } from '@bloodhoundenterprise/doodleui';
import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { capitalize } from 'lodash';
import { useMemo, useState } from 'react';
import { makeFormattedObjectInfoFieldsMap } from '../../utils';
import NodeIcon from '../NodeIcon';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox';
import { TableControls } from './TableControls';

export const makeMap = (items) =>
    items.reduce((acc, col) => {
        return { ...acc, [col?.accessorKey || col?.id]: true };
    }, {});

type HasData = { data?: object };

interface ExploreTableProps<TData extends HasData> {
    open?: boolean;
    onClose?: () => void;
    data?: Record<string, TData>;
    visibleColumns?: Record<string, boolean>;
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
}

const ExploreTable = <TData extends HasData>({
    data,
    open,
    onClose,
    onManageColumnsChange,
    visibleColumns,
}: ExploreTableProps<TData>) => {
    const [searchInput, setSearchInput] = useState('');
    const mungedData = useMemo(
        () => (data && Object.keys(data).map((id) => ({ ...data?.[id]?.data, id }))) || [],
        [data]
    );
    const firstItem = mungedData?.[0];

    const labelsMap = makeFormattedObjectInfoFieldsMap(firstItem);

    const columns: ColumnDef<any, any>[] = useMemo(
        () =>
            firstItem
                ? // If column order exists in redux/localStorage, use that
                  Object.keys(firstItem).map((key: any) => {
                      return {
                          accessorKey: key,
                          header: labelsMap?.[key]?.label || capitalize(key),
                          cell: (info: any) => String(info.getValue()),
                          id: key,
                          size: 150,
                      } as ColumnDef<any, any>;
                  })
                : [],
        [labelsMap, firstItem]
    );

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

    const fallbackInitialColumns = makeMap(columns);

    if (!open || !data) return null;

    const finalColumns = [...initialColumns, ...columns];
    return (
        <div
            className={`border-2 overflow-hidden absolute z-10 bottom-16 left-4 right-4 max-h-1/2 h-[475px] bg-neutral-light-2`}>
            <div className='explore-table-container w-full h-full'>
                <TableControls
                    columns={columns}
                    visibleColumns={visibleColumns || fallbackInitialColumns}
                    onDownloadClick={() => console.log('download icon clicked')}
                    onExpandClick={() => console.log('expand icon clicked')}
                    onManageColumnsClick={() => console.log('manage columns button clicked')}
                    onManageColumnsChange={onManageColumnsChange}
                    onCloseClick={onClose}
                    tableName='Results'
                    resultsCount={mungedData?.length}
                    SearchInputProps={{
                        onChange: (e) => setSearchInput(e.target.value),
                        value: searchInput,
                        placeholder: 'Search',
                    }}
                />
                <DataTable
                    className='h-full *:h-[calc(100%-72px)]'
                    // TableProps={{
                    //     containerClassName: 'h-full',
                    // }}
                    TableHeaderProps={{
                        className: 'sticky top-0 z-10',
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
