import { createFileRoute } from '@tanstack/react-router';
import { entityRelationshipEndpoints, EntityRelationshipQueryTypes, MappedStringLiteral } from 'bh-shared-ui';
import { EdgeCheckboxType } from 'node_modules/bh-shared-ui/dist/views/Explore/ExploreSearch/EdgeFilter/edgeCategories';
import { authenticateToRoute } from './-utils';

export type ExploreSearchTab = 'node' | 'pathfinding' | 'cypher';
type SearchType = ExploreSearchTab | 'relationship' | 'composition' | 'aclinheritance';

export type ExploreQueryParams = {
    exploreSearchTab: ExploreSearchTab | null;
    primarySearch: string | null;
    secondarySearch: string | null;
    cypherSearch: string | null;
    searchType: SearchType | null;
    expandedPanelSections: string[] | null;
    selectedItem: string | null;
    relationshipQueryType: EntityRelationshipQueryTypes | null;
    relationshipQueryItemId: string | null;
    pathFilters: EdgeCheckboxType['edgeType'][] | null;
};

const acceptedExploreSearchTabs = {
    node: 'node',
    pathfinding: 'pathfinding',
    cypher: 'cypher',
} satisfies MappedStringLiteral<ExploreSearchTab, ExploreSearchTab>;

const parseSearchTab = (paramValue: string | null): ExploreSearchTab | null => {
    if (paramValue && paramValue in acceptedExploreSearchTabs) {
        return paramValue as ExploreSearchTab;
    }
    return null;
};

const acceptedSearchTypes = {
    ...acceptedExploreSearchTabs,
    relationship: 'relationship',
    composition: 'composition',
    aclinheritance: 'aclinheritance',
} satisfies MappedStringLiteral<SearchType, SearchType>;

const parseSearchType = (paramValue: string | null): SearchType | null => {
    if (paramValue && paramValue in acceptedSearchTypes) {
        return paramValue as SearchType;
    }
    return null;
};

const parseRelationshipQueryType = (paramValue: string | null): EntityRelationshipQueryTypes | null => {
    if (paramValue && paramValue in entityRelationshipEndpoints) {
        return paramValue as EntityRelationshipQueryTypes;
    }
    return null;
};

export const Route = createFileRoute('/explore')({
    beforeLoad: ({ context }) => authenticateToRoute(context.auth),
    staticData: { showNavbar: true },
    validateSearch: (search: Record<string, unknown>): ExploreQueryParams => {
        return {
            exploreSearchTab: parseSearchTab(search.exploreSearchTab as string),
            primarySearch: search.primarySearch as string,
            secondarySearch: search.secondarySearch as string,
            cypherSearch: search.cypherSearch as string,
            searchType: parseSearchType(search.searchType as string),
            expandedPanelSections: (search.expandedPanelSections as string[]) ?? [],
            selectedItem: search.selectedItem as string,
            relationshipQueryType: parseRelationshipQueryType(search.relationshipQueryType as string),
            relationshipQueryItemId: search.relationshipQueryItemId as string,
            pathFilters: (search.pathFilters as ExploreQueryParams['pathFilters']) ?? [],
        };
    },
});
