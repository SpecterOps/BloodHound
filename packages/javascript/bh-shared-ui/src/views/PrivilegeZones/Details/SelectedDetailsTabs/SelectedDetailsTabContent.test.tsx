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
import { useAutomaticGraphActions, useExploreGraph, useExploreSelectedItem } from '../../../../hooks';
import { mockSourceKindsHandler, zoneHandlers } from '../../../../mocks';
import * as zoneMocks from '../../../../mocks/factories/privilegeZones';
import { render, renderHook, screen, waitFor } from '../../../../test-utils';
import { ObjectTabValue, RuleTabValue, TagTabValue } from '../../utils';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';

const server = setupServer(
    rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (req, res, ctx) => {
        const { tagId, selectorId } = req.params;
        return res(
            ctx.json({
                data: {
                    selector: zoneMocks.createRuleWithCypher(parseInt(tagId as string), parseInt(selectorId as string)),
                },
            })
        );
    }),
    ...zoneHandlers,
    mockSourceKindsHandler()
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Selected Details Tab Content', () => {
    it('renders the Zone/Labels Tab content when first tab is chosen', async () => {
        render(<SelectedDetailsTabContent currentDetailsTab={TagTabValue} tagId='1' />);

        const zoneTitle = await screen.findByText(/tag-0/i); // can find the structure of title in mocks/factories/privilegeZones

        expect(zoneTitle).toBeInTheDocument();
    });
    it('renders the Rule Tab content when Rule tab is chosen', async () => {
        render(<SelectedDetailsTabContent currentDetailsTab={RuleTabValue} tagId='1' ruleId='2' />);

        const ruleTitle = await screen.findByText(/tag-0-rule-2/i); // can find the structure of title in mocks/factories/privilegeZones

        expect(ruleTitle).toBeInTheDocument();
    });
    it('renders the Cypher Rules Panel when clicking the Rule Tab', async () => {
        render(<SelectedDetailsTabContent currentDetailsTab={RuleTabValue} tagId='2' ruleId='2' />);

        const cypherTitle = await screen.findByText(/cypher preview/i);
        const cypherContainer = await screen.findByTestId('privilege-zones_cypher-container');

        expect(cypherTitle).toBeInTheDocument();
        expect(cypherContainer).toBeInTheDocument();
    });
    it('renders the Object Tab content when Object tab is chosen', async () => {
        server.use(
            rest.get('/api/v2/nodes/1', (_req, res, ctx) =>
                res(ctx.json({ data: { node_id: 1, kinds: [{ name: 'User', node_kind_id: 1 }], properties: {} } }))
            )
        );

        render(<SelectedDetailsTabContent currentDetailsTab={ObjectTabValue} tagId='1' ruleId='2' memberId='1' />);

        const entityInfoPanel = await screen.findByTestId('explore_entity-information-panel');

        expect(entityInfoPanel).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /open in explore/i })).toBeVisible();
    });
    it.each([
        { nodeId: 123, objectId: 'CUSTOM/OBJECT #123' },
        { nodeId: 0, objectId: 'CUSTOM/OBJECT #0' },
        { nodeId: 123, objectId: '' },
    ])('opens graph node $nodeId with object ID "$objectId" in Explore', async ({ nodeId, objectId }) => {
        const properties = { name: 'Shared name', objectid: objectId, customprop: 'OpenGraph Value' };
        const namesakeSearch = vi.fn();

        server.use(
            rest.get(`/api/v2/asset-group-tags/1/members/${nodeId}`, (_req, res, ctx) =>
                res(ctx.json({ data: { member: { id: nodeId, object_id: objectId, name: properties.name } } }))
            ),
            rest.get(`/api/v2/nodes/${nodeId}`, (_req, res, ctx) =>
                res(ctx.json({ data: { node_id: nodeId, kinds: [{ name: 'Custom', node_kind_id: 1 }], properties } }))
            ),
            rest.get('/api/v2/graph-search', (_req, res, ctx) => {
                namesakeSearch();
                return res(ctx.json({ data: { 456: { data: { name: objectId, objectid: 'DIFFERENT-ID' } } } }));
            }),
            rest.post('/api/v2/graphs/cypher', async (req, res, ctx) => {
                const { query } = await req.json<{ query: string }>();
                if (query !== `MATCH (n) WHERE id(n) = ${nodeId} RETURN n`) return res(ctx.status(400));

                return res(ctx.json({ data: { nodes: { [nodeId]: { kinds: ['Custom'], properties } }, edges: [] } }));
            })
        );

        const { unmount } = render(
            <SelectedDetailsTabContent currentDetailsTab={ObjectTabValue} tagId='1' memberId={String(nodeId)} />
        );

        expect(await screen.findByText('OpenGraph Value')).toBeInTheDocument();
        const link = screen.getByRole('link', { name: /open in explore/i });
        const url = new URL((link as HTMLAnchorElement).href);

        expect(url.pathname).toBe('/explore');
        expect(link).toHaveAttribute('target', '_blank');
        unmount();

        const { result } = renderHook(
            () => {
                const graphQuery = useExploreGraph();
                useAutomaticGraphActions(graphQuery.data);
                return { graphQuery, ...useExploreSelectedItem() };
            },
            { route: url.pathname + url.search }
        );

        await waitFor(() => expect(result.current.graphQuery.isSuccess).toBe(true));
        expect(result.current.graphQuery.data).toEqual({
            data: { nodes: { [nodeId]: { kinds: ['Custom'], properties } }, edges: [] },
        });
        expect(result.current.selectedItem).toBe(String(nodeId));
        await waitFor(() =>
            expect(result.current.selectedItemQuery.data).toMatchObject({ node_id: nodeId, properties })
        );
        expect(namesakeSearch).not.toHaveBeenCalled();
    });
});
