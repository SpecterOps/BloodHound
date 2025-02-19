import { UseQueryOptions } from 'react-query';

type QueryKeys = ('explore-graph-query' | string | undefined)[];

export type ExploreGraphQueryOptions = UseQueryOptions<unknown, unknown, unknown, QueryKeys>;

export type GraphItemMutationFn = (items: any) => unknown;

export const ExploreGraphQueryKey = 'explore-graph-query';
