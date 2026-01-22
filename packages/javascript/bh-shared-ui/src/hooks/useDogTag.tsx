// Copyright 2024 Specter Ops, Inc.
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

import type { AxiosError } from 'axios';
import { RequestOptions } from 'js-client-library';
import { UseQueryOptions, useQuery } from 'react-query';
import { DogTagValue } from '../components/DogTag/DogTag';
import { apiClient } from '../utils';

export type DogTag<T extends DogTagValue> = {
    key: string;
    value: T;
};

export type DogTagResponse = Record<string, DogTagValue>;

export const dogTagKeys = {
    all: ['dogTags'],
    getKey: (customKey?: readonly unknown[]) =>
        customKey?.length ? [...dogTagKeys.all, ...customKey] : dogTagKeys.all,
};

export const getDogTags = async (options?: RequestOptions): Promise<DogTagResponse> => {
    const response = await apiClient.getDogTags(options);
    return response.data.data;
};

type QueryOptions<T> = Omit<UseQueryOptions<DogTagResponse, AxiosError, T, readonly unknown[]>, 'queryFn'>;

export function useDogTags<T = DogTagResponse>(options?: QueryOptions<T>) {
    const { queryKey, ...rest } = options ?? {};

    return useQuery<DogTagResponse, AxiosError, T, readonly unknown[]>({
        ...rest,
        queryKey: dogTagKeys.getKey(queryKey),
        queryFn: ({ signal }) => getDogTags({ signal }),
    });
}

export function useDogTag<T extends DogTagValue>(
    dogTagKey: string,
    options?: Omit<QueryOptions<T | undefined>, 'select'>
) {
    return useDogTags<T | undefined>({
        ...options,
        select: (data) => data[dogTagKey] as T,
    });
}
