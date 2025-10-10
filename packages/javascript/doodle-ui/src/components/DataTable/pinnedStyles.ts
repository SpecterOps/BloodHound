import { CSSProperties } from 'react';

export const getCommonPinnedStyles = (startOffset: number): CSSProperties => {
    return {
        position: 'sticky',
        zIndex: 30,
        left: `${startOffset}px`,
    };
};

export const getConditionalPinnedStyles = (
    isLastPinnedColumn: boolean,
    isLastRow: boolean,
    isSelected: boolean
): CSSProperties => {
    let result: CSSProperties = {};

    // Cells around the edges of the pinned area have a box shadow, along with a clip path to prevent the shadow from showing around
    // the interior edges
    if (isLastPinnedColumn || isLastRow) {
        result = {
            filter: 'drop-shadow(rgba(0, 0, 0, 0.1) 2px 0px 6px)',
            clipPath: `inset(0 ${isLastPinnedColumn ? -20 : 0}px ${isLastRow ? -20 : 0}px 0`,
        };
    }

    // If a row is selected, the pinned cells in that row get an overriding clip path that shows the underlying row's inset box shadow
    if (isSelected) {
        result.clipPath = 'inset(3px -16px 3px 0)';
    }

    return result;
};
