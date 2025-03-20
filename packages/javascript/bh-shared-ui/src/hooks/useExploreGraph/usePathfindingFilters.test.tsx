import { createMemoryHistory } from 'history';
import { act, renderHook } from '../../test-utils';
import { INITIAL_FILTERS } from './queries';
import { usePathfindingFilters } from './usePathfindingFilters';
import { extractEdgeTypes } from './utils';

const TEST_FILTER = INITIAL_FILTERS[0];

describe('usePathfindingFilters', () => {
    it('initializes the list with all filters checked by default', () => {
        const hook = renderHook(() => usePathfindingFilters());
        expect(hook.result.current.selectedFilters).toEqual(INITIAL_FILTERS);
    });

    it('will update the selected filters based on the values stored in query params when the initialize function is called', async () => {
        const history = createMemoryHistory({ initialEntries: [`?pathFilters=${TEST_FILTER.edgeType}`] });
        const hook = renderHook(() => usePathfindingFilters(), { history });

        expect(hook.result.current.selectedFilters).toEqual(INITIAL_FILTERS);

        await act(() => hook.result.current.initialize());

        const edgeTypesInFilter = extractEdgeTypes(hook.result.current.selectedFilters);
        expect(edgeTypesInFilter).toEqual([TEST_FILTER.edgeType]);
    });

    it('allows you to update the list of selected filters, only updating the url after calling the apply function', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => usePathfindingFilters(), { history });

        await act(() => hook.result.current.handleUpdateFilters([TEST_FILTER]));

        expect(hook.result.current.selectedFilters).toEqual([TEST_FILTER]);
        expect(history.location.search).toEqual('');

        await act(() => hook.result.current.handleApplyFilters());

        expect(hook.result.current.selectedFilters).toEqual([TEST_FILTER]);
        expect(history.location.search).toEqual(`?pathFilters=${TEST_FILTER.edgeType}`);
    });

    it('allows you to cancel filter updates before applying them to the url and reset to the default filter state', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => usePathfindingFilters(), { history });

        await act(() => hook.result.current.handleUpdateFilters([TEST_FILTER]));
        await act(() => hook.result.current.handleCancelFilters());

        expect(hook.result.current.selectedFilters).toEqual(INITIAL_FILTERS);
        expect(history.location.search).toEqual('');
    });
});
