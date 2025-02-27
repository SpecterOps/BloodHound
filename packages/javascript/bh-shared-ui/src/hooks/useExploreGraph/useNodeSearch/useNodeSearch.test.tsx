import { act, renderHook } from '@testing-library/react';
import { ReactNode } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { useNodeSearch } from './useNodeSearch';

const TEST_STRING_1 = 'Test1';
const TEST_STRING_2 = 'test2';

const wrapper = ({ children }: { children: ReactNode }) => <BrowserRouter>{children}</BrowserRouter>;

describe('useNodeSearch', () => {
    it('stores the state of a search term without modifying the query params', () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        expect(window.location.search).toBe('');

        act(() => hook.result.current.editSourceNode(TEST_STRING_1));

        expect(hook.result.current.searchTerm).toBe(TEST_STRING_1);
        expect(window.location.search).toBe('');
    });

    it("upon selecting a source node, updates the URL with a searchType of 'node' and primarySearch of the node name", () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        act(() => hook.result.current.selectSourceNode({ name: TEST_STRING_1, objectid: TEST_STRING_2 }));

        expect(window.location.search).toContain(`primarySearch=${TEST_STRING_1}`);
        expect(window.location.search).toContain('searchType=node');
    });

    it('uses the objectid as the search term in the case of a node with no name', () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        act(() => hook.result.current.selectSourceNode({ objectid: TEST_STRING_2 }));

        expect(window.location.search).toContain(`primarySearch=${TEST_STRING_2}`);
        expect(window.location.search).toContain('searchType=node');
    });

    it('does not add a query param if the search term is empty', () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        act(() => hook.result.current.selectSourceNode({ name: '', objectid: '' }));

        expect(window.location.search).not.toContain('primarySearch');
    });
});
