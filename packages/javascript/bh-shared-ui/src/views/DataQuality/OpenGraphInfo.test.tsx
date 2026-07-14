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
import { render, screen, waitFor } from '../../test-utils';
import OpenGraphInfo, { OpenGraphPlatformInfo, getLatestMetricStats } from './OpenGraphInfo';

const nodeStat = (overrides: Record<string, unknown> = {}) => ({
    metric_kind_id: 1,
    metric_type: 'node',
    metric_name: 'User',
    metric_value: 10,
    created_at: '2026-01-01T00:00:00.000Z',
    ...overrides,
});

const relationshipStat = (overrides: Record<string, unknown> = {}) => ({
    metric_kind_id: 100,
    metric_type: 'relationship',
    metric_name: 'HasRelationship',
    metric_value: 50,
    created_at: '2026-01-01T00:00:00.000Z',
    ...overrides,
});

describe('getLatestMetricStats', () => {
    it('returns empty node stats and undefined relationship stats for empty input', () => {
        const { nodeStats, relationshipStats } = getLatestMetricStats([]);

        expect(nodeStats).toEqual([]);
        expect(relationshipStats).toBeUndefined();
    });

    it('returns a single node stat and no relationship stat', () => {
        const stat = nodeStat();

        const { nodeStats, relationshipStats } = getLatestMetricStats([stat]);

        expect(nodeStats).toEqual([stat]);
        expect(relationshipStats).toBeUndefined();
    });

    it('keeps only the latest stat per metric_kind_id by created_at', () => {
        const older = nodeStat({ created_at: '2026-01-01T00:00:00.000Z', metric_value: 10 });
        const newer = nodeStat({ created_at: '2026-01-02T00:00:00.000Z', metric_value: 25 });

        const { nodeStats } = getLatestMetricStats([older, newer]);

        expect(nodeStats).toHaveLength(1);
        expect(nodeStats[0]).toBe(newer);
    });

    it('keeps the latest regardless of input ordering', () => {
        const older = nodeStat({ created_at: '2026-01-01T00:00:00.000Z', metric_value: 10 });
        const newer = nodeStat({ created_at: '2026-01-02T00:00:00.000Z', metric_value: 25 });

        const { nodeStats } = getLatestMetricStats([newer, older]);

        expect(nodeStats).toHaveLength(1);
        expect(nodeStats[0]).toBe(newer);
    });

    it('splits mixed node and relationship stats', () => {
        const node1 = nodeStat({ metric_kind_id: 1, metric_name: 'User' });
        const node2 = nodeStat({ metric_kind_id: 2, metric_name: 'Group' });
        const relationship = relationshipStat({ metric_kind_id: 100 });

        const { nodeStats, relationshipStats } = getLatestMetricStats([node1, node2, relationship]);

        expect(nodeStats).toEqual([node1, node2]);
        expect(relationshipStats).toBe(relationship);
    });

    it('returns undefined relationship stats when only node stats are present', () => {
        const node1 = nodeStat({ metric_kind_id: 1 });
        const node2 = nodeStat({ metric_kind_id: 2 });

        const { nodeStats, relationshipStats } = getLatestMetricStats([node1, node2]);

        expect(nodeStats).toEqual([node1, node2]);
        expect(relationshipStats).toBeUndefined();
    });

    it('keeps the first-seen stat when created_at values are missing/invalid', () => {
        const first = nodeStat({ created_at: undefined, metric_value: 10 });
        const second = nodeStat({ created_at: undefined, metric_value: 25 });

        const { nodeStats } = getLatestMetricStats([first, second]);

        expect(nodeStats).toHaveLength(1);
        expect(nodeStats[0]).toBe(first);
    });
});

const STATS_URL = '/api/v2/data-quality-stats';
const AGGREGATION_URL = '/api/v2/data-quality-stats-aggregation';

const ogStat = (overrides: Record<string, unknown> = {}) => ({
    id: 1,
    metric_kind_id: 1,
    environment_kind_id: 999,
    metric_type: 'node',
    metric_name: 'User',
    metric_value: 10,
    created_at: '2026-01-01T00:00:00.000Z',
    ...overrides,
});

const server = setupServer(rest.get('/api/v2/custom-nodes', (_req, res, ctx) => res(ctx.json({ data: {} }))));

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const respondWith = (url: string, stats: unknown[], status = 200) =>
    server.use(rest.get(url, (_req, res, ctx) => res(ctx.status(status), ctx.json({ data: stats }))));

