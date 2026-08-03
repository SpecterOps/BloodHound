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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useExploreGraph } from '../../../hooks/useExploreGraph/useExploreGraph';
import { render, screen, waitFor } from '../../../test-utils';
import Composition from './Composition';

const server = setupServer(
    rest.get('/api/v2/relationships/:relationshipId', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    relationship_id: Number(req.params.relationshipId),
                    kind: { relationship_kind_id: 1, name: 'AddEntraDSGroupMember' },
                    source_node_id: 1,
                    target_node_id: 2,
                    properties: {},
                },
            })
        );
    }),
    rest.get('/api/v2/graphs/edge-composition', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: { nodes: {}, edges: [] },
            })
        );
    }),
    rest.get('/api/v2/config', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const CompositionGraphProbe = () => {
    const { data } = useExploreGraph();

    return data ? <span>graph-loaded</span> : null;
};

describe('AddEntraDSGroupMember Composition', () => {
    it('uses the tuple-form relationship key required by the composition graph query', async () => {
        render(
            <>
                <Composition sourceDBId={1} targetDBId={2} edgeName='AddEntraDSGroupMember' />
                <CompositionGraphProbe />
            </>,
            {
                route: '/?searchType=composition&relationshipQueryItemId=rel_99',
            }
        );

        await waitFor(() => {
            expect(window.location.search).toContain(
                'relationshipQueryItemId=rel_1_AddEntraDSGroupMember_2'
            );
        });
        expect(await screen.findByText('graph-loaded')).toBeInTheDocument();
    });
});
