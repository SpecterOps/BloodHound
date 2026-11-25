// Copyright 2023 Specter Ops, Inc.
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

import { getCollapseSummaryId, type CollapsePlan, type CollapseView } from 'bh-shared-ui';
import type { GraphData, GraphEdge, GraphNodes } from 'js-client-library';

type SummaryDetails = { groupId: string; count: number };

export type SigmaExploreEdge = Pick<GraphEdge, 'source' | 'target' | 'kind' | 'label' | 'exploreGraphId'> &
    ({ id: GraphEdge['id'] } | { summaryId: string; collapseGroupId: string });

export type SigmaExploreItems = {
    nodes: GraphNodes;
    edges: SigmaExploreEdge[];
    summaryNodes?: ReadonlyMap<string, SummaryDetails>;
};

export function projectSigmaCollapse(items: GraphData, plan: CollapsePlan, view: CollapseView): SigmaExploreItems {
    if (view.collapsedGroups.size === 0) return items;

    const nodes: GraphNodes = Object.fromEntries(Object.entries(items.nodes).filter(([id]) => view.nodes.has(id)));
    const edges: SigmaExploreEdge[] = items.edges.filter(
        (edge) => view.nodes.has(edge.source) && view.nodes.has(edge.target)
    );
    const summaryNodes = new Map<string, SummaryDetails>();
    const isReserved = (id: string) => plan.graph.nodes.has(id) || plan.graph.edges.has(id) || summaryNodes.has(id);

    for (const group of view.collapsedGroups.values()) {
        if (group.kind !== 'leaves') continue;

        const member = group.members[0];
        const node = items.nodes[member.nodeId];
        const relationshipKind = plan.graph.edges.get(member.edgeId)?.kind ?? '';
        const nodeId = getCollapseSummaryId(group.id, 'node', isReserved);
        const edgeId = getCollapseSummaryId(group.id, 'edge', isReserved);

        nodes[nodeId] = {
            kind: node.kind,
            kinds: node.kinds,
            isTierZero: node.isTierZero,
            isOwnedObject: node.isOwnedObject,
            label: `${node.kind} nodes`,
            objectId: '',
            lastSeen: '',
        };
        summaryNodes.set(nodeId, { groupId: group.id, count: group.members.length });

        const towardAnchor = group.direction === 'toward-anchor';
        edges.push({
            source: towardAnchor ? nodeId : group.anchorId,
            target: towardAnchor ? group.anchorId : nodeId,
            label: relationshipKind,
            kind: relationshipKind,
            summaryId: edgeId,
            collapseGroupId: group.id,
        });
    }

    return { nodes, edges, summaryNodes };
}
