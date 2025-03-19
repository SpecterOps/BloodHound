// Taken from https://stackoverflow.com/a/57706747

import { useEffect, useRef } from 'react';

export const usePreviousValue = <T,>(value: T): T | undefined => {
    const ref = useRef<T>();
    useEffect(() => {
        ref.current = value;
    });
    return ref.current;
};
