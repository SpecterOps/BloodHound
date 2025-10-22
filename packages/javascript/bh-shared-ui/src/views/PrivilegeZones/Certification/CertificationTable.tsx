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
import { createColumnHelper, DataTable } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagCertificationRecord,
    CertificationAuto,
    CertificationManual,
    CertificationPending,
    CertificationRevoked,
    CertificationTypeMap,
    ExtendedCertificationFilters,
} from 'js-client-library';
import { DateTime } from 'luxon';
import { Dispatch, FC, SetStateAction, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { AppIcon, DropdownOption, DropdownSelector, NodeIcon } from '../../../components';
import { SearchInput } from '../../../components/SearchInput';
import { useAssetGroupTags, useAvailableEnvironments } from '../../../hooks';
import FilterDialog from './FilterDialog';
import { InfiniteData } from 'react-query';

type CertificationsPage = {
    count: number;
    limit: number;
    skip: number;
    data: {
        members: AssetGroupTagCertificationRecord[];
    };
};

type CertificationTableProps = {
    data: InfiniteData<CertificationsPage> | undefined;
    filters: ExtendedCertificationFilters;
    setFilters: Dispatch<SetStateAction<ExtendedCertificationFilters>>;
    search: string;
    setSearch: Dispatch<SetStateAction<string>>;
    onRowSelect: (row: AssetGroupTagCertificationRecord | null) => void;
    isLoading: boolean;
    isFetching: boolean;
    isSuccess: boolean;
    fetchNextPage: () => Promise<unknown>;
    filterRows: (dropdownSelection: DropdownOption) => void;
    applyAdvancedFilters?: (advancedFilters: Partial<ExtendedCertificationFilters>) => void;
    selectedRows: number[];
    setSelectedRows: Dispatch<SetStateAction<number[]>>;
};

const CertificationTable: FC<CertificationTableProps> = ({
    data,
    filters,
    setFilters,
    search,
    setSearch,
    onRowSelect,
    isFetching,
    isSuccess,
    fetchNextPage,
    filterRows,
    applyAdvancedFilters,
    selectedRows,
    setSelectedRows,
}) => {
    const scrollRef = useRef<HTMLDivElement>(null);

    const [dropdownSelection, setDropdownSelection] = useState('Pending');

    const { data: availableEnvironments = [] } = useAvailableEnvironments();
    const { data: assetGroupTags = [] } = useAssetGroupTags();

    const domainMap = useMemo(() => {
        const map = new Map<string, string>();
        for (const domain of availableEnvironments) {
            map.set(domain.id, domain.name);
        }
        return map;
    }, [availableEnvironments]);

    const tagMap = useMemo(() => {
        const map = new Map<number, string>();
        for (const tag of assetGroupTags) {
            map.set(tag.id, tag.name);
        }
        return map;
    }, [assetGroupTags]);

    const certificationsData = data ?? { pages: [{ count: 0, data: { members: [] } }] };
    const count = certificationsData.pages[0].count;
    const certificationsItemsRaw =
        data?.pages.flatMap(
            (page: { data?: { members: AssetGroupTagCertificationRecord[] } }) => page.data?.members ?? []
        ) ?? [];
    const totalFetched = certificationsItemsRaw.length;

    const certificationsItems = isSuccess
        ? certificationsItemsRaw
              .map((item: AssetGroupTagCertificationRecord) => {
                  return {
                      ...item,
                      date: DateTime.fromISO(item.created_at).toFormat('MM-dd-yyyy'),
                      domainName: domainMap.get(item.environment_id) ?? 'Unknown',
                      zoneName: tagMap.get(item.asset_group_tag_id) ?? 'Unknown',
                  };
              })
              .filter((item: AssetGroupTagCertificationRecord) => {
                  if (filters.objectType && item.primary_kind !== filters.objectType) return false;
                  if (filters.approvedBy && item.certified_by !== filters.approvedBy) return false;
                  if (filters.startDate && DateTime.fromISO(item.created_at) < DateTime.fromISO(filters.startDate))
                      return false;
                  if (filters.endDate && DateTime.fromISO(item.created_at) > DateTime.fromISO(filters.endDate))
                      return false;
                  if (search) {
                      const query = search.toLowerCase();
                      return (
                          item.name?.toLowerCase().includes(query) || item.primary_kind?.toLowerCase().includes(query)
                      );
                  }
                  return true;
              })
        : [];

    const fetchMoreOnBottomReached = useCallback(
        (containerRefElement?: HTMLDivElement | null) => {
            if (containerRefElement) {
                const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
                //once the user has scrolled within 500px of the bottom of the table, fetch more data if we can
                if (scrollHeight - scrollTop - clientHeight < 500 && !isFetching && totalFetched < count) {
                    fetchNextPage();
                }
            }
        },
        [fetchNextPage, isFetching, totalFetched, count]
    );

    useEffect(() => {
        fetchMoreOnBottomReached(scrollRef.current);
    }, [fetchMoreOnBottomReached]);

    const allSelected = certificationsItems?.length > 0 && selectedRows.length === certificationsItems?.length;
    const someSelected = selectedRows.length > 0 && !allSelected;

    const toggleAll = (checked: boolean) => {
        if (checked) {
            setSelectedRows(certificationsItems.map((row: any) => row.id));
        } else {
            setSelectedRows([]);
        }
    };

    const toggleRow = (checked: boolean, selectedId: number) => {
        if (checked) {
            setSelectedRows([...selectedRows, selectedId]);
        } else {
            setSelectedRows(selectedRows.filter((rowId: number) => rowId !== selectedId));
        }
    };

    const columnHelper = createColumnHelper<any>();

    const columns = [
        columnHelper.display({
            id: 'bulk-certify',
            header: () => (
                <div className='pl-8'>
                    <input
                        data-testid='certification-table-select-all'
                        type='checkbox'
                        checked={allSelected}
                        ref={(el) => {
                            if (el) el.indeterminate = someSelected;
                        }}
                        onChange={(event) => toggleAll(event.target.checked)}
                    />
                </div>
            ),
            cell: (info) => (
                <div className='pl-8'>
                    <input
                        type='checkbox'
                        data-testid={`certification-table-row-${info.row.original.id}`}
                        checked={selectedRows.includes(info.row.original.id)}
                        onChange={(event) => toggleRow(event.target.checked, info.row.original.id)}
                    />
                </div>
            ),
            size: 65,
        }),
        columnHelper.accessor('primary_kind', {
            header: 'Type',
            cell: (info) => (
                <div className='text-primary dark:text-secondary-variant-2'>
                    {<NodeIcon nodeType={info.getValue()} />}
                </div>
            ),
            size: 100,
        }),
        columnHelper.accessor('name', {
            header: 'Member Name',
            cell: (info) => <div className='min-w-0 w-[150px] truncate'>{info.getValue()}</div>,
            size: 150,
        }),
        columnHelper.accessor('domainName', {
            header: 'Domain',
            cell: (info) => <div className='min-w-0 w-[150px] truncate'>{info.getValue()}</div>,
        }),
        columnHelper.accessor('zoneName', {
            header: 'Zone',
            cell: (info) => <div className='min-w-0 w-[150px] truncate'>{info.getValue()}</div>,
        }),
        columnHelper.accessor('date', {
            header: 'First Seen',
            cell: (info) => <div className='text-left'>{info.getValue()}</div>,
        }),
    ];

    type DataTableProps = React.ComponentProps<typeof DataTable>;

    const tableProps: DataTableProps['TableProps'] = {
        className: 'table-fixed',
        disableDefaultOverflowAuto: true,
    };

    const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
        className: 'sticky top-0 z-10 shadow-sm text-base',
    };

    const tableHeadProps: DataTableProps['TableHeadProps'] = {
        className: 'pl-8 text-left',
    };

    const tableCellProps: DataTableProps['TableCellProps'] = {
        className: 'pl-8 text-left truncate group relative',
    };

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        estimateSize: () => 79,
    };

    const certOptions: DropdownOption[] = [
        CertificationPending,
        CertificationManual,
        CertificationAuto,
        CertificationRevoked,
    ].map((certType) => {
        return { key: certType, value: CertificationTypeMap[certType] };
    });

    const certificationCountTextMap: Record<string, string> = {
        Pending: 'pending',
        Certified: 'certified',
        'Automatic Certification': 'automatically certified',
        Rejected: 'rejected',
    };

    return (
        <div className='bg-neutral-light-2 dark:bg-neutral-dark-2'>
            <div className='flex items-center px-8 py-4'>
                <h1 className='text-xl font-bold pr-4'>Certifications</h1>
                {count && <p>{`${count} ${certificationCountTextMap[dropdownSelection] ?? 'pending'}`}</p>}
            </div>
            <div className='pl-8 flex justify-between'>
                <DropdownSelector
                    variant='transparent'
                    options={certOptions}
                    selectedText={
                        <span className='flex items-center gap-3'>
                            <AppIcon.CertStatus size={24} /> <p>{`${dropdownSelection}`}</p>
                        </span>
                    }
                    onChange={(selectedCertificationType: DropdownOption) => {
                        setDropdownSelection(selectedCertificationType.value);
                        filterRows(selectedCertificationType);
                    }}
                />
                <div className='flex items-center'>
                    <SearchInput value={search} onInputChange={setSearch} />
                    {applyAdvancedFilters && (
                        <FilterDialog
                            filters={filters}
                            setFilters={setFilters}
                            onApplyFilters={applyAdvancedFilters}
                            data={certificationsItems}
                        />
                    )}
                </div>
            </div>
            <div
                onScroll={(e) => fetchMoreOnBottomReached(e.currentTarget)}
                ref={scrollRef}
                className='overflow-y-auto h-[calc(90vh_-_255px)]'>
                <DataTable
                    data={certificationsItems ?? []}
                    onRowClick={(row) => onRowSelect(row)}
                    TableHeaderProps={tableHeaderProps}
                    TableHeadProps={tableHeadProps}
                    TableProps={tableProps}
                    TableCellProps={tableCellProps}
                    columns={columns}
                    virtualizationOptions={virtualizationOptions}
                />
            </div>
        </div>
    );
};

export default CertificationTable;
