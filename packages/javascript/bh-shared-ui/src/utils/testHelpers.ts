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

import { QueryClient, QueryKey } from 'react-query';

/**
 * Tests interacting with a codemirror editor can output unwanted errors relating to missing DOM methods; running this
 * function in your test file will prevent those errors.
 */
export const mockCodemirrorLayoutMethods = () => {
    const getBoundingClientRect = (): DOMRect => {
        const rec = {
            x: 0,
            y: 0,
            bottom: 0,
            height: 0,
            left: 0,
            right: 0,
            top: 0,
            width: 0,
        };
        return { ...rec, toJSON: () => rec };
    };

    class FakeDOMRectList extends Array<DOMRect> implements DOMRectList {
        item(index: number): DOMRect | null {
            return this[index];
        }
    }

    document.elementFromPoint = (): null => null;
    HTMLElement.prototype.getBoundingClientRect = getBoundingClientRect;
    HTMLElement.prototype.getClientRects = (): DOMRectList => new FakeDOMRectList();
    Range.prototype.getBoundingClientRect = getBoundingClientRect;
    Range.prototype.getClientRects = (): DOMRectList => new FakeDOMRectList();
};

/**
 * @description SetUpQueryClient takes in stateMaps in the form of an array of objects where each object has a "key" key and a "data" key
 *
 * @param  stateMaps These maps are looped over for hydrating the queryClient with the state that is required for the test(s) the queryClient is being used for
 *
 */
export const setUpQueryClient = (stateMaps: { key: QueryKey; data: any }[]) => {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
                refetchOnMount: false,
                refetchOnWindowFocus: false,
                staleTime: Infinity,
            },
        },
    });

    stateMaps.forEach(({ key, data }) => {
        queryClient.setQueryData(key, data);
    });

    return queryClient;
};
