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
import { arrayMove } from '@dnd-kit/sortable';
import { Table } from '@tanstack/react-table';

export const getTextWidth = (text: string, valueType: string, ref: HTMLTableCellElement | null) => {
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d') as CanvasRenderingContext2D;

    // const theRef = valueType === 'header' ? tableHeadRef.current : tableCellRef.current;
    const computedStyle = ref ? window.getComputedStyle(ref) : null;
    const fallbackStyle = valueType === 'header' ? 'bold 1.2rem sans-serif' : 'bold .875rem sans-serif';
    context.font = computedStyle
        ? `${computedStyle?.fontWeight} ${computedStyle?.fontSize} ${computedStyle?.fontFamily}`
        : fallbackStyle;

    const metrics = context.measureText(text);
    return metrics.width + 24;
};

export const getLongestCellValue = <TData,>(table: Table<TData>, columnId: string) => {
    const allRows = table.getCoreRowModel().flatRows;
    const stringArr: string[] = allRows.map((row) => {
        return row.getValue(columnId);
    });

    const longestCellString = stringArr.reduce((longest, current) => {
        return current && current.length > longest.length ? current : longest;
    }, '');

    return longestCellString;
};

export const updateColumnOrder = (arr: string[], activeId: string | number, overId: string | number): string[] => {
    const oldIndex = arr.indexOf(activeId as string);
    const newIndex = arr.indexOf(overId as string);
    return arrayMove(arr, oldIndex, newIndex);
};

export const getExpandedColWidth = <TData,>(
    table: Table<TData>,
    colId: string,
    headerId: string,
    tableHeadRef: React.RefObject<HTMLTableCellElement>,
    tableCellRef: React.RefObject<HTMLTableCellElement>,
    isDragAndDropEnabled: boolean
) => {
    const longestCellVal = getLongestCellValue(table, colId);
    const cellWidth = getTextWidth(longestCellVal, 'cell', tableCellRef.current);
    const headerWidth = getTextWidth(headerId, 'header', tableHeadRef.current);
    const extraPadding = isDragAndDropEnabled ? 40 : 16;
    //extra padding to allow for header icons
    return cellWidth > headerWidth ? cellWidth : headerWidth + extraPadding;
};
