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

import { useQuery } from 'react-query';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { EntityKinds, apiClient } from '../../utils';
import { useTimeoutLimitConfiguration } from '../useConfiguration';

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
    detail: (keyword: string, type: EntityKinds | undefined) => [...searchKeys.all, keyword, type] as const,
};

export const useSearch = (keyword: string, type: EntityKinds | undefined) => {
    const timeoutLimitEnabled = useTimeoutLimitConfiguration();
    const applyTimeoutLimit: number = timeoutLimitEnabled ? 6000 : 0;

    return useQuery<SearchResults, any>(
        searchKeys.detail(keyword, type),
        ({ signal }) => {
            if (keyword === '') return [];
            return apiClient.searchHandler(keyword, type, { signal, timeout: applyTimeoutLimit }).then((result) => {
                if (!result.data.data) return [];
                return result.data.data;
            });
        },
        {
            keepPreviousData: true,
            retry: false,
        }
    );
};

export const getKeywordAndTypeValues = (inputValue = ''): { keyword: string; type: EntityKinds | undefined } => {
    const splitValue = inputValue.split(':');

    let keyword = '';
    let type: EntityKinds | undefined = undefined;

    if (splitValue.length > 1) {
        type = validateNodeType(splitValue[0]);
        keyword = splitValue.slice(1).join(':');
    } else keyword = splitValue[0];

    return { keyword: keyword, type: type };
};

export const validateNodeType = (type: string): EntityKinds | undefined => {
    let result = undefined;

    if (type?.toLowerCase() === 'meta') result = 'Meta' as EntityKinds;

    Object.values(ActiveDirectoryNodeKind).forEach((activeDirectoryType) => {
        if (activeDirectoryType.localeCompare(type, undefined, { sensitivity: 'base' }) === 0)
            result = activeDirectoryType;
    });

    Object.values(AzureNodeKind).forEach((azureType) => {
        if (azureType.localeCompare(type, undefined, { sensitivity: 'base' }) === 0) result = azureType;
    });

    return result;
};

const getErrorText = (error: any): string => {
    if (error.response?.status === 504) return 'Search has timed out. Please try again.';
    else return 'An error has occurred. Please try again.';
};

const getNoDataText = (debouncedInputValue: string, type: EntityKinds | undefined, keyword: string): string => {
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
    type: EntityKinds | undefined,
    keyword: string,
    data: SearchResults | undefined
): string => {
    if (isLoading || isFetching) {
        return 'Loading...';
    } else if (isError) {
        return getErrorText(error);
    } else if (data?.length === 0) {
        return getNoDataText(debouncedInputValue, type, keyword);
    } else return '';
};