describe('OpenGraphInfo', () => {
    it('renders loading skeletons while the query is pending', () => {
        respondWith(STATS_URL, [ogStat()]);

        const { container } = render(<OpenGraphInfo contextId='env-1' />);

        expect(container.querySelectorAll('.MuiSkeleton-root').length).toBeGreaterThan(0);
    });

    it('calls onDataError and renders nothing when the request fails', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        const onDataError = vi.fn();
        respondWith(STATS_URL, [], 500);

        const { container } = render(<OpenGraphInfo contextId='env-1' onDataError={onDataError} />);

        await waitFor(() => expect(onDataError).toHaveBeenCalled());
        expect(container.querySelectorAll('.MuiSkeleton-root')).toHaveLength(0);
        expect(screen.queryByText('Relationships')).not.toBeInTheDocument();
        consoleError.mockRestore();
    });

    it('renders nothing when the response contains no stats', async () => {
        respondWith(STATS_URL, []);

        const { container } = render(<OpenGraphInfo contextId='env-1' />);

        await waitFor(() => expect(container.querySelectorAll('.MuiSkeleton-root')).toHaveLength(0));
        expect(screen.queryByText('Relationships')).not.toBeInTheDocument();
    });

    it('renders a row per node stat plus the relationship row when populated', async () => {
        respondWith(STATS_URL, [
            ogStat({ id: 1, metric_kind_id: 1, metric_name: 'User', metric_value: 10 }),
            ogStat({ id: 2, metric_kind_id: 2, metric_name: 'Group', metric_value: 20 }),
            ogStat({ id: 3, metric_kind_id: 100, metric_type: 'relationship', metric_value: 50 }),
        ]);

        render(<OpenGraphInfo contextId='env-1' />);

        expect(await screen.findByText('User')).toBeInTheDocument();
        expect(screen.getByText('Group')).toBeInTheDocument();
        expect(screen.getByText('Relationships')).toBeInTheDocument();
        expect(screen.getByText('50')).toBeInTheDocument();
    });

    it('skips node rows whose metric_kind_id matches the environment_kind_id', async () => {
        respondWith(STATS_URL, [
            ogStat({ id: 1, metric_kind_id: 1, environment_kind_id: 999, metric_name: 'VisibleType' }),
            ogStat({ id: 2, metric_kind_id: 5, environment_kind_id: 5, metric_name: 'SkippedType' }),
        ]);

        render(<OpenGraphInfo contextId='env-1' />);

        expect(await screen.findByText('VisibleType')).toBeInTheDocument();
        expect(screen.queryByText('SkippedType')).not.toBeInTheDocument();
    });

    it('falls back to 0 for the relationship row when no relationship stat is present', async () => {
        respondWith(STATS_URL, [ogStat({ id: 1, metric_kind_id: 1, metric_name: 'User', metric_value: 10 })]);

        render(<OpenGraphInfo contextId='env-1' />);

        expect(await screen.findByText('Relationships')).toBeInTheDocument();
        expect(screen.getByText('0')).toBeInTheDocument();
    });
});

describe('OpenGraphPlatformInfo', () => {
    it('renders node and relationship rows from the aggregation endpoint', async () => {
        respondWith(AGGREGATION_URL, [
            ogStat({ id: 1, metric_kind_id: 1, metric_name: 'User', metric_value: 10 }),
            ogStat({ id: 2, metric_kind_id: 100, metric_type: 'relationship', metric_value: 50 }),
        ]);

        render(<OpenGraphPlatformInfo contextKindId={101} />);

        expect(await screen.findByText('User')).toBeInTheDocument();
        expect(screen.getByText('Relationships')).toBeInTheDocument();
        expect(screen.getByText('50')).toBeInTheDocument();
    });

    it('calls onDataError and renders nothing when the aggregation request fails', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        const onDataError = vi.fn();
        respondWith(AGGREGATION_URL, [], 500);

        const { container } = render(<OpenGraphPlatformInfo contextKindId={101} onDataError={onDataError} />);

        await waitFor(() => expect(onDataError).toHaveBeenCalled());
        expect(container.querySelectorAll('.MuiSkeleton-root')).toHaveLength(0);
        expect(screen.queryByText('Relationships')).not.toBeInTheDocument();
        consoleError.mockRestore();
    });
});

