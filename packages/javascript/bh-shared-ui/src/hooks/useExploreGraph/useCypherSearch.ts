import { useEffect, useState } from 'react';
import { useExploreParams } from '../useExploreParams';
import { decodeCypherQuery, encodeCypherQuery } from './utils';

export const useCypherSearch = () => {
    const [cypherQuery, setCypherQuery] = useState<string>('');

    const { cypherSearch, setExploreParams } = useExploreParams();

    useEffect(() => {
        if (cypherSearch) {
            const decoded = decodeCypherQuery(cypherSearch);
            setCypherQuery(decoded);
        }
    }, [cypherSearch]);

    // create query param with a query string if it is passed, and the field state otherwise
    const performSearch = (query?: string) => {
        setExploreParams({
            searchType: 'cypher',
            cypherSearch: encodeCypherQuery(query ?? cypherQuery),
        });
    };

    return {
        cypherQuery,
        setCypherQuery,
        performSearch,
    };
};
