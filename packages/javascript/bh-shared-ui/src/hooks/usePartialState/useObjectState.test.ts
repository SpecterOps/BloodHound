import { act, renderHook } from '../../test-utils';
import { useObjectState } from './useObjectState';

describe('usePartialState', () => {
    it('initializes with the given state', () => {
        const { result } = renderHook(() => useObjectState({ foo: 'bar', count: 1 }));

        expect(result.current.state).toEqual({ foo: 'bar', count: 1 });
    });

    it('updates partial state', () => {
        const { result } = renderHook(() => useObjectState({ a: 1, b: 2 }));

        act(() => result.current.merge({ b: 99 }));
        expect(result.current.state).toEqual({ a: 1, b: 99 });
    });

    it('adds new keys via merge', () => {
        const { result } = renderHook(() => useObjectState<{ x: number; y?: number }>({ x: 10 }));

        act(() => result.current.merge({ y: 42 }));
        expect(result.current.state).toEqual({ x: 10, y: 42 });
    });

    it('removes a key from the state', () => {
        const { result } = renderHook(() => useObjectState({ name: 'Bilbo', age: 111 }));

        act(() => result.current.deleteKeys('age'));
        expect(result.current.state).toEqual({ name: 'Bilbo' });
    });

    it('removes multiple keys at once', () => {
        const { result } = renderHook(() => useObjectState({ a: 1, b: 2, c: 3 }));

        act(() => result.current.deleteKeys('a', 'c'));
        expect(result.current.state).toEqual({ b: 2 });
    });

    it('removal of nonexistent key is safe (no crash)', () => {
        const { result } = renderHook(() => useObjectState({ one: 1, two: 2 }));

        act(() => result.current.deleteKeys('three' as any));
        expect(result.current.state).toEqual({ one: 1, two: 2 });
    });
});
