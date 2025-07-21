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
import { FlatGraphResponse, GraphNodeSpreadWithProperties } from 'js-client-library';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox/ManageColumnsComboBox';

export const makeStoreMapFromColumnOptions = (columnOptions: ManageColumnsComboBoxOption[]) =>
    columnOptions.reduce(
        (acc, col) => {
            acc[col?.id] = true;

            return acc;
        },
        {} as Record<string, boolean>
    );

export type NodeClickInfo = { id: string; x: number; y: number };
export type MungedTableRowWithId = GraphNodeSpreadWithProperties & { id: string };

export const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = ['nodetype', 'isTierZero', 'name', 'objectid'];

export const requiredColumns = Object.fromEntries(REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.map((key) => [key, true]));

export const compareForExploreTableSort = (a: any, b: any) => {
    if (typeof a === 'number' || typeof b === 'number') {
        if (typeof a === 'number' && typeof b === 'number') return a > b;
        if (!b) return 1;
        if (!a) return -1;
    }

    if (typeof a === 'boolean' || typeof b === 'boolean') {
        if (a === true && b === true) return 0;
        if (a === true && !b) return 1;
        if (b === true && !a) return -1;
    }

    if (typeof a === 'undefined' || typeof b === 'undefined') {
        if (a === undefined && b === undefined) return 0;
        if (b === undefined) return 1;
        if (a === undefined) return -1;
    }

    if ((typeof a === 'object' && Object.is(a, null)) || (typeof b === 'object' && Object.is(b, null))) {
        if (a === null && b === null) return 0;
        if (b === null) return 1;
        if (a === null) return -1;
    }

    return a.toString().localeCompare(b.toString(), undefined, { numeric: true });
};

export const isSmallColumn = (key: string, value: any) => key === 'nodetype' || typeof value === 'boolean';

export const isIconField = (value: any) => typeof value === 'boolean';

export interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    data?: FlatGraphResponse;
    selectedColumns?: Record<string, boolean>;
    onRowClick?: (row: MungedTableRowWithId) => void;
    allColumnKeys?: string[];
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    selectedNode: string | null;
    onDownloadClick: () => void;
    onKebabMenuClick: (clickInfo: NodeClickInfo) => void;
}
