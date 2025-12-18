import { useCallback, useEffect } from 'react';
import { NavigateFunction, useNavigate } from 'react-router-dom';

type KeyBindingCallbackOptions = {
    navigate: NavigateFunction;
};

interface KeyBindings extends Record<string, KeyBindings | ((options: KeyBindingCallbackOptions) => void)> {}

type KeyBindingsWithShift = { shift?: KeyBindings } & KeyBindings;

export const useKeybindings = (bindings: KeyBindingsWithShift = {}) => {
    const navigate = useNavigate();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (event.altKey && !event.metaKey) {
                event.preventDefault();

                if (event.shiftKey && !bindings.shift) {
                    return;
                }

                const bindingsMap: KeyBindingsWithShift = event.shiftKey && bindings.shift ? bindings.shift : bindings;

                const key = event.code;
                const func = bindingsMap[key] || bindingsMap[key?.toLowerCase()];

                if (typeof func === 'function') {
                    func({
                        navigate,
                    });
                }
            }
        },
        [bindings, navigate]
    );

    useEffect(() => {
        document.addEventListener('keydown', handleKeyDown);

        return () => {
            document.removeEventListener('keydown', handleKeyDown);
        };
    }, [handleKeyDown]);
};

export default useKeybindings;
