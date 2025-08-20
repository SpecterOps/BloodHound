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
});
