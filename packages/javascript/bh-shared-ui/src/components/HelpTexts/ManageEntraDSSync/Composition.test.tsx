// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// SPDX-License-Identifier: Apache-2.0

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useExploreGraph } from '../../../hooks/useExploreGraph/useExploreGraph';
import { render, screen, waitFor } from '../../../test-utils';
import Composition from './Composition';

const server = setupServer(
    rest.get('/api/v2/relationships/:relationshipId', (req, res, ctx) =>
        res(
            ctx.json({
                data: {
                    relationship_id: Number(req.params.relationshipId),
                    kind: { relationship_kind_id: 1, name: 'ManageEntraDSSync' },
                    source_node_id: 1,
                    target_node_id: 2,
                    properties: {},
                },
            })
        )
    ),
    rest.get('/api/v2/graphs/edge-composition', (_req, res, ctx) => res(ctx.json({ data: { nodes: {}, edges: [] } }))),
    rest.get('/api/v2/config', (_req, res, ctx) => res(ctx.json({ data: [] })))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const CompositionGraphProbe = () => {
    const { data } = useExploreGraph();
    return data ? <span>graph-loaded</span> : null;
};

describe('ManageEntraDSSync Composition', () => {
    it('preserves the selected relationship ID used by the composition graph query', async () => {
        render(
            <>
                <Composition sourceDBId={1} targetDBId={2} edgeName='ManageEntraDSSync' />
                <CompositionGraphProbe />
            </>,
            { route: '/?searchType=composition&relationshipQueryItemId=rel_99' }
        );

        expect(await screen.findByText('graph-loaded')).toBeInTheDocument();
        await waitFor(() => {
            expect(window.location.search).toContain('relationshipQueryItemId=rel_99');
        });
        expect(window.location.search).not.toContain('relationshipQueryItemId=rel_1_ManageEntraDSSync_2');
    });
});
