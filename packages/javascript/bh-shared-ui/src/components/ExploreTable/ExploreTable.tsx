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

import React, { useEffect, useMemo, useState } from 'react';

// needed for table body level scope DnD setup

// needed for table body level scope DnD setup

// needed for row & cell level scope DnD setup
// needed for table body level scope DnD setup
import { ColumnDef, DataTable } from '@bloodhoundenterprise/doodleui';
import { capitalize } from 'lodash';
import { makeFormattedObjectInfoFieldsMap } from '../../utils';

interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    items?: any;
}

const ExploreTable: React.FC<ExploreTableProps> = ({ items, open, onClose }) => {
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

    useEffect(
        () => () => {
            if (typeof onClose === 'function') onClose();
        },
        [onClose]
    );

    const columns: ColumnDef<any, any>[] = useMemo(
        () =>
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

    if (!open) return null;

    return (
        <div className='border-2 absolute bottom-24 left-4 right-4 w-4/5 h-1/2'>
            <div className='explore-table-container w-full h-full'>
                <DataTable
                    className='h-full'
                    TableProps={{
                        containerClassName: 'h-full',
                    }}
                    TableControlsProps={{
                        onDownloadClick: () => alert('download icon clicked'),
                        onExpandClick: () => alert('expand icon clicked'),
                        onManageColumnsClick: () => alert('manage columns button clicked'),
                        onCloseClick: () => alert('close icon clicked'),
                        tableName: 'Results',
                        resultsCount: 230,
                        SearchInputProps: {
                            onChange: (e) => setSearchInput(e.target.value),
                            value: searchInput,
                            placeholder: 'Search',
                        },
                    }}
                    data={mungedData}
                    columns={columns}
                />
            </div>
        </div>
    );
};

export default ExploreTable;
