// Copyright 2025 Specter Ops, Inc.
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

export type PageParam = {
    skip: number;
    limit: number;
};

export type PaginatedResult<T> = {
    items: T[];
    nextPageParam?: PageParam;
};

export type PaginatedApiResponse<T> = {
    data: {
        data: { [key: string]: T[] };
        count: number;
    };
};

/**
 * Generic helper to wrap paginated API responses into a React Query-friendly shape.
 */
export const createPaginatedFetcher = <T>(
    fetchFn: () => Promise<PaginatedApiResponse<T>>,
    key: string,
    skip: number,
    limit: number
): Promise<PaginatedResult<T>> =>
    fetchFn().then((res) => {
        const items = res.data.data[key];
        const hasMore = skip + limit < res.data.count;
        return {
            items,
            nextPageParam: hasMore ? { skip: skip + limit, limit } : undefined,
        };
    });
