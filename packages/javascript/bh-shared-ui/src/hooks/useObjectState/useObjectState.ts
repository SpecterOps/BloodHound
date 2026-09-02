// Copyright 2025 Specter Ops, Inc.
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
import { useCallback, useState } from 'react';

export type ObjectState<T extends object> = {
    applyState: (updates: Partial<T>) => void;
    deleteKeys: (...keys: (keyof T)[]) => void;
    setState: React.Dispatch<React.SetStateAction<T>>;
    state: T;
};

/** Creates an updatable object state  */
export const useObjectState = <T extends object>(initialState: T): ObjectState<T> => {
    const [state, setState] = useState<T>(initialState);

    const applyState = useCallback((updates: Partial<T>) => {
        setState((prev) => ({ ...prev, ...updates }));
    }, []);

    const deleteKeys = useCallback((...keys: (keyof T)[]) => {
        if (keys.length === 0) {
            return;
        }

        setState((prev) => {
            const next = { ...prev };
            for (const key of keys) {
                delete next[key];
            }
            return next;
        });
    }, []);

    return { applyState, deleteKeys, setState, state };
};
