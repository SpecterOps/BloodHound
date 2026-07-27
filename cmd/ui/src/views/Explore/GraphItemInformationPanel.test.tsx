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
import {
    RACF_USER_GROUPS_SECTION,
    RACF_USER_INBOUND_RELATIONSHIPS_SECTION,
    RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION,
} from 'src/racfhound/groupMembers';
import { render, screen } from 'src/test-utils';
import GraphItemInformationPanel from './GraphItemInformationPanel';

// A RACFUser node as returned by GET /api/v2/nodes/:id. RACF kinds are custom (OpenGraph)
// kinds, so they are NOT "built in" — see the regression note in the test below.
const racfUserNode = {
    node_id: 105,
    kinds: [{ node_kind_id: 1, name: 'RACFUser' }],
    properties: { objectid: 'USER1', name: 'USER1' },
};

const server = setupServer(
    // Feature flags off (keeps tier-management asset-group-tags query disabled)
    rest.get('/api/v2/features', (_request, response, context) => response(context.json({ data: [] }))),
    rest.get('/api/v2/graphs/source-kinds', (_request, response, context) =>
        response(context.json({ data: { kinds: [] } }))
    ),
    rest.get('/api/v2/custom-nodes', (_request, response, context) => response(context.json({ data: [] }))),
    // The selected node
    rest.get('/api/v2/nodes/:id', (_request, response, context) => response(context.json({ data: racfUserNode }))),
    // RACF relationship section counts (empty is fine; we only assert the section headers render)
    rest.post('/api/v2/graphs/cypher', (_request, response, context) =>
        response(context.json({ data: { nodes: {}, edges: [] } }))
    ),
    // Incidental provider/permission requests made while rendering the panel; empty responses
    // keep them from failing noisily. None affect the assertions below.
    rest.get('/api/v2/self', (_request, response, context) => response(context.json({ data: {} }))),
    rest.get('/api/v2/roles', (_request, response, context) => response(context.json({ data: { roles: [] } }))),
    rest.get('/api/v2/asset-group-tags', (_request, response, context) =>
        response(context.json({ data: { tags: [] } }))
    )
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('GraphItemInformationPanel RACF integration', () => {
    // Regression guard for the sections silently disappearing after the upstream merge.
    //
    // EntityInfoContent only renders `additionalTables` for built-in kinds (via EntityInfoList);
    // custom/OpenGraph kinds like RACFUser are routed to KindInfoItems, which ignores them. The
    // wrapper therefore injects the RACF tables as `priorityTables`, which render unconditionally.
    // If that ever regresses back to `additionalTables`, these RACF sections vanish for RACF
    // nodes and this test fails.
    it('renders the RACF relationship sections for a selected RACF node', async () => {
        render(<GraphItemInformationPanel />, { route: '/?selectedItem=105' });

        expect(await screen.findByText(RACF_USER_GROUPS_SECTION)).toBeInTheDocument();
        expect(await screen.findByText(RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION)).toBeInTheDocument();
        expect(await screen.findByText(RACF_USER_INBOUND_RELATIONSHIPS_SECTION)).toBeInTheDocument();
    });
});
