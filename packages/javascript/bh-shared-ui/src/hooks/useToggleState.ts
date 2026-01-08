import { useCallback, useState } from 'react';

/**
 * Manage a boolean state with a toggle helper and a direct setter.
 *
 * @param defaultValue - Initial state; if omitted the state starts as `false`. Non-boolean inputs are coerced to a boolean.
 * @returns A tuple with: the current boolean value, a function that inverts the value, and the state setter (`setState`) for direct updates.
 */
export function useToggleState(
    defaultValue?: boolean
): [boolean, () => void, React.Dispatch<React.SetStateAction<boolean>>] {
    const [value, setValue] = useState(!!defaultValue);

    const toggle = useCallback(() => {
        setValue((x) => !x);
    }, []);

    return [value, toggle, setValue];
}