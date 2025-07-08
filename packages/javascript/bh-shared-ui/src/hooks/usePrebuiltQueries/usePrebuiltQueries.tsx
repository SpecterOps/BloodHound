import { CommonSearches as prebuiltSearchListAGI } from '../../commonSearchesAGI';
import { CommonSearches as prebuiltSearchListAGT } from '../../commonSearchesAGT';
import { useFeatureFlag, useSavedQueries } from '../../hooks';
import { QueryLineItem } from '../../types';

export const usePrebuiltQueries = () => {
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');
    const userQueries = useSavedQueries();

    //Get master list of queries to validate against
    const savedLineItems: QueryLineItem[] =
        userQueries.data?.map((query) => ({
            description: query.name,
            cypher: query.query,
            canEdit: true,
            id: query.id,
        })) || [];

    const savedQueries = {
        category: 'Saved Queries',
        subheader: '',
        queries: savedLineItems,
    };
    const queryList = tierFlag?.enabled
        ? [...prebuiltSearchListAGT, savedQueries]
        : [...prebuiltSearchListAGI, savedQueries];

    return queryList;
};
