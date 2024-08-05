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

import { useFetchGraphItem, GraphItemQueryCacheId, GraphItemData, GraphItemProperties } from './useFetchGraphItem';
import { renderHook, waitFor, queryClientProvider } from '../../test-utils';
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const graphItemApiCall = (apiItemPath: string, expectedPropsResponse: GraphItemProperties) => {
    return rest.get(`/api/v2/${apiItemPath}/:id`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    props: expectedPropsResponse,
                },
            })
        );
    });
};

describe('useFetchGraphItem', () => {
    const server = setupServer(
        graphItemApiCall('computers', {
            haslaps: true,
            objectid: 'testing-id-3456',
        })
    );

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('Search for Node without databaseId returns Node properties', async () => {
        const hookInitialProps = {
            cacheId: GraphItemQueryCacheId.Node,
            objectId: 'testing-id-3456',
            nodeType: 'Computer',
        };

        const { result } = renderHook((nodeItemParams: GraphItemData) => useFetchGraphItem(nodeItemParams), {
            wrapper: queryClientProvider(),
            initialProps: hookInitialProps,
        });

        await waitFor(() => {
            expect(result.current.isSuccess).toBe(true);
        });

        expect(result.current.graphItemProperties).toEqual({
            haslaps: true,
            objectid: 'testing-id-3456',
        });
    });
    it.todo('Search for Node with databaseId returns Node properties', () => {});
    it.todo('Search for Edge returns Edge properties', () => {});
});
