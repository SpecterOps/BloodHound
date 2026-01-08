'use client';

import * as React from 'react';

/**
 * Hook that provides a boolean state value with a stable toggle function.
 *
 * @param defaultValue - Optional initial state; treated as `false` when omitted or `undefined`.
 * @returns A tuple `[value, toggle, setValue]` where:
 *  - `value` is the current boolean state.
 *  - `toggle` inverts `value`.
 *  - `setValue` is the state setter (accepts a boolean or an updater function).
 */
export function useToggle(
    defaultValue?: boolean
): [boolean, () => void, React.Dispatch<React.SetStateAction<boolean>>] {
    const [value, setValue] = React.useState(!!defaultValue);

    const toggle = React.useCallback(() => {
        setValue((x) => !x);
    }, []);

    return [value, toggle, setValue];
}