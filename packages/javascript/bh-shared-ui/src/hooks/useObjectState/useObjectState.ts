import { useCallback, useState } from 'react';

export type ObjectState<T extends object> = {
    applyState: (updates: Partial<T>) => void;
    deleteKeys: (...keys: (keyof T)[]) => void;
    setState: React.Dispatch<React.SetStateAction<T>>;
    state: T;
};

/** Creates an updatable object state  */
export const useObjectState = <T extends object>(initialState: T): ObjectState<T> => {
    const [state, setState] = useState<T>(initialState);

    const applyState = useCallback((updates: Partial<T>) => {
        setState((prev) => ({ ...prev, ...updates }));
    }, []);

    const deleteKeys = useCallback((...keys: (keyof T)[]) => {
        if (keys.length === 0) {
            return;
        }

        setState((prev) => {
            const next = { ...prev };
            for (const key of keys) {
                delete next[key];
            }
            return next;
        });
    }, []);

    return { applyState, deleteKeys, setState, state };
};
