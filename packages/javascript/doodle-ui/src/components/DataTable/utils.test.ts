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
import { createRef } from 'react';
import { Table } from '@tanstack/react-table';
import { getExpandedColWidth, getLongestCellValue, getTextWidth, updateColumnOrder } from './utils';

const createMockTable = <TData,>(rows: Record<string, TData>[]): Table<TData> => {
    const flatRows = rows.map((row) => ({
        getValue: (columnId: string) => row[columnId],
    }));
    return { getCoreRowModel: () => ({ flatRows }) } as unknown as Table<TData>;
};

describe('getTextWidth', () => {
    let mockMeasureText: ReturnType<typeof vi.fn>;
    let mockContext: { font: string; measureText: ReturnType<typeof vi.fn> };

    beforeEach(() => {
        mockMeasureText = vi.fn().mockReturnValue({ width: 100 });
        mockContext = { font: '', measureText: mockMeasureText };
        vi.spyOn(document, 'createElement').mockReturnValue({
            getContext: vi.fn().mockReturnValue(mockContext),
        } as unknown as HTMLCanvasElement);
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('uses header fallback font style when ref is null and valueType is "header"', () => {
        getTextWidth('text', 'header', null);
        expect(mockContext.font).toBe('bold 1.2rem sans-serif');
    });

    it('uses cell fallback font style when ref is null and valueType is not "header"', () => {
        getTextWidth('text', 'cell', null);
        expect(mockContext.font).toBe('bold .875rem sans-serif');
    });

    it('uses computed style from ref when ref is provided', () => {
        vi.spyOn(window, 'getComputedStyle').mockReturnValue({
            fontWeight: 'bold',
            fontSize: '16px',
            fontFamily: 'Arial',
        } as unknown as CSSStyleDeclaration);

        getTextWidth('text', 'header', document.createElement('td') as HTMLTableCellElement);

        expect(mockContext.font).toBe('bold 16px Arial');
    });

    it('returns measured text width plus 24px padding', () => {
        const result = getTextWidth('hello', 'cell', null);
        expect(result).toBe(124); // 100 + 24
    });

    it('passes the provided text string to measureText', () => {
        getTextWidth('hello world', 'cell', null);
        expect(mockMeasureText).toHaveBeenCalledWith('hello world');
    });
});

describe('getLongestCellValue', () => {
    it('returns the longest string value across all rows for the given column', () => {
        const table = createMockTable([{ name: 'short' }, { name: 'a very long name' }, { name: 'medium name' }]);
        expect(getLongestCellValue(table, 'name')).toBe('a very long name');
    });

    it('returns an empty string when there are no rows', () => {
        const table = createMockTable([]);
        expect(getLongestCellValue(table, 'name')).toBe('');
    });

    it('skips null values when finding the longest string', () => {
        const table = createMockTable([{ name: null }, { name: 'valid' }, { name: null }]);
        expect(getLongestCellValue(table, 'name')).toBe('valid');
    });

    it('returns empty string when all values are null or falsy', () => {
        const table = createMockTable([{ name: null }, { name: '' }]);
        expect(getLongestCellValue(table, 'name')).toBe('');
    });
});

describe('updateColumnOrder', () => {
    it('moves an item forward in the array', () => {
        expect(updateColumnOrder(['a', 'b', 'c', 'd'], 'b', 'd')).toEqual(['a', 'c', 'd', 'b']);
    });

    it('moves an item backward in the array', () => {
        expect(updateColumnOrder(['a', 'b', 'c', 'd'], 'd', 'b')).toEqual(['a', 'd', 'b', 'c']);
    });

    it('returns the same order when activeId and overId are the same element', () => {
        expect(updateColumnOrder(['a', 'b', 'c'], 'b', 'b')).toEqual(['a', 'b', 'c']);
    });

    it('can move an item to the beginning of the array', () => {
        expect(updateColumnOrder(['a', 'b', 'c'], 'c', 'a')).toEqual(['c', 'a', 'b']);
    });

    it('can move an item to the end of the array', () => {
        expect(updateColumnOrder(['a', 'b', 'c'], 'a', 'c')).toEqual(['b', 'c', 'a']);
    });
});

describe('getExpandedColWidth', () => {
    let mockMeasureText: ReturnType<typeof vi.fn>;
    let mockContext: { font: string; measureText: ReturnType<typeof vi.fn> };

    beforeEach(() => {
        mockMeasureText = vi.fn();
        mockContext = { font: '', measureText: mockMeasureText };
        vi.spyOn(document, 'createElement').mockReturnValue({
            getContext: vi.fn().mockReturnValue(mockContext),
        } as unknown as HTMLCanvasElement);
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('returns cellWidth when cell content is wider than the header', () => {
        const table = createMockTable([{ col: 'a very long cell value' }]);
        // first measureText call → cell, second → header
        mockMeasureText.mockReturnValueOnce({ width: 200 }).mockReturnValueOnce({ width: 50 });
        const tableHeadRef = createRef<HTMLTableCellElement>();
        const tableCellRef = createRef<HTMLTableCellElement>();

        const result = getExpandedColWidth(table, 'col', 'Header', tableHeadRef, tableCellRef, false);

        expect(result).toBe(224); // cellWidth = 200 + 24
    });

    it('returns headerWidth plus 16px extra padding when header is wider and DnD is disabled', () => {
        const table = createMockTable([{ col: 'short' }]);
        mockMeasureText.mockReturnValueOnce({ width: 50 }).mockReturnValueOnce({ width: 200 });
        const tableHeadRef = createRef<HTMLTableCellElement>();
        const tableCellRef = createRef<HTMLTableCellElement>();

        const result = getExpandedColWidth(table, 'col', 'Very Long Header', tableHeadRef, tableCellRef, false);

        expect(result).toBe(240); // headerWidth = 200 + 24, plus extraPadding 16
    });

    it('returns headerWidth plus 40px extra padding when header is wider and DnD is enabled', () => {
        const table = createMockTable([{ col: 'short' }]);
        mockMeasureText.mockReturnValueOnce({ width: 50 }).mockReturnValueOnce({ width: 200 });
        const tableHeadRef = createRef<HTMLTableCellElement>();
        const tableCellRef = createRef<HTMLTableCellElement>();

        const result = getExpandedColWidth(table, 'col', 'Very Long Header', tableHeadRef, tableCellRef, true);

        expect(result).toBe(264); // headerWidth = 200 + 24, plus extraPadding 40
    });
});

