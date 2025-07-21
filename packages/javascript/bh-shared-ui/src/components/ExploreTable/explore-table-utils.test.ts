import { compareForExploreTableSort } from './explore-table-utils';

describe('Compare function for explore table sort', () => {
    test('function should return 1 when first param is larger, no matter the data type', () => {
        const FIRST_PARAM_IS_LARGER = 1;
        expect(compareForExploreTableSort(6, 5)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('6', '5')).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('6', 5)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(6, '5')).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, false)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, undefined)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, null)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('banana', 'apple')).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', 3)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', undefined)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', null)).toBe(FIRST_PARAM_IS_LARGER);

        const SECOND_PARAM_IS_LARGER = -1;

        expect(compareForExploreTableSort(5, 6)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('5', '6')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('5', 6)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(false, true)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(undefined, true)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(null, true)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', 'banana')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(3, 'apple')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(undefined, 'apple')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(null, 'apple')).toBe(SECOND_PARAM_IS_LARGER);
    });
});
