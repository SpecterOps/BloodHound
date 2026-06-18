// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { isAxiosError } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';
import { useTimeoutLimitConfiguration } from '../useConfiguration';
import { useGraphNodeKinds } from '../useGraphKinds';

export type SearchResult = {
    distinguishedname?: string;
    name: string;
    objectid: string;
    system_tags?: string;
    type: string;
};

export type SearchResults = SearchResult[];

export const searchKeys = {
    all: ['search'] as const,
    detail: (keyword: string, type: string | undefined) => [...searchKeys.all, keyword, type] as const,
};

export const useSearch = (keyword = '', type: string | undefined) => {
    const timeoutLimitEnabled = useTimeoutLimitConfiguration();
    const timeout = timeoutLimitEnabled ? 60000 : 0;

    return useQuery<SearchResults, any>({
        queryKey: searchKeys.detail(keyword, type),
        queryFn: ({ signal }) => {
            if (keyword === '') return [];
            return apiClient.searchHandler(keyword, type, { signal, timeout }).then((result) => {
                if (!result.data.data) return [];
                return result.data.data;
            });
        },
        keepPreviousData: true,
        retry: false,
    });
};

export const useKeywordAndTypeValues = (
    inputValue: string | undefined
): { keyword: string | undefined; type: string | undefined } => {
    const { data } = useGraphNodeKinds();

    let keyword: string | undefined = inputValue;
    let type: string | undefined = undefined;

    if (inputValue && inputValue.length > 1) {
        // We use the : as q search/qualifier operator for searches so we need to perform the below to parse it if necessary
        // Only do it if more than one char and that way if a user wants to use the : as a search string it works
        const splitValue = inputValue.split(':');
        if (splitValue.length > 1) {
            const nodeKind = data?.kinds.find(
                (kind) => kind.toLocaleLowerCase() === splitValue[0].toLocaleLowerCase() //abstract
            );
            if (nodeKind) {
                type = nodeKind;
                keyword = splitValue.slice(1).join(':');
            }
        }
    }

    return { keyword, type };
};

const getErrorText = (error: any, type: string | undefined): string => {
    let errorMessage = 'An error has occurred. Please try again.';

    if (error.response?.status === 504) errorMessage = 'Search has timed out. Please try again.';

    if (isAxiosError(error)) {
        const errors = error.response?.data?.errors;
        if (errors?.length) errorMessage = errors[0].message;
        if (errorMessage === 'Invalid type parameter' && type !== undefined)
            errorMessage = `Invalid node kind: ${type}`;
    }

    return errorMessage;
};

const getNoDataText = (debouncedInputValue: string, type: string | undefined, keyword: string | undefined): string => {
    if (debouncedInputValue === '' && type === undefined)
        return 'Begin typing to search. Prepend a type followed by a colon to search by type, e.g., user:bob';
    else if (debouncedInputValue === '' && type !== undefined)
        return `Begin typing to search.`; //For OU and Domain searches which have hard coded type
    else if (type !== undefined && keyword === '') return `Include a keyword to search for a node of type ${type}`;
    else if (type !== undefined && keyword !== '') return `No results found for "${keyword}" of type ${type}`;
    else return `No results found for "${keyword}"`;
};

export const getEmptyResultsText = (
    isLoading: boolean,
    isFetching: boolean,
    isError: boolean,
    error: any,
    debouncedInputValue: string,
    type: string | undefined,
    keyword: string | undefined,
    data: SearchResults | undefined
): string => {
    if (isLoading || isFetching) {
        return 'Loading...';
    } else if (isError) {
        return getErrorText(error, type);
    } else if (data?.length === 0) {
        return getNoDataText(debouncedInputValue, type, keyword);
    } else return '';
};
