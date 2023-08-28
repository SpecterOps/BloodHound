import { RefObject, useEffect } from 'react';

// wrap handler in useCallback before passing in to prevent rerunning every render
const useOnClickOutside = (ref: RefObject<any>, handler: (e: Event) => void) => {
    useEffect(() => {
        const listener = (e: Event) => {
            if (!ref.current || ref.current.contains(e.target)) {
                return;
            }
            handler(e);
        };
        document.addEventListener('mousedown', listener);
        document.addEventListener('touchstart', listener);
        return () => {
            document.removeEventListener('mousedown', listener);
            document.removeEventListener('touchstart', listener);
        };
    }, [ref, handler]);
};

export default useOnClickOutside;
