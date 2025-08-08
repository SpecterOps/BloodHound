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
import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import isEmpty from 'lodash/isEmpty';
import { useCallback, useMemo, useState } from 'react';
import {
    compareForExploreTableSort,
    getExploreTableData,
    isSmallColumn,
    type ExploreTableProps,
    type MungedTableRowWithId,
} from './explore-table-utils';
import ExploreTableDataCell from './ExploreTableDataCell';
import ExploreTableHeaderCell from './ExploreTableHeaderCell';

const columnHelper = createColumnHelper<MungedTableRowWithId>();

type DataTableProps = React.ComponentProps<typeof DataTable>;

const filterKeys: (keyof MungedTableRowWithId)[] = ['label', 'objectid'];

type UseExploreTableRowsAndColumnsProps = Pick<ExploreTableProps, 'onKebabMenuClick' | 'selectedColumns'> & {
    searchInput: string;
    exploreTableData: ReturnType<typeof getExploreTableData>;
};

const useExploreTableRowsAndColumns = ({
    onKebabMenuClick,
    searchInput,
    selectedColumns,
    exploreTableData,
}: UseExploreTableRowsAndColumnsProps) => {
    const [sortBy, setSortBy] = useState<keyof MungedTableRowWithId>();
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>();

    const rows: MungedTableRowWithId[] = useMemo(
        () =>
            exploreTableData?.nodes
                ? Object.entries(exploreTableData?.nodes).map(([key, node]) => {
                      // To avoid extra enumerations for spread operators, the known properties are manually set
                      const flattenedNode = {
                          id: key,
                          label: node.label,
                          kind: node.kind,
                          objectId: node.objectId,
                          lastSeen: node.lastSeen,
                          isTierZero: node.isTierZero,
                          isOwnedObject: node.isOwnedObject,
                          ...node.properties,
                      } satisfies MungedTableRowWithId;

                      return flattenedNode;
                  })
                : [],
        [exploreTableData?.nodes]
    );

    const handleSort = useCallback(
        (sortByColumn: keyof MungedTableRowWithId) => {
            if (sortByColumn) {
                if (sortBy === sortByColumn) {
                    switch (sortOrder) {
                        case 'desc':
                            setSortOrder('asc');
                            break;
                        case 'asc':
                            setSortOrder('desc');
                            break;
                        default:
                            setSortOrder('desc');
                            break;
                    }
                } else {
                    setSortBy(sortByColumn);
                    setSortOrder('asc');
                }
            }
        },
        [sortBy, sortOrder]
    );

    const handleKebabMenuClick = useCallback(
        (e: React.MouseEvent, id: string) => {
            if (onKebabMenuClick) onKebabMenuClick({ x: e.clientX, y: e.clientY, id });
        },
        [onKebabMenuClick]
    );

    const firstTenRows = useMemo(() => rows?.slice(0, 10), [rows]);
    const makeColumnDef = useCallback(
        (rawKey: keyof MungedTableRowWithId) => {
            const key = rawKey?.toString();
            const firstTruthyValueInFirst10Rows = firstTenRows.find((row) => !!row?.[key])?.[key];
            const bestGuessAtDataType = typeof firstTruthyValueInFirst10Rows;

            return columnHelper.accessor(String(key), {
                header: () => {
                    return (
                        <ExploreTableHeaderCell
                            sortBy={sortBy}
                            sortOrder={sortOrder}
                            onClick={() => handleSort(key)}
                            headerKey={key}
                            dataType={bestGuessAtDataType}
                        />
                    );
                },
                size: isSmallColumn(key, bestGuessAtDataType) ? 100 : 250,
                cell: (info) => {
                    const value = info.getValue();

                    return (
                        <Tooltip
                            title={<p>{info.getValue()}</p>}
                            disableHoverListener={key === 'kind' || isEmpty(value)}>
                            <div data-testid={`table-cell-${key}`} className='truncate'>
                                <ExploreTableDataCell value={value} columnKey={key?.toString()} />
                            </div>
                        </Tooltip>
                    );
                },
                id: key?.toString(),
            });
        },
        [handleSort, sortOrder, sortBy, firstTenRows]
    );

    const kebabColumDefinition = useMemo(
        () =>
            columnHelper.accessor('', {
                id: 'action-menu',
                size: 50,
                maxSize: 50,
                cell: ({ row }) => (
                    <div
                        data-testid='kebab-menu'
                        onClick={(e) => handleKebabMenuClick(e, row?.original?.id)}
                        className='explore-table-cell-icon h-full flex justify-center items-center'>
                        <FontAwesomeIcon
                            icon={faEllipsis}
                            className='p-4 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0 rotate-90 dark:text-neutral-light-1 text-black'
                        />
                    </div>
                ),
            }),
        [handleKebabMenuClick]
    );

    const filteredRows: MungedTableRowWithId[] = useMemo(() => {
        const lowercaseSearchInput = searchInput?.toLowerCase();

        return rows.filter((item) => {
            const filterTargets = filterKeys.map((filterKey) => {
                const stringyValue = String(item?.[filterKey]);

                return stringyValue?.toLowerCase();
            });

            return filterTargets.some((filterTarget) => filterTarget?.includes(lowercaseSearchInput));
        });
    }, [rows, searchInput]);

    const sortedFilteredRows = useMemo(() => {
        const dataToSort = filteredRows.slice();
        if (sortBy) {
            if (sortOrder === 'asc') {
                return dataToSort.sort((a, b) => compareForExploreTableSort(a?.[sortBy], b?.[sortBy]));
            } else {
                return dataToSort.sort((a, b) => compareForExploreTableSort(b?.[sortBy], a?.[sortBy]));
            }
        }

        return dataToSort;
    }, [filteredRows, sortBy, sortOrder]);

    const allColumnDefinitions = useMemo(
        () => exploreTableData?.node_keys?.map(makeColumnDef) || [],
        [exploreTableData?.node_keys, makeColumnDef]
    );

    const selectedColumnDefinitions = useMemo(
        () => allColumnDefinitions.filter((columnDef) => selectedColumns?.[columnDef?.id || '']),
        [allColumnDefinitions, selectedColumns]
    );

    const tableColumns = useMemo(
        () => [kebabColumDefinition, ...selectedColumnDefinitions],
        [kebabColumDefinition, selectedColumnDefinitions]
    ) as DataTableProps['columns'];

    return {
        rows,
        columnOptionsForDropdown: allColumnDefinitions,
        tableColumns,
        sortedFilteredRows,
        resultsCount: rows.length,
    };
};

export default useExploreTableRowsAndColumns;
