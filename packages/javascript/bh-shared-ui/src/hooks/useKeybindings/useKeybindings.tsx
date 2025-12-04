import { useCallback, useEffect } from 'react';
import { NavigateFunction, useNavigate } from 'react-router-dom';

type KeyBindingCallbackOptions = {
    navigate: NavigateFunction;
};

type KeyBinding = Record<string, (options: KeyBindingCallbackOptions) => void>;

type KeyBindings = { shift?: KeyBinding } & KeyBinding;

export const globalKeybindings: KeyBindings = {
    // g: () => {
    //     console.log('Focus global search');
    // },
    // d: () => {
    //     console.log('Open Documentation panel');
    // },
    h: () => {
        console.log('Open shortcut legend');
    },
    // m: () => {
    //     console.log('Toggle Light/Dark Mode');
    // },
    n: () => {
        console.log('Open Notifications');
    },
    // s: () => {
    //     console.log('Go to Settings');
    // },
    k: () => {
        console.log('Skip to Main Content (accessibility)');
    },
    ']': () => {
        console.log('Jump to Next Landmark Region');
    },
    '[': () => {
        console.log('Jump to Previous Landmark Region');
    },
    'Shift + N': () => {
        console.log('Return Focus to Main Navigation');
    },
    ';': () => {
        console.log('Go to last focused region');
    },
};

export const attackPathsPageKeybindings: KeyBindings = {
    j: () => {
        console.log('Jump to next finding in panel');
    },
    k: () => {
        console.log('Jump to previous finding in panel CONFLICT WITH GLOBAL');
    },
    e: () => {
        console.log('Jump to Environment Selector');
    },
    'Shift + R': () => {
        console.log('Reset to Default View');
    },
};

export const explorePageKeybindings: KeyBindings = {
    // e: () => {
    //     console.log('Focus Cypher Query Editor');
    // },
    l: () => {
        console.log('Toggle Cypher Library Panel');
    },
    // r: () => {
    //     console.log('Run current Cypher query');
    // },
    // '/': () => {
    //     console.log('Jump to Search');
    // },
    // t: () => {
    //     console.log('Toggle Table View');
    // },
    i: () => {
        console.log('Toggle Node Info Panel');
    },
    '.': () => {
        console.log('Next Node Result');
    },
    ',': () => {
        console.log('Previous Node Result');
    },
    'Shift + S': () => {
        console.log('Save current query (Might need additional options later)');
    },
    'Shift + /': () => {
        console.log('Search current results');
    },
    'Shift + G': () => {
        console.log('Reset Graph View');
    },
};

export const posturePageKeybindings: KeyBindings = {
    e: () => {
        console.log('Jump to Environment Filter');
    },
    t: () => {
        console.log('Jump to Tier Filter');
    },
    y: () => {
        console.log('Jump to Time Filter');
    },
    p: () => {
        console.log('Jump to Table');
    },
    c: () => {
        console.log('Jump to Carousel');
    },
    g: () => {
        console.log('Jump to Carousel');
    },
};

export const useKeybindings = (bindings: KeyBindings = {}) => {
    const navigate = useNavigate();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (event.altKey && !event.metaKey) {
                event.preventDefault();

                const key = event.code.slice(-1);
                const bindingsMap: KeyBinding = event.shiftKey && bindings.shift ? bindings.shift : bindings;
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
