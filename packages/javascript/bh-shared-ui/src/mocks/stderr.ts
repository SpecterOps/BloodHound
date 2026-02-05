// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
type ErrorSilencer = {
    silence: () => void;
    restore: () => void;
};

/**
 * Returns a set of helper functions for silencing/restoring console.error output. Can be used for tests where an error is
 * expected to keep logging clean. If `silence()` is called without later calling `restore()`, error logging may be disabled for additional
 * tests.
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

/**
 * Can be used to wrap any test logic for which an error is expected, ensuring that the cleanup function is always called after the
 * callback is executed.
 */
export const withoutErrorLogging = async <T>(cb: () => T | Promise<T>): Promise<T> => {
    const silencer = errorSilencer();
    silencer.silence();

    try {
        return await cb();
    } finally {
        silencer.restore();
    }
};
