import { useCallback } from 'react';

/*
 *  This function returns a callback to be applied to the ref={} prop of an HTML element
 *  which you want to disable browzer magnifying zoom (in our case, the explore graph).
 *  This function accepts a callback to be applied to the wheel handler.
 */
const useCreateDisableZoomRef = <T extends HTMLElement>(onWheel?: (e: WheelEvent) => void) => {
    return useCallback(
        (ref: T) => {
            ref?.addEventListener(
                'wheel',
                (e) => {
                    e.preventDefault();

                    if (typeof onWheel === 'function') {
                        onWheel(e);
                    }
                },
                { passive: false }
            );
        },
        [onWheel]
    );
};

export default useCreateDisableZoomRef;
