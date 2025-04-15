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

import { createMemoryHistory } from 'history';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { VirtualizedNodeListItem } from '../../components/VirtualizedNodeList';
import { renderHook, waitFor } from '../../test-utils';
import { EdgeInfoItems, EdgeInfoItemsProps, useEdgeInfoItems } from './useEdgeInfoItems';

const testDataNodes = {
    19699711: {
        isOwnedObject: false,
        isTierZero: true,
        kind: 'Domain',
        label: 'DUMPSTER.FIRE',
        lastSeen: '2025-03-28T17:44:37.218202364Z',
        objectId: 'S-1-5-21-2697957641-2271029196-387917394',
        properties: {},
    },
    24748936: {
        isOwnedObject: false,
        isTierZero: false,
        kind: 'CertTemplate',
        label: 'SUBCA@DUMPSTER.FIRE',
        lastSeen: '2025-03-28T17:44:37.286745073Z',
        objectId: '6DDBB525-E75C-4DA2-A311-DC8339C24B71',
        properties: {},
    },
    24748974: {
        isOwnedObject: false,
        isTierZero: false,
        kind: 'CertTemplate',
        label: 'ESC1@DUMPSTER.FIRE',
        lastSeen: '2025-03-28T17:44:37.286745073Z',
        objectId: '15AB8124-9C4A-4FCC-9E51-BE991E2CC112',
        properties: {},
    },
};
const testDataEdges = [
    {
        kind: 'Enroll',
        label: 'Enroll',
        lastSeen: '2025-03-28T17:44:37.286745073Z',
        properties: { isacl: true, isinherited: false, lastseen: '2025-03-28T17:44:37.286745073Z' },
        source: '39194477',
        target: '24749153',
    },
    {
        kind: 'PublishedTo',
        label: 'PublishedTo',
        lastSeen: '2025-03-28T17:44:36.980837831Z',
        properties: { isacl: false, lastseen: '2025-03-28T17:44:36.980837831Z' },
        source: '24749153',
        target: '24749357',
    },
];

const testHookParams: EdgeInfoItemsProps = {
    sourceDBId: 24749100,
    targetDBId: 19699711,
    edgeName: 'ADCSESC1',
    type: EdgeInfoItems['composition'],
};

const handleNodeClick = () => {};

const testDataNodeArray: VirtualizedNodeListItem = {
    name: 'DUMPSTER.FIRE',
    objectId: 'S-1-5-21-2697957641-2271029196-387917394',
    graphId: '19699711',
    kind: 'Domain',
    onClick: handleNodeClick,
};

const backButtonSupportFF = {
    key: 'back_button_support',
    enabled: true,
};

const server = setupServer(
    rest.get('/api/v2/graphs/edge-composition', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: { nodes: testDataNodes, edges: testDataEdges },
            })
        );
    }),
    rest.get('/api/v2/graphs/relay-targets', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: { nodes: testDataNodes, edges: testDataEdges },
            })
        );
    }),
    rest.get('/api/v2/asset-groups/1/members/counts', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: 0,
            })
        );
    }),
    rest.get('/api/v2/features', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [backButtonSupportFF],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useEdgeInfoItems', () => {
    it('returns a nodes array', async () => {
        const hook = renderHook(() => useEdgeInfoItems(testHookParams), {});

        await waitFor(async () => {
            expect(hook.result.current.isLoading).toBe(false);
        });

        const nodesArray = hook.result.current.nodesArray;

        expect(nodesArray).toHaveLength(3);
        expect(JSON.stringify(nodesArray[0])).toBe(JSON.stringify(testDataNodeArray));
    });
    it('sets selectedItem,primarySearch, searchType in URL params when the click handler is executed', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useEdgeInfoItems(testHookParams), { history });

        await waitFor(async () => {
            expect(hook.result.current.isLoading).toBe(false);
        });

        hook.result.current.nodesArray[0].onClick(0);

        expect(history.location.search).toContain('selectedItem');
        expect(history.location.search).toContain('primarySearch');
        expect(history.location.search).toContain('searchType');
    });
});
