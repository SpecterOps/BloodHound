import { useSearchParams } from 'react-router-dom';
import { setParamsFactory } from '../utils/searchParams';

interface ExploreQuery {
    primarySearch: string | null;
    secondarySearch: string | null;
    cypherSearch: string | null;
    searchType: 'node' | 'pathfinding' | 'cypher' | 'relationship' | 'composition' | null;
}

export type ExploreQueryParams = Partial<ExploreQuery>;

const parseSearchQuery = (param: string | null) => {
    if (
        param === 'node' ||
        param === 'pathfinding' ||
        param === 'cypher' ||
        param === 'relationship' ||
        param === 'composition'
    ) {
        return param;
    }
    return null;
};

interface useExploreParamsReturn extends ExploreQueryParams {
    setExploreParams: (params: ExploreQueryParams) => void;
}

export const useExploreParams = (): useExploreParamsReturn => {
    const [searchParams, setSearchParams] = useSearchParams();

    const primarySearch = searchParams.get('primarySearch');
    const secondarySearch = searchParams.get('secondarySearch');
    const cypherSearch = searchParams.get('cypherSearch');
    const searchType = parseSearchQuery(searchParams.get('searchType'));

    const setExploreParams = setParamsFactory<ExploreQueryParams>(setSearchParams, [
        'primarySearch',
        'secondarySearch',
        'cypherSearch',
        'searchType',
    ]);

    return { primarySearch, secondarySearch, cypherSearch, searchType, setExploreParams };
};
