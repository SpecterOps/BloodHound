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
import { FlatGraphResponse, GraphNodeSpreadWithProperties, GraphResponse } from 'js-client-library';
import { CommonKindProperties } from '../../graphSchema';
import { isGraphResponse } from '../../hooks/useExploreGraph/queries/utils';
import { KnownNodeProperties } from '../../utils/entityInfoDisplay';
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

export const REQUIRED_EXPLORE_TABLE_COLUMN_KEYS = [
    'kind',
    'label',
    'objectId',
    'isTierZero',
] satisfies KnownNodeProperties[];
export const KNOWN_NODE_KEYS = [
    ...REQUIRED_EXPLORE_TABLE_COLUMN_KEYS,
    'isOwnedObject',
    'lastSeen',
] satisfies KnownNodeProperties[];
/**
 * Keys that can be found in a nodes property bag that are lifted into the UnifiedNode type
 */
export const DUPLICATED_KNOWN_KEYS: string[] = [
    CommonKindProperties.ObjectID,
    CommonKindProperties.Name,
    CommonKindProperties.LastSeen,
];

export const getExploreTableData = (graphData: GraphResponse | FlatGraphResponse | undefined) => {
    if (!graphData || !isGraphResponse(graphData)) return;

    const nodes = graphData.data.nodes;
    const unknownNodeKeys = graphData.data.node_keys ?? [];

    const completeNodeKeys = unknownNodeKeys
        .filter((key) => !DUPLICATED_KNOWN_KEYS.includes(key)) // remove property bag duplicates of the known keys
        .concat(KNOWN_NODE_KEYS); // add known keys

    return {
        nodes,
        node_keys: completeNodeKeys,
    };
};

/**
 * Handles keys that were originally available in v8.0.1 when ExploreTable used the same data transformed for the graph.
 * But because we skip the graph specific transformations now, those keys should match the GraphNode type.
 * Users will have the new keys after resetting columns, so after enough time, this function will be useless.
 * Future 2026 DEV should remove this function entirely.
 */
export const shimGraphSpecificKeys = (selectedColumns: NonNullable<ExploreTableProps['selectedColumns']>) => {
    const graphSpecificKeysMap = {
        nodetype: 'kind',
        objectid: 'objectId',
        name: 'label',
    };

    // check if any columns in the migration exist in selected columns
    const keysToShim = Object.keys(graphSpecificKeysMap).filter((key) => selectedColumns[key]);
    // if no migrations needed, then return early with selectedColumns as-is. This will be most cases.
    if (keysToShim.length === 0) return selectedColumns;

    // otherwise clone selectedColumns and shim old column keys
    const shimmedColumns = { ...selectedColumns };

    keysToShim.forEach((key) => {
        const newKey = graphSpecificKeysMap[key as keyof typeof graphSpecificKeysMap];

        shimmedColumns[newKey] = selectedColumns[key];
        delete shimmedColumns[key];
    });

    return shimmedColumns;
};

export const requiredColumns = Object.fromEntries(REQUIRED_EXPLORE_TABLE_COLUMN_KEYS.map((key) => [key, true]));
export const isRequiredColumn = (value: string): value is (typeof REQUIRED_EXPLORE_TABLE_COLUMN_KEYS)[number] => {
    return requiredColumns[value];
};

export const compareForExploreTableSort = (a: any, b: any) => {
    if (typeof a === 'number' || typeof b === 'number') {
        if (typeof a === 'number' && typeof b === 'number') {
            if (a === b) return 0;

            return a > b ? 1 : -1;
        }
        if (!b) return 1;
        if (!a) return -1;
    }

    if (typeof a === 'boolean' || typeof b === 'boolean') {
        if (a === b) return 0;
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

export const isSmallColumn = (key: string, type: string) =>
    key === 'kind' || key === 'isTierZero' || type === 'boolean';

export interface ExploreTableProps {
    open?: boolean;
    onClose?: () => void;
    selectedColumns?: Record<string, boolean>;
    onManageColumnsChange?: (columns: ManageColumnsComboBoxOption[]) => void;
    onKebabMenuClick: (clickInfo: NodeClickInfo) => void;
}
