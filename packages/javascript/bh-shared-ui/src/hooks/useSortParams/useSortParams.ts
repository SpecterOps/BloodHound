// Copyright 2026 Specter Ops, Inc.
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
