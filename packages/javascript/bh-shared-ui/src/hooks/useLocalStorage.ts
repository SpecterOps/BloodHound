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

import { useCallback, useEffect, useSyncExternalStore } from 'react';

/**
 * Dispatches a synthetic `storage` event on the window, allowing same-tab
 * listeners to react to localStorage changes as if they originated from
 * another tab.
 */
function dispatchStorageEvent(key: string, newValue: string | null): void {
    window.dispatchEvent(new StorageEvent('storage', { key, newValue }));
}

/**
 * Server snapshot stub required by `useSyncExternalStore`. Always throws
 * because this hook is client-only and must not be called during SSR.
 */
function getLocalStorageServerSnapshot(): never {
    throw Error('useLocalStorage is a client-only hook');
}

/**
 * Removes an item from localStorage and notifies same-tab listeners by
 * dispatching a synthetic `storage` event with `newValue` set to `null`.
 */
function removeLocalStorageItem(key: string): void {
    window.localStorage.removeItem(key);
    dispatchStorageEvent(key, null);
}

/**
 * Serialises `value` to JSON, writes it to localStorage under `key`, and
 * dispatches a synthetic `storage` event so same-tab listeners stay in sync.
 */
function setLocalStorageItem<T>(key: string, value: T): void {
    const stringifiedValue = JSON.stringify(value);
    window.localStorage.setItem(key, stringifiedValue);
    dispatchStorageEvent(key, stringifiedValue);
}

/**
 * Subscribes `callback` to the window `storage` event and returns an
 * unsubscribe function, as required by `useSyncExternalStore`.
 */
function useLocalStorageSubscribe(callback: () => void): () => void {
    window.addEventListener('storage', callback);
    return () => window.removeEventListener('storage', callback);
}

/**
 * Safely parses a JSON string, returning `fallback` when the input is `null`
 * or contains malformed JSON instead of throwing.
 */
function safeParse<T>(value: string | null, fallback: T): T {
    if (value === null) return fallback;
    try {
        return JSON.parse(value) as T;
    } catch (e) {
        console.warn('useLocalStorage: failed to parse stored value', e);
        return fallback;
    }
}

/**
 * A React hook that syncs state with a localStorage key.
 *
 * - Reads and writes are JSON-serialised automatically.
 * - Changes made in other tabs (via the native `storage` event) and in the
 *   same tab (via the synthetic event dispatched by the helpers above) both
 *   trigger a re-render.
 * - If the key is absent from localStorage and `initialValue` is provided,
 *   the value is written on first render.
 * - Passing `undefined` or `null` to the setter removes the key entirely.
 *
 * @param key - The localStorage key to read from and write to.
 * @param initialValue - Value to use (and persist) when the key is absent.
 * @returns A stateful tuple `[value, setValue]` analogous to `useState`.
 *
 * @example
 * const [theme, setTheme] = useLocalStorage('theme', 'light');
 */
export function useLocalStorage<T>(key: string, initialValue: T): [T, (v: T | ((prev: T) => T)) => void] {
    const getSnapshot = () => window.localStorage.getItem(key);

    const store = useSyncExternalStore(useLocalStorageSubscribe, getSnapshot, getLocalStorageServerSnapshot);

    const setState = useCallback(
        (v: T | ((prev: T) => T)) => {
            try {
                const currentRaw = window.localStorage.getItem(key) ?? 'null';
                const current = safeParse<T>(currentRaw, null as unknown as T);
                const baseline = current === null ? initialValue : current;
                const nextState = typeof v === 'function' ? (v as (prev: T) => T)(baseline) : v;

                if (nextState === undefined || nextState === null) {
                    removeLocalStorageItem(key);
                } else {
                    setLocalStorageItem(key, nextState);
                }
            } catch (e) {
                console.warn(e);
            }
        },
        [key, initialValue]
    );

    useEffect(() => {
        if (window.localStorage.getItem(key) === null && typeof initialValue !== 'undefined') {
            setLocalStorageItem(key, initialValue);
        }
    }, [key, initialValue]);

    return [safeParse(store, initialValue), setState];
}
