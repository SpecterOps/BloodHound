import { RefObject, useEffect, useState } from 'react';

export function useMeasure(ref: RefObject<HTMLElement>) {
    const [width, setWidth] = useState(0);
    const [height, setHeight] = useState(0);

    useEffect(() => {
        if (!ref.current) return;

        const updateMeasurements = () => {
            if (ref.current) {
                setWidth(ref.current.clientWidth);
                setHeight(ref.current.clientHeight);
            }
        };

        updateMeasurements();
        const observer = new ResizeObserver(updateMeasurements);
        observer.observe(ref.current);

        return () => observer.disconnect();
    }, [ref]);

    return [width, height];
}
