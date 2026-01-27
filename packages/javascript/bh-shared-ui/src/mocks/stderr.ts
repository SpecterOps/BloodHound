type ErrorSilencer = {
    silence: () => void;
    restore: () => void;
};

/**
 * Returns a set of helper functions for silencing/restoring console.error output. Can be used for tests where an error is
 * expected to keep logging clean.
 */
export const errorSilencer = (): ErrorSilencer => {
    let originalError = console.error;

    return {
        silence: () => {
            originalError = console.error;
            console.error = vi.fn();
        },
        restore: () => {
            console.error = originalError;
        },
    };
};
