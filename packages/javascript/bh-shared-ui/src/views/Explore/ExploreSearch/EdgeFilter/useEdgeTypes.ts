import { EdgeType } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from '../../../../utils';
import { AllEdgeTypes, Category, Subcategory } from './edgeTypes';

const AD_SCHEMA_TYPE = 'ad';
const AZ_SCHEMA_TYPE = 'az';
const BUILT_IN_TYPES = [AD_SCHEMA_TYPE, AZ_SCHEMA_TYPE];

export const useEdgeTypes = () => {
    const edgeTypesQuery = useQuery({
        queryKey: 'getEdgeTypes',
        queryFn: ({ signal }) => apiClient.getEdgeTypes({ signal }).then((res) => res.data.data),
    });

    const customEdgeTypes = filterUneededTypes(edgeTypesQuery.data);

    const combinedEdgeTypes = customEdgeTypes
        ? [...AllEdgeTypes, mapEdgeTypesToCategory(customEdgeTypes, 'OpenGraph')]
        : AllEdgeTypes;

    return {
        isLoading: edgeTypesQuery.isLoading,
        isError: edgeTypesQuery.isError,
        edgeTypes: combinedEdgeTypes,
    };
};

const filterUneededTypes = (data: EdgeType[] | undefined): EdgeType[] | undefined => {
    return data?.filter((edge) => !BUILT_IN_TYPES.includes(edge.schema_name) && edge.is_traversable);
};

const mapEdgeTypesToCategory = (edgeTypes: EdgeType[], categoryName: string): Category => {
    const subcategories = new Map<string, Subcategory>();

    for (const edge of edgeTypes) {
        const existing = subcategories.get(edge.schema_name);

        if (existing) {
            existing.edgeTypes.push(edge.name);
        } else {
            subcategories.set(edge.schema_name, {
                name: edge.schema_name,
                edgeTypes: [edge.name],
            });
        }
    }

    return {
        categoryName,
        subcategories: Array.from(subcategories.values()),
    };
};
