import { createColumnHelper, DataTable } from '@bloodhoundenterprise/doodleui';
import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import { StyledGraphEdge } from 'js-client-library';
import { isEmpty } from 'lodash';
import { useCallback, useMemo, useState } from 'react';
import { cn } from '../../utils';
import { isSmallColumn, type ExploreTableProps, type MungedTableRowWithId } from './explore-table-utils';
import ExploreTableDataCell from './ExploreTableDataCell';
import ExploreTableHeaderCell from './ExploreTableHeaderCell';

const columnHelper = createColumnHelper<MungedTableRowWithId>();

type DataTableProps = React.ComponentProps<typeof DataTable>;

const filterKeys: (keyof MungedTableRowWithId)[] = ['displayname', 'objectid'];

type UseExploreTableRowsAndColumnsProps = Pick<
    ExploreTableProps,
    'onKebabMenuClick' | 'allColumnKeys' | 'selectedColumns' | 'data'
> & { searchInput: string };

const useExploreTableRowsAndColumns = ({
    onKebabMenuClick,
    searchInput,
    allColumnKeys,
    selectedColumns,
    data,
}: UseExploreTableRowsAndColumnsProps) => {
    const [sortBy, setSortBy] = useState<keyof MungedTableRowWithId>();
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>();

    const rows = useMemo(
        () =>
            (data &&
                [
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                    ...Object.entries(data),
                ].reduce((acc: MungedTableRowWithId[], curr) => {
                    const [key, value] = curr;

                    const valueAsPotentialEdge = value as StyledGraphEdge;

                    if (!!valueAsPotentialEdge.id1 || !!valueAsPotentialEdge.id2) {
                        return acc;
                    }

                    const nextRow = Object.assign({}, value.data);

                    nextRow.id = key;
                    nextRow.displayname = value?.label?.text;

                    acc.push(nextRow as MungedTableRowWithId);

                    return acc;
                }, [])) ||
            [],
        [data]
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

    const makeColumnDef = useCallback(
        (rawKey: keyof MungedTableRowWithId) => {
            const key = rawKey?.toString();

            const firstRowCellValue = rows?.length ? rows[0]?.[key] : null;
            const dataType = typeof firstRowCellValue;

            return columnHelper.accessor(String(key), {
                header: () => {
                    return (
                        <ExploreTableHeaderCell
                            sortBy={sortBy}
                            sortOrder={sortOrder}
                            onClick={() => handleSort(key)}
                            headerKey={key}
                            dataType={dataType}
                        />
                    );
                },
                size: isSmallColumn(key, firstRowCellValue) ? 100 : 250,
                cell: (info) => {
                    const value = info.getValue();
                    const useIcon = isSmallColumn(key, value);

                    return (
                        <Tooltip title={info.getValue()} disableHoverListener={key === 'nodetype' || isEmpty(value)}>
                            <div
                                className={cn('truncate', {
                                    'explore-table-cell-icon': useIcon,
                                    'explore-table-cell-string': !useIcon,
                                })}>
                                <ExploreTableDataCell value={value} columnKey={key?.toString()} />
                            </div>
                        </Tooltip>
                    );
                },
                id: key?.toString(),
            });
        },
        [handleSort, sortOrder, sortBy, rows]
    );

    const kebabColumDefinition = useMemo(
        () =>
            columnHelper.accessor('', {
                id: 'action-menu',
                size: 50,
                cell: ({ row }) => (
                    <div className='explore-table-cell-icon h-full flex justify-center items-center'>
                        <FontAwesomeIcon
                            icon={faEllipsis}
                            data-testid='kebab-menu'
                            className='p-4 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0 rotate-90 dark:text-neutral-light-1 text-black'
                            onClick={(e) => handleKebabMenuClick(e, row?.original?.id)}
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
                dataToSort.sort((a, b) => {
                    if (a[sortBy] === true) return 1;

                    return a[sortBy]?.toString().localeCompare(b[sortBy]?.toString());
                });
            } else {
                dataToSort.sort((a, b) => {
                    if (b[sortBy] === true) return 1;

                    return b[sortBy]?.toString().localeCompare(a[sortBy]?.toString());
                });
            }
        }

        return dataToSort;
    }, [filteredRows, sortBy, sortOrder]);

    const allColumnDefintions = useMemo(() => allColumnKeys?.map(makeColumnDef) || [], [allColumnKeys, makeColumnDef]);

    const selectedColumnDefinitions = useMemo(
        () => allColumnDefintions.filter((columnDef) => selectedColumns?.[columnDef?.id || '']),
        [allColumnDefintions, selectedColumns]
    );

    const tableColumns = useMemo(
        () => [kebabColumDefinition, ...selectedColumnDefinitions],
        [kebabColumDefinition, selectedColumnDefinitions]
    ) as DataTableProps['columns'];

    return {
        rows,
        columnOptionsForDropdown: allColumnDefintions,
        tableColumns,
        sortedFilteredRows,
        resultsCount: rows.length,
    };
};

export default useExploreTableRowsAndColumns;
