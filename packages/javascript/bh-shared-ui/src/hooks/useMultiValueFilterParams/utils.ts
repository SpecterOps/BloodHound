import { MultiValueSelection } from './types';

// dedupes and sorts the selected values, also accounts for the ambiguous case of { kind: 'none' } === { kind: 'some', values: [] }
export const normalizeSelection = (selection: MultiValueSelection): MultiValueSelection => {
    if (selection.kind !== 'some') {
        return selection;
    }

    const normalizedValues = [...new Set(selection.values)].filter(Boolean).sort();

    if (normalizedValues.length > 0) {
        return { kind: 'some', values: normalizedValues };
    } else {
        return { kind: 'none' };
    }
};

// Strictly compares equality between two selections, arguments should be normalized beforehand
export const areSelectionsEqual = (left: MultiValueSelection, right: MultiValueSelection): boolean => {
    if (left.kind !== right.kind) return false;

    if (left.kind === 'some' && right.kind === 'some') {
        return (
            left.values.length === right.values.length &&
            left.values.every((value, index) => value === right.values[index])
        );
    }

    return true;
};
