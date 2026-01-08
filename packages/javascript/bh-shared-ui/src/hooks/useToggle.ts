'use client';

import * as React from 'react';

export function useToggle(
    defaultValue?: boolean
): [boolean, () => void, React.Dispatch<React.SetStateAction<boolean>>] {
    const [value, setValue] = React.useState(!!defaultValue);

    const toggle = React.useCallback(() => {
        setValue((x) => !x);
    }, []);

    return [value, toggle, setValue];
}
