import { areArraysSimilar } from './array';

describe('areArraysSimilar', () => {
    const primeArray = [1, 2, 3, 3, 4];
    const equalArray = [1, 2, 3, 3, 4];
    const reorderedArray = [3, 2, 3, 4, 1];
    const shorterArray = [1, 2, 3, 4];
    const differentArray = ['a', 'b', 'c', 'd'];

    it('returns true if arrays are referentially equal', () => {
        expect(areArraysSimilar(primeArray, primeArray)).toEqual(true);
    });

    it('returns true if arrays are equal', () => {
        expect(areArraysSimilar(primeArray, equalArray)).toEqual(true);
    });

    it('returns true if both arrays are empty', () => {
        expect(areArraysSimilar([], [])).toEqual(true);
    });

    it('returns true if arrays have same content in different order', () => {
        expect(areArraysSimilar(primeArray, reorderedArray)).toEqual(true);
    });

    it('returns false if arrays have different length', () => {
        expect(areArraysSimilar(primeArray, shorterArray)).toEqual(false);
    });

    it('returns false if arrays have different elements', () => {
        // @ts-expect-error: types differ to test negative case
        expect(areArraysSimilar(primeArray, differentArray)).toEqual(false);
    });
});
