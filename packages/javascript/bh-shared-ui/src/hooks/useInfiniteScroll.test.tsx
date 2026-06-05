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

import { act, renderHook, waitFor } from '@testing-library/react';
import { useInfiniteScroll } from './useInfiniteScroll';

const makeScrollElement = ({
    scrollHeight = 1000,
    scrollTop = 850,
    clientHeight = 100,
}: {
    scrollHeight?: number;
    scrollTop?: number;
    clientHeight?: number;
} = {}) => {
    const element = document.createElement('div');

    Object.defineProperties(element, {
        scrollHeight: {
            configurable: true,
            value: scrollHeight,
        },
        scrollTop: {
            configurable: true,
            value: scrollTop,
        },
        clientHeight: {
            configurable: true,
            value: clientHeight,
        },
    });

    return element;
};

const callOnScroll = (onScroll: ReturnType<typeof useInfiniteScroll>['onScroll'], element: HTMLDivElement) => {
    act(() => {
        onScroll({ currentTarget: element } as Parameters<typeof onScroll>[0]);
    });
};

const setup = (options: Partial<Parameters<typeof useInfiniteScroll>[0]> = {}) => {
    const fetchMore = vi.fn().mockResolvedValue(undefined);
    const renderResult = renderHook(() =>
        useInfiniteScroll({
            canFetchMore: true,
            isFetching: false,
            fetchMore,
            loadedCount: 0,
            ...options,
        })
    );

    return { fetchMore, ...renderResult };
};

describe('useInfiniteScroll', () => {
    it('fetches more data when scrolling within the default bottom threshold', () => {
        const { result, fetchMore } = setup();

        callOnScroll(result.current.onScroll, makeScrollElement());
        expect(fetchMore).toHaveBeenCalledTimes(1);
    });

    it('does not fetch more data when scrolling outside the default bottom threshold', () => {
        const { result, fetchMore } = setup();

        callOnScroll(result.current.onScroll, makeScrollElement({ scrollTop: 799 }));
        expect(fetchMore).not.toHaveBeenCalled();
    });

    it('does not fetch more data when there are no more pages', () => {
        const { result, fetchMore } = setup({ canFetchMore: false });

        callOnScroll(result.current.onScroll, makeScrollElement());
        expect(fetchMore).not.toHaveBeenCalled();
    });

    it('does not fetch more data while a page is already fetching', () => {
        const { result, fetchMore } = setup({ isFetching: true });

        callOnScroll(result.current.onScroll, makeScrollElement());
        expect(fetchMore).not.toHaveBeenCalled();
    });

    it('uses the provided threshold when determining whether the scroll is close enough to the bottom', () => {
        const { result, fetchMore } = setup({ threshold: 200 });

        callOnScroll(result.current.onScroll, makeScrollElement({ scrollTop: 750 }));
        expect(fetchMore).toHaveBeenCalledTimes(1);
    });

    it('checks whether more data is needed when loadedCount changes', async () => {
        const fetchMore = vi.fn().mockResolvedValue(undefined);
        const { result, rerender } = renderHook(
            ({ loadedCount }) =>
                useInfiniteScroll({
                    canFetchMore: true,
                    isFetching: false,
                    fetchMore,
                    loadedCount,
                }),
            { initialProps: { loadedCount: 0 } }
        );

        (result.current.scrollRef as { current: HTMLDivElement | null }).current = makeScrollElement();

        rerender({ loadedCount: 25 });

        await waitFor(() => expect(fetchMore).toHaveBeenCalledTimes(1));
    });
});
