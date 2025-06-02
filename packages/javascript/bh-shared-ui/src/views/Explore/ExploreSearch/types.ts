import { SearchResult } from '../../../hooks/useSearch';
import { EntityKinds } from '../../../utils/content';

export interface SearchNodeType {
    objectid: string;
    type?: EntityKinds;
    name?: string;
}

//The search value usually aligns with the results from hitting the search endpoint but when
//we are pulling the data from a different page and filling out the value ourselves it might
//not conform to our expected type
export type SearchValue = SearchNodeType | SearchResult;
