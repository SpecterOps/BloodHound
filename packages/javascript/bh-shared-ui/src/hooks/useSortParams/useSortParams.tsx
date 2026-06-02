import { useMemo, useState } from 'react';
import { SortOrder } from '../../types';

type UseSortParamsOptions<TSortColumn extends string> = {
    initialSortColumn?: TSortColumn;
    initialSortOrder?: SortOrder;
};

export const useSortParams = <TSortColumn extends string>({
    initialSortColumn,
    initialSortOrder,
}: UseSortParamsOptions<TSortColumn> = {}) => {
    const [sortColumn, setSortColumn] = useState<TSortColumn | undefined>(initialSortColumn);
    const [sortOrder, setSortOrder] = useState<SortOrder>(initialSortOrder);

    const sortBy = useMemo(() => {
        if (!sortColumn || !sortOrder) return undefined;
        return [`${sortOrder === 'desc' ? '-' : ''}${sortColumn}`];
    }, [sortColumn, sortOrder]);

    const clearSort = () => {
        setSortColumn(undefined);
        setSortOrder(undefined);
    };

    const handleSortChange = (column: TSortColumn) => {
        if (sortColumn !== column || sortOrder === undefined) {
            setSortColumn(column);
            setSortOrder('desc');
        } else if (sortOrder === 'desc') {
            setSortOrder('asc');
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
