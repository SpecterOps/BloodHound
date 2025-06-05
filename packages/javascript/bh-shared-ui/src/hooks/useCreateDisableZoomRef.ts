import { useCallback } from 'react';

/*
 *  This function returns a callback to be applied to the ref={} prop of an HTML element
 *  which you want to disable browzer magnifying zoom (in our case, the explore graph).
 *  This function accepts a callback to be applied to the wheel handler.
 */
const useCreateDisableZoomRef = <T extends HTMLElement>(onWheel?: (e: WheelEvent) => void) => {
    return useCallback(
        (ref: T) => {
            // Clean up existing listener if any
            ref?.removeEventListener('wheel', (ref as any)._wheelHandler as any);

            const wheelHandler = (e: WheelEvent) => {
                e.preventDefault();
                if (typeof onWheel === 'function') {
                    onWheel(e);
                }
            };

            if (ref) {
                // Store reference for cleanup
                (ref as any)._wheelHandler = wheelHandler;

                ref?.addEventListener('wheel', wheelHandler, { passive: false });
            }
        },
        [onWheel]
    );
};

export default useCreateDisableZoomRef;
