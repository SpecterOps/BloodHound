import { act, renderHook } from '../../test-utils';
import { SortOrderAscending, SortOrderDescending } from '../../types';
import { useSortParams } from './useSortParams';

const COLUMN_1 = 'column1';
const COLUMN_2 = 'column2';
const COLUMN_3 = 'column3';

type TestColumns = typeof COLUMN_1 | typeof COLUMN_2 | typeof COLUMN_3;

describe('useSortParams', () => {
    it('returns undefined for sort params when no initial state is provided', () => {
        const { result } = renderHook(() => useSortParams<TestColumns>());

        expect(result.current.sortColumn).toBeUndefined();
        expect(result.current.sortOrder).toBeUndefined();
        expect(result.current.sortBy).toBeUndefined();
    });

    it('respects the initial sort state parameters passed in', () => {
        const { result } = renderHook(() =>
            useSortParams<TestColumns>({
                initialSortColumn: COLUMN_1,
                initialSortOrder: SortOrderDescending,
            })
        );

        expect(result.current.sortColumn).toEqual(COLUMN_1);
        expect(result.current.sortOrder).toEqual(SortOrderDescending);
        expect(result.current.sortBy).toEqual(`-${COLUMN_1}`);
    });

    it('sets descending sort state for a new sorted column', () => {
        const { result } = renderHook(() => useSortParams<TestColumns>());

        act(() => result.current.handleSortChange(COLUMN_1));

        expect(result.current.sortColumn).toEqual(COLUMN_1);
        expect(result.current.sortOrder).toEqual(SortOrderDescending);
        expect(result.current.sortBy).toEqual(`-${COLUMN_1}`);
    });

    it('toggles a descending sort to ascending when sorting the same column', () => {
        const { result } = renderHook(() =>
            useSortParams<TestColumns>({
                initialSortColumn: COLUMN_1,
                initialSortOrder: SortOrderDescending,
            })
        );

        act(() => result.current.handleSortChange(COLUMN_1));

        expect(result.current.sortColumn).toEqual(COLUMN_1);
        expect(result.current.sortOrder).toEqual(SortOrderAscending);
        expect(result.current.sortBy).toEqual(COLUMN_1);
    });

    it('clears an ascending sort when sorting the same column', () => {
        const { result } = renderHook(() =>
            useSortParams<TestColumns>({
                initialSortColumn: COLUMN_1,
                initialSortOrder: SortOrderAscending,
            })
        );

        act(() => result.current.handleSortChange(COLUMN_1));

        expect(result.current.sortColumn).toBeUndefined();
        expect(result.current.sortOrder).toBeUndefined();
        expect(result.current.sortBy).toBeUndefined();
    });

    it('resets to descending sort when sorting a different column', () => {
        const { result } = renderHook(() =>
            useSortParams<TestColumns>({
                initialSortColumn: COLUMN_1,
                initialSortOrder: SortOrderAscending,
            })
        );

        act(() => result.current.handleSortChange(COLUMN_2));

        expect(result.current.sortColumn).toEqual(COLUMN_2);
        expect(result.current.sortOrder).toEqual(SortOrderDescending);
        expect(result.current.sortBy).toEqual(`-${COLUMN_2}`);
    });

    it('clears sort state', () => {
        const { result } = renderHook(() =>
            useSortParams<TestColumns>({
                initialSortColumn: COLUMN_3,
                initialSortOrder: SortOrderDescending,
            })
        );

        act(() => result.current.clearSort());

        expect(result.current.sortColumn).toBeUndefined();
        expect(result.current.sortOrder).toBeUndefined();
        expect(result.current.sortBy).toBeUndefined();
    });
});
