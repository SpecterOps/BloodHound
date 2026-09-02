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

import { act, renderHook } from '@testing-library/react';
import { useLocalStorage } from './useLocalStorage';

const TEST_KEY = 'test-key';

afterEach(() => {
    localStorage.clear();
});

describe('useLocalStorage', () => {
    describe('initial value', () => {
        it('returns initialValue when the key is absent from localStorage', () => {
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, 'default'));
            expect(result.current[0]).toBe('default');
        });

        it('writes initialValue to localStorage when the key is absent', () => {
            renderHook(() => useLocalStorage(TEST_KEY, 'persisted'));
            expect(localStorage.getItem(TEST_KEY)).toBe(JSON.stringify('persisted'));
        });

        it('reads and parses an existing value from localStorage', () => {
            localStorage.setItem(TEST_KEY, JSON.stringify('existing'));
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, 'default'));
            expect(result.current[0]).toBe('existing');
        });

        it('does not overwrite an existing localStorage value with initialValue', () => {
            localStorage.setItem(TEST_KEY, JSON.stringify('existing'));
            renderHook(() => useLocalStorage(TEST_KEY, 'default'));
            expect(localStorage.getItem(TEST_KEY)).toBe(JSON.stringify('existing'));
        });
    });

    describe('setValue', () => {
        it('stores a string value in localStorage', () => {
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, ''));
            act(() => result.current[1]('hello'));
            expect(localStorage.getItem(TEST_KEY)).toBe(JSON.stringify('hello'));
            expect(result.current[0]).toBe('hello');
        });

        it('stores a number value in localStorage', () => {
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, 0));
            act(() => result.current[1](42));
            expect(result.current[0]).toBe(42);
        });

        it('stores a boolean value in localStorage', () => {
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, false));
            act(() => result.current[1](true));
            expect(result.current[0]).toBe(true);
        });

        it('stores an object value in localStorage', () => {
            const initialValue = { a: 1 };
            const nextValue = { a: 2, b: 'x' };
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, initialValue));
            act(() => result.current[1](nextValue));
            expect(result.current[0]).toEqual(nextValue);
            expect(localStorage.getItem(TEST_KEY)).toBe(JSON.stringify(nextValue));
        });

        it('stores an array value in localStorage', () => {
            const { result } = renderHook(() => useLocalStorage<number[]>(TEST_KEY, []));
            act(() => result.current[1]([1, 2, 3]));
            expect(result.current[0]).toEqual([1, 2, 3]);
        });

        it('removes the key when null is passed to the setter', () => {
            localStorage.setItem(TEST_KEY, JSON.stringify('existing'));
            const { result } = renderHook(() => useLocalStorage<string | null>(TEST_KEY, null));
            act(() => result.current[1](null));
            expect(localStorage.getItem(TEST_KEY)).toBeNull();
        });

        it('removes the key when undefined is passed to the setter', () => {
            localStorage.setItem(TEST_KEY, JSON.stringify('existing'));
            const { result } = renderHook(() => useLocalStorage<string | undefined>(TEST_KEY, undefined));
            act(() => result.current[1](undefined));
            expect(localStorage.getItem(TEST_KEY)).toBeNull();
        });
    });

    describe('functional updater', () => {
        it('receives the current stored value as the previous state', () => {
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, 1));
            act(() => result.current[1]((prev) => prev + 10));
            expect(result.current[0]).toBe(11);
        });

        it('treats a missing key as null when computing the next value', () => {
            const { result } = renderHook(() => useLocalStorage<number | null>(TEST_KEY, null));
            act(() => result.current[1]((prev) => (prev === null ? 99 : prev)));
            expect(result.current[0]).toBe(99);
        });
    });

    describe('storage event synchronisation', () => {
        it('dispatches a storage event when a value is set', () => {
            const listener = vi.fn();
            window.addEventListener('storage', listener);
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, ''));
            // Clear any storage events fired by the init effect so we only
            // assert on the event dispatched by the setter below.
            listener.mockClear();
            act(() => result.current[1]('new'));
            expect(listener).toHaveBeenCalledTimes(1);
            window.removeEventListener('storage', listener);
        });

        it('dispatches a storage event with null newValue when the key is removed', () => {
            localStorage.setItem(TEST_KEY, JSON.stringify('value'));
            const listener = vi.fn();
            window.addEventListener('storage', listener);
            const { result } = renderHook(() => useLocalStorage<string | null>(TEST_KEY, null));
            act(() => result.current[1](null));
            const event: StorageEvent = listener.mock.calls[0][0];
            expect(event.newValue).toBeNull();
            window.removeEventListener('storage', listener);
        });

        it('updates state when an external storage event is fired for the same key', () => {
            const { result } = renderHook(() => useLocalStorage(TEST_KEY, 'initial'));
            act(() => {
                localStorage.setItem(TEST_KEY, JSON.stringify('external'));
                window.dispatchEvent(
                    new StorageEvent('storage', { key: TEST_KEY, newValue: JSON.stringify('external') })
                );
            });
            expect(result.current[0]).toBe('external');
        });
    });
});
