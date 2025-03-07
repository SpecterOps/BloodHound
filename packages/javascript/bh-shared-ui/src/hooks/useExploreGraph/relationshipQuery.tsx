import { SearchResults } from '../../hooks/useSearch';
import { apiClient } from '../../utils/api';
import { EntityInfoDataTableProps, EntityKinds, EntitySectionEndpointParams, allSections } from '../../utils/content';
import { ExploreQueryParams } from '../useExploreParams';
import { ExploreGraphQueryKey, ExploreGraphQueryOptions } from './search-modes/utils';

type SectionEndpoint = (p: EntitySectionEndpointParams) => Promise<any>;

const getEndpoint: (p: any, k: any) => SectionEndpoint | undefined = (
    sectionList: EntityInfoDataTableProps[] | undefined,
    expandedRelationships: Record<string, string>
) => {
    const section = sectionList?.find((section) => expandedRelationships[section.label]);
    if (section?.endpoint) return section.endpoint;
    if (section?.sections) return getEndpoint(section.sections, expandedRelationships);
};

const getRelationshipEndpoint = (
    nodeType: EntityKinds,
    nodeId: string,
    expandedRelationships: Record<string, string>
) => {
    const entitySections = allSections[nodeType]?.(nodeId);

    return getEndpoint(entitySections, expandedRelationships);
};

export const relationshipSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { expandedRelationships, panelSelection, searchType } = paramOptions;

    const isEdgeId = panelSelection?.includes('_'); // TODO: reuse whatever Francisco uses to determine node/edge

    if (searchType !== 'relationship' || !expandedRelationships?.length || !panelSelection || isEdgeId) {
        return {
            enabled: false,
        };
    }

    const accordionMap = expandedRelationships.reduce((a: Record<string, string>, c) => {
        a[c] = c;
        return a;
    }, {});

    return {
        queryKey: [ExploreGraphQueryKey, ...expandedRelationships, panelSelection, searchType],
        queryFn: async ({ signal }) => {
            // gets node info -> get endpoing based on relationships -> return that promise
            const nodeDetails: SearchResults | undefined = await apiClient
                .searchHandler(panelSelection, undefined, { signal })
                .then((result) => {
                    if (!result.data.data) return [];
                    return result.data.data;
                });

            const nodeType = nodeDetails?.[0].type;
            if (!nodeType) {
                throw new Error('unable to fetch relationship');
            }
            const isValidNodeType = nodeType in allSections;

            if (!isValidNodeType) {
                throw new Error('invalid node type');
            }

            const endpoint = getRelationshipEndpoint(nodeType as EntityKinds, panelSelection, accordionMap);

            if (!endpoint) {
                throw new Error('unable to fetch relationship');
            }

            return endpoint({ type: 'graph' });
        },
        refetchOnWindowFocus: false,
    };
};

// ?panelSelection=S-1-5-21-3702535222-3822678775-2090119576-1143&expandedRelationships=Inbound%20Object%20Control&expandedRelationships=Member%20Of&searchType=relationship

/**
 * TODO:
 * try to break
 *   incorrect section ran -- wrong sequence of panels -- depends on how the accordions are going to work.
 *   if we support multiple accordions (not just nested), then we need to reverse the array and cant do the map
 * how can we test
 * can we somehow get some optimization back here with react-query
 * how should we handle nodes that would exceed max nodes rendered? -- accordion side - dont push that
 */
