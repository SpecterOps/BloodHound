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
import { act, renderHook } from '../../test-utils';
import { useObjectState } from './useObjectState';

describe('useObjectState', () => {
    it('initializes with the given state', () => {
        const { result } = renderHook(() => useObjectState({ foo: 'bar', count: 1 }));

        expect(result.current.state).toEqual({ foo: 'bar', count: 1 });
    });

    it('fully replaces state', () => {
        const { result } = renderHook(() => useObjectState({ a: 1, b: 2 }));

        act(() => result.current.setState({ b: 99 }));
        expect(result.current.state).toEqual({ b: 99 });
    });

    it('updates key in state', () => {
        const { result } = renderHook(() => useObjectState({ a: 1, b: 2 }));

        act(() => result.current.applyState({ b: 99 }));
        expect(result.current.state).toEqual({ a: 1, b: 99 });
    });

    it('adds new key to state', () => {
        const { result } = renderHook(() => useObjectState<{ x: number; y?: number }>({ x: 10 }));

        act(() => result.current.applyState({ y: 42 }));
        expect(result.current.state).toEqual({ x: 10, y: 42 });
    });

    it('removes key from state', () => {
        const { result } = renderHook(() => useObjectState({ name: 'Bilbo', age: 111 }));

        act(() => result.current.deleteKeys('age'));
        expect(result.current.state).toEqual({ name: 'Bilbo' });
    });

    it('removes multiple keys from state', () => {
        const { result } = renderHook(() => useObjectState({ a: 1, b: 2, c: 3 }));

        act(() => result.current.deleteKeys('a', 'c'));
        expect(result.current.state).toEqual({ b: 2 });
    });

    it('safely handles removing nonexistent key', () => {
        const { result } = renderHook(() => useObjectState({ one: 1, two: 2 }));

        act(() => result.current.deleteKeys('three' as any));
        expect(result.current.state).toEqual({ one: 1, two: 2 });
    });

    it('keeps callbacks stable across renders', () => {
        const { result, rerender } = renderHook(() => useObjectState({ a: 1 }));
        const { applyState, deleteKeys } = result.current;
        rerender();
        expect(result.current.applyState).toBe(applyState);
        expect(result.current.deleteKeys).toBe(deleteKeys);
    });
});
