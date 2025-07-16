import { createColumnHelper, DataTable } from '@bloodhoundenterprise/doodleui';
import { faCancel, faCheck, faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useCallback, useMemo, useState } from 'react';
import { EntityField, format } from '../../utils';
import NodeIcon from '../NodeIcon';
import { ExploreTableProps, MungedTableRowWithId } from './ExploreTable';
import ExploreTableHeaderCell from './ExploreTableHeaderCell';

const columnHelper = createColumnHelper<MungedTableRowWithId>();
type DataTableProps = React.ComponentProps<typeof DataTable>;

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
                    setSortOrder('desc');
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
                header: () => (
                    <ExploreTableHeaderCell
                        sortBy={sortBy}
                        sortOrder={sortOrder}
                        onClick={() => handleSort(key)}
                        headerKey={key}
                    />
                ),
                cell: (info) => {
                    const value = info.getValue() as EntityField['value'];

                    if (info.column.id === 'nodetype') {
                        return (
                            <div className='flex justify-center items-center relative'>
                                <NodeIcon nodeType={(info.getValue() as string) || ''} />
                            </div>
                        );
                    }
                    console.log(info.column.id);
                    if (typeof value === 'boolean' || !value) {
                        return (
                            <div className='h-full w-full flex justify-center items-center'>
                                <FontAwesomeIcon
                                    icon={value ? faCheck : faCancel}
                                    color={value ? 'green' : 'lightgray'}
                                    className='scale-125'
                                />
                            </div>
                        );
                    }

                    return format({ keyprop: String(key), value, label: String(key) }) || '--';
                },
                id: String(key),
            }),
        [handleSort, sortOrder, sortBy]
    );

    const requiredColumnDefinitions = useMemo(
        () => [
            columnHelper.accessor('', {
                id: 'action-menu',
                cell: ({ row }) => (
                    <FontAwesomeIcon
                        icon={faEllipsis}
                        data-testid='kebab-menu'
                        onClick={(e) => handleKebabMenuClick(e, row?.original?.id)}
                        className='p-4 cursor-pointer hover:bg-transparent bg-transparent shadow-outer-0 rotate-90 dark:text-neutral-light-1 text-black'
                    />
                ),
            }),
        ],
        [handleKebabMenuClick]
    );

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

    const filteredRows = useMemo(
        () =>
            rows?.filter((item) => {
                const filterKeys: (keyof MungedTableRowWithId)[] = ['displayname', 'objectid'];
                const filterTargets = filterKeys.map((filterKey) => {
                    const stringyValue = String(item?.[filterKey]);

                    return stringyValue?.toLowerCase();
                });

                return filterTargets.some((filterTarget) => filterTarget?.includes(searchInput?.toLowerCase()));
            }),
        [searchInput, rows]
    );

    const sortedFilteredRows = useMemo(() => {
        const dataToSort = filteredRows.slice();
        if (sortBy) {
            if (sortOrder === 'asc') {
                dataToSort.sort((a, b) => {
                    return a[sortBy] < b[sortBy] ? 1 : -1;
                });
            } else {
                dataToSort.sort((a, b) => {
                    return a[sortBy] < b[sortBy] ? -1 : 1;
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
        () => [...requiredColumnDefinitions, ...selectedColumnDefinitions],
        [requiredColumnDefinitions, selectedColumnDefinitions]
    ) as DataTableProps['columns'];

    const columnOptionsForDropdown = useMemo(
        () => [...requiredColumnDefinitions, ...allColumnDefintions],
        [requiredColumnDefinitions, allColumnDefintions]
    );

    return {
        columnOptionsForDropdown,
        tableColumns,
        sortedFilteredRows,
    };
};

export default useExploreTableRowsAndColumns;
