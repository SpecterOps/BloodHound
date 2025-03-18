import { createMemoryHistory } from 'history';
import { act, renderHook, waitFor } from '../../test-utils';
import { useCypherSearch } from './useCypherSearch';

const TEST_CYPHER = 'match (n) return n limit 10';
const TEST_BASE64 = 'bWF0Y2ggKG4pIHJldHVybiBuIGxpbWl0IDEw';

describe('useCypherSearch', () => {
    it('stores the state of a search term without modifying the query params', () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useCypherSearch(), { history });

        expect(history.location.search).toBe('');

        act(() => hook.result.current.setCypherQuery(TEST_CYPHER));

        expect(hook.result.current.cypherQuery).toBe(TEST_CYPHER);
        expect(history.location.search).toBe('');
    });

    it("upon performing a search, updates the url params with the base64-encoded current query and sets searchType to 'cypher'", async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useCypherSearch(), { history });

        await act(async () => hook.result.current.setCypherQuery(TEST_CYPHER));
        await act(async () => hook.result.current.performSearch());

        expect(history.location.search).toContain('searchType=cypher');
        expect(history.location.search).toContain(`cypherSearch=${TEST_BASE64}`);
    });

    it('optionally allows the consumer to pass a query to performSearch() and adds that query to the url', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useCypherSearch(), { history });

        await act(async () => hook.result.current.performSearch(TEST_CYPHER));

        expect(history.location.search).toContain('searchType=cypher');
        expect(history.location.search).toContain(`cypherSearch=${TEST_BASE64}`);
    });

    it('populates the cypher search field with the decoded query when the associated query params are set', async () => {
        const url = `?searchType=cypher&cypherSearch=${TEST_BASE64}`;
        const history = createMemoryHistory({ initialEntries: [url] });

        const hook = renderHook(() => useCypherSearch(), { history });

        await waitFor(() => expect(hook.result.current.cypherQuery).toEqual(TEST_CYPHER));
    });
});
