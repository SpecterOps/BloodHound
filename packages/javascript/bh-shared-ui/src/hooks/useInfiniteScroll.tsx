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
