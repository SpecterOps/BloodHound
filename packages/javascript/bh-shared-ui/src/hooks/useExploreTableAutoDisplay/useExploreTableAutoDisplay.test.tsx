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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../../test-utils';
import { useExploreTableAutoDisplay } from './useExploreTableAutoDisplay';

const graphShapedResponse = { edges: [{ testEdge: {} }], nodes: { testNode: { objectId: '' } } };
const tableShapedResponse = { edges: [], nodes: { testNode: { objectId: '' } } };
const getCypherAPIMock = (results: Record<string, any>) => {
    return rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(
            ctx.json({
                data: results,
            })
        );
    });
};
const server = setupServer(
    getCypherAPIMock(tableShapedResponse),
    rest.get('/api/v2/features', (_, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 1, key: 'explore_table_view', enabled: true }],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useExploreTableAutoDisplay', () => {
    const setup = ({ enabled = true, initialRoute = '' } = {}) => {
        const actual = renderHook(() => useExploreTableAutoDisplay(enabled), { route: initialRoute });
        return actual;
    };

    it('sets autoDisplayTable to false if enabled param is false', () => {
        const actual = setup({ enabled: false });

        expect(actual.result.current[0]).toBe(false);
    });

    it('sets autoDisplayTable to false if query returns edges', async () => {
        server.use(getCypherAPIMock(graphShapedResponse));

        const initialRoute = '?searchType=cypher&cypherSearch=YQ%3D%3D';
        const { result } = setup({ initialRoute });

        await waitFor(() => expect(result.current[0]).toBe(false));
    });

    it('sets autoDisplayTable to false if query returns no results', async () => {
        const initialRoute = '?searchType=cypher&cypherSearch=YQ%3D%3D';
        const { result } = setup({ initialRoute });

        await waitFor(() => expect(result.current[0]).toBe(false));
    });

    it('sets autoDisplayTable to false if query returns nodes only but the enabled prop is false', async () => {
        const initialRoute = '?searchType=cypher&cypherSearch=YQ%3D%3D';
        const { result } = setup({ enabled: false, initialRoute });

        await waitFor(() => expect(result.current[0]).toBe(false));
    });

    it.each(['node', 'pathfinding', 'relationship', 'composition'])(
        'sets autoDisplayTable to false if query type is %s',
        async (searchType) => {
            const initialRoute = `?searchType=${searchType}&cypherSearch=YQ%3D%3D`;
            const { result } = setup({ initialRoute });

            await waitFor(() => expect(result.current[0]).toBe(false));
        }
    );

    it('sets autoDisplayTable to true if query returns nodes only, searchType = cypher, and and enabled is true', async () => {
        const initialRoute = '?searchType=cypher&cypherSearch=YQ%3D%3D';
        const { result } = setup({ enabled: true, initialRoute });

        await waitFor(() => expect(result.current[0]).toBe(true));
    });
});
