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

import { DataTable } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagCertificationRecord } from 'js-client-library';
import { FC, useCallback, useEffect, useRef } from 'react';
import { measureElement } from '../utils';
import { useAssetGroupTagsCertificationsQuery, useCertificationColumns } from './hooks';

type DataTableProps = React.ComponentProps<typeof DataTable>;

const tableProps: DataTableProps['TableProps'] = {
    className: 'table-fixed',
    disableDefaultOverflowAuto: true,
};

const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
    className: 'sticky top-0 z-10 shadow-sm text-base',
};

const tableHeadProps: DataTableProps['TableHeadProps'] = {
    className: 'text-left',
};

const tableCellProps: DataTableProps['TableCellProps'] = {
    className: 'text-left truncate py-0',
};

type CertificationTableProps = {
    query: ReturnType<typeof useAssetGroupTagsCertificationsQuery>;
    onRowSelect: (row: AssetGroupTagCertificationRecord) => void;
    onRowCheck: (row: AssetGroupTagCertificationRecord) => void;
    toggleAllRowsSelected: () => void;
    selectedRows: Record<string, boolean>;
    items: AssetGroupTagCertificationRecord[];
    count: number;
};

const CertificationTable: FC<CertificationTableProps> = ({
    toggleAllRowsSelected,
    query,
    onRowSelect,
    onRowCheck,
    selectedRows = {},
    items,
    count,
}) => {
    const scrollRef = useRef<HTMLDivElement>(null);
    const { isFetching, fetchNextPage } = query;

    const totalFetched = items.length;

    const fetchMoreOnBottomReached = useCallback(
        (containerRefElement?: HTMLDivElement | null) => {
            if (!containerRefElement) return;

            const { scrollHeight, scrollTop, clientHeight } = containerRefElement;

            if (scrollHeight - scrollTop - clientHeight < 100 && !isFetching && totalFetched < count) fetchNextPage();
        },
        [fetchNextPage, isFetching, totalFetched, count]
    );

    useEffect(() => {
        fetchMoreOnBottomReached(scrollRef.current);
    }, [fetchMoreOnBottomReached]);

    const ids = Object.keys(selectedRows);
    const allRowsAreSelected = ids.length > 0 && ids.length === items.length;

    const columns = useCertificationColumns({
        onRowSelect: onRowCheck,
        toggleAllRowsSelected,
        selectedRows,
        allRowsAreSelected,
    });

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        count: totalFetched ?? 0,
        getScrollElement: () => scrollRef.current,
        measureElement,
        overscan: 2,
        estimateSize: () => 50,
    };

    return (
        <div
            onScroll={(e) => fetchMoreOnBottomReached(e.currentTarget)}
            ref={scrollRef}
            className='overflow-y-auto h-[65dvh]'>
            <DataTable
                data={items}
                onRowClick={onRowSelect}
                TableHeaderProps={tableHeaderProps}
                TableHeadProps={tableHeadProps}
                TableProps={tableProps}
                TableCellProps={tableCellProps}
                columns={columns as any}
                virtualizationOptions={virtualizationOptions}
            />
        </div>
    );
};

export default CertificationTable;
