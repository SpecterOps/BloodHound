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

import { getLatestMetricStats } from './OpenGraphInfo';

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
