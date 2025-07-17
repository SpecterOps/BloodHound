import { createColumnHelper, DataTable } from '@bloodhoundenterprise/doodleui';
import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useCallback, useMemo, useState } from 'react';
import { type ExploreTableProps, type MungedTableRowWithId } from './explore-table-utils';
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
            ((data &&
                Object.entries(data).map(([key, value]) => ({
                    ...value.data,
                    id: key,
                    displayname: value?.label?.text,
                }))) ||
                []) as MungedTableRowWithId[],
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
        (key: keyof MungedTableRowWithId) =>
            columnHelper.accessor(String(key), {
                header: () => {
                    const dataType = rows?.length ? typeof rows[0][key] : '';

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
                cell: (info) => (
                    <div className='max-w-80 pt-1 pb-1 overflow-hidden line-clamp-1'>
                        <ExploreTableDataCell value={info.getValue()} columnKey={key?.toString()} />
                    </div>
                ),
                id: key?.toString(),
            }),
        [handleSort, sortOrder, sortBy, rows]
    );

    const kebabColumDefinition = useMemo(
        () =>
            columnHelper.accessor('', {
                id: 'action-menu',
                cell: ({ row }) => (
                    <div className='h-full w-8 flex justify-center items-center'>
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
        columnOptionsForDropdown: allColumnDefintions,
        tableColumns,
        sortedFilteredRows,
        resultsCount: rows.length,
    };
};

export default useExploreTableRowsAndColumns;
