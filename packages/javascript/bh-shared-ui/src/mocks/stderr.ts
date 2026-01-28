type ErrorSilencer = {
    silence: () => void;
    restore: () => void;
};

/**
 * Returns a set of helper functions for silencing/restoring console.error output. Can be used for tests where an error is
 * expected to keep logging clean.
 */
export const errorSilencer = (): ErrorSilencer => {
    let originalError: typeof console.error | null = null;

    return {
        silence: () => {
            if (originalError === null) {
                originalError = console.error;
            }
            console.error = vi.fn();
        },
        restore: () => {
            if (originalError !== null) {
                console.error = originalError;
                originalError = null;
            }
        },
    };
};

export const withoutErrorLogging = async <T>(cb: () => T | Promise<T>): Promise<T> => {
    const silencer = errorSilencer();
    silencer.silence();

    try {
        return await cb();
    } finally {
        silencer.restore();
    }
};
