import { useEffect, useRef, useState } from 'react';

export const useIsMouseDragging = () => {
    const [isMouseDragging, setIsMouseDragging] = useState<boolean>(false);
    const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const handlePointerDown = () => {
        // We are setting a timeout here so that state is not changed unless the user holds down the mouse button for some period of time
        timeoutRef.current = setTimeout(() => {
            setIsMouseDragging(true);
        }, 200);
    };

    const handlePointerUp = () => {
        setIsMouseDragging(false);

        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
        }
    };

    useEffect(() => {
        document.addEventListener('mousedown', handlePointerDown);
        document.addEventListener('mouseup', handlePointerUp);

        return () => {
            document.removeEventListener('mousedown', handlePointerDown);
            document.removeEventListener('mouseup', handlePointerUp);
        };
    }, []);

    return { isMouseDragging };
};
