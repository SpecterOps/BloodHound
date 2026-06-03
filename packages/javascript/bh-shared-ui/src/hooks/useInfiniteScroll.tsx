// Copyright 2026 Specter Ops, Inc.
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
import { UIEvent, useCallback, useEffect, useRef } from 'react';

type InfiniteScrollOptions = {
    canFetchMore: boolean;
    isFetching: boolean;
    fetchMore: () => Promise<unknown>;
    loadedCount: number;
    threshold?: number;
};

export const useInfiniteScroll = ({
    canFetchMore,
    isFetching,
    fetchMore,
    loadedCount,
    threshold = 100,
}: InfiniteScrollOptions) => {
    const scrollRef = useRef<HTMLDivElement>(null);

    const fetchMoreDataOnBottomReached = useCallback(
        (e?: HTMLDivElement | null) => {
            if (!e || isFetching || !canFetchMore) return;

            if (e.scrollHeight - e.scrollTop - e.clientHeight < threshold) {
                fetchMore();
            }
        },
        [fetchMore, isFetching, canFetchMore, threshold]
    );

    // Checks if we need to fetch more rows immediately after render instead of waiting for onScroll to fire -- handles the case
    // where the first page of results does not fill the table height.
    useEffect(() => {
        fetchMoreDataOnBottomReached(scrollRef.current);
    }, [fetchMoreDataOnBottomReached, loadedCount]);

    return {
        scrollRef,
        onScroll: (e: UIEvent<HTMLDivElement>) => {
            fetchMoreDataOnBottomReached(e.currentTarget);
        },
    };
};
