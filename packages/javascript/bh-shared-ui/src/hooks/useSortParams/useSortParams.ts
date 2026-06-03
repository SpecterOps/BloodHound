import { useMemo, useState } from 'react';
import { SortOrder, SortOrderAscending, SortOrderDescending } from '../../types';

type UseSortParamsOptions<TSortColumn extends string> = {
    initialSortColumn?: TSortColumn;
    initialSortOrder?: SortOrder;
};

// Reusable hook for managing a table's sorting state and formatting a valid sort_by query parameter
export const useSortParams = <TSortColumn extends string>({
    initialSortColumn,
    initialSortOrder,
}: UseSortParamsOptions<TSortColumn> = {}) => {
    const [sortColumn, setSortColumn] = useState<TSortColumn | undefined>(initialSortColumn);
    const [sortOrder, setSortOrder] = useState<SortOrder>(initialSortOrder);

    const sortBy = useMemo(() => {
        if (!sortColumn || !sortOrder) return undefined;
        return `${sortOrder === SortOrderDescending ? '-' : ''}${sortColumn}`;
    }, [sortColumn, sortOrder]);

    const clearSort = () => {
        setSortColumn(undefined);
        setSortOrder(undefined);
    };

    const handleSortChange = (column: TSortColumn) => {
        if (sortColumn !== column || sortOrder === undefined) {
            setSortColumn(column);
            setSortOrder(SortOrderDescending);
        } else if (sortOrder === SortOrderDescending) {
            setSortOrder(SortOrderAscending);
        } else {
            clearSort();
        }
    };

    return {
        sortColumn,
        sortOrder,
        sortBy,
        clearSort,
        handleSortChange,
    };
};
