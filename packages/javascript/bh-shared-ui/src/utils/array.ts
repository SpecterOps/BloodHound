/** Returns true if array `a` and `b` contain the same elements in any order */
export const areArraysSimilar = <T>(a: T[], b: T[], compareFn?: (i: T, j: T) => number) => {
    if (a.length !== b.length) {
        return false;
    }

    const sortedA = compareFn ? [...a].sort(compareFn) : [...a].sort();
    const sortedB = compareFn ? [...b].sort(compareFn) : [...b].sort();

    return sortedA.every((item, index) => item === sortedB[index]);
};
