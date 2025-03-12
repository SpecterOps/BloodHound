import { useEffect, useState } from 'react';
import { useExploreParams } from '../useExploreParams';

export const useCypherSearch = () => {
    const [cypherQuery, setCypherQuery] = useState<string>('');

    const { cypherSearch, setExploreParams } = useExploreParams();

    useEffect(() => {
        if (cypherSearch) {
            const decoded = atob(cypherSearch);
            setCypherQuery(decoded);
        }
    }, [cypherSearch]);

    const performSearch = () => {
        setExploreParams({
            searchType: 'cypher',
            cypherSearch: btoa(cypherQuery),
        });
    };

    return {
        cypherQuery,
        setCypherQuery,
        performSearch,
    };
};
