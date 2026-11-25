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

import type { CollapseGraph, CollapseTopology } from './sourceGraph';

type LeafMember = Readonly<{ nodeId: string; edgeId: string }>;

export type LeafCollapseProposal = Readonly<{
    kind: 'leaves';
    key: readonly string[];
    anchorId: string;
    direction: 'toward-anchor' | 'away-from-anchor';
    members: readonly LeafMember[];
}>;

export type ParallelCollapseProposal = Readonly<{
    kind: 'parallel';
    key: readonly string[];
    source: string;
    target: string;
    edgeIds: readonly string[];
}>;

type CollapseProposal = LeafCollapseProposal | ParallelCollapseProposal;

export type CollapseGroup = CollapseProposal & Readonly<{ id: string; ruleId: string }>;

export type CollapseRule = Readonly<{
    id: string;
    /** Proposal keys identify groups within the rule and should be independent of item order. */
    propose: (graph: CollapseTopology) => readonly CollapseProposal[];
}>;

export type CollapsePlan = Readonly<{
    graph: CollapseGraph;
    groups: ReadonlyMap<string, CollapseGroup>;
    groupByNodeId: ReadonlyMap<string, string>;
    groupByEdgeId: ReadonlyMap<string, string>;
}>;

export type CollapseViewOptions = Readonly<{
    mode?: 'collapsed' | 'expanded';
    expandedGroupIds?: ReadonlySet<string>;
}>;

export type CollapseView = Pick<CollapseGraph, 'nodes' | 'edges'> &
    Readonly<{ collapsedGroups: ReadonlyMap<string, CollapseGroup> }>;

function isValidLeafGroup(graph: CollapseGraph, group: LeafCollapseProposal): boolean {
    return group.members.every(({ nodeId, edgeId }) => {
        const edge = graph.edges.get(edgeId);
        if (!edge || nodeId === group.anchorId || !graph.nodes.has(nodeId)) return false;
        if (graph.incidentEdges.get(nodeId)?.length !== 1) return false;

        return group.direction === 'toward-anchor'
            ? edge.source === nodeId && edge.target === group.anchorId
            : edge.source === group.anchorId && edge.target === nodeId;
    });
}

/** Evaluate every rule on the original graph. Earlier valid proposals own their items. */
export function planGraphCollapse(graph: CollapseGraph, rules: readonly CollapseRule[]): CollapsePlan {
    const groups = new Map<string, CollapseGroup>();
    const groupByNodeId = new Map<string, string>();
    const groupByEdgeId = new Map<string, string>();
    const requiredNodes = new Set<string>();

    for (const rule of rules) {
        for (const proposal of rule.propose(graph)) {
            const id = JSON.stringify([rule.id, proposal.key]);
            const nodeIds = proposal.kind === 'leaves' ? proposal.members.map((member) => member.nodeId) : [];
            const edgeIds =
                proposal.kind === 'leaves' ? proposal.members.map((member) => member.edgeId) : proposal.edgeIds;
            const endpoints = proposal.kind === 'leaves' ? [proposal.anchorId] : [proposal.source, proposal.target];

            if (
                groups.has(id) ||
                edgeIds.length < 2 ||
                new Set(edgeIds).size !== edgeIds.length ||
                endpoints.some((nodeId) => !graph.nodes.has(nodeId) || groupByNodeId.has(nodeId)) ||
                nodeIds.some((nodeId) => groupByNodeId.has(nodeId) || requiredNodes.has(nodeId)) ||
                edgeIds.some((edgeId) => groupByEdgeId.has(edgeId))
            )
                continue;

            const valid =
                proposal.kind === 'leaves'
                    ? isValidLeafGroup(graph, proposal)
                    : edgeIds.every((edgeId) => {
                          const edge = graph.edges.get(edgeId);
                          return edge?.source === proposal.source && edge.target === proposal.target;
                      });
            if (!valid) continue;

            groups.set(id, { ...proposal, id, ruleId: rule.id });
            endpoints.forEach((nodeId) => requiredNodes.add(nodeId));
            nodeIds.forEach((nodeId) => groupByNodeId.set(nodeId, id));
            edgeIds.forEach((edgeId) => groupByEdgeId.set(edgeId, id));
        }
    }

    return { graph, groups, groupByNodeId, groupByEdgeId };
}

function getVisibleItems<T>(
    items: ReadonlyMap<string, T>,
    groupById: ReadonlyMap<string, string>,
    collapsedGroups: ReadonlyMap<string, CollapseGroup>
): ReadonlyMap<string, T> {
    if (groupById.size === 0) return items;

    const visible = new Map<string, T>();
    for (const [id, item] of items) {
        const groupId = groupById.get(id);
        if (groupId === undefined || !collapsedGroups.has(groupId)) visible.set(id, item);
    }
    return visible;
}

/** Return visible source records and summary groups for the renderer to draw. */
export function projectGraphCollapse(plan: CollapsePlan, options: CollapseViewOptions = {}): CollapseView {
    const { graph, groups, groupByNodeId, groupByEdgeId } = plan;
    const collapsedGroups = new Map<string, CollapseGroup>();
    if (options.mode !== 'expanded') {
        for (const [id, group] of groups) {
            if (!options.expandedGroupIds?.has(id)) collapsedGroups.set(id, group);
        }
    }

    if (collapsedGroups.size === 0) return { nodes: graph.nodes, edges: graph.edges, collapsedGroups };

    return {
        nodes: getVisibleItems(graph.nodes, groupByNodeId, collapsedGroups),
        edges: getVisibleItems(graph.edges, groupByEdgeId, collapsedGroups),
        collapsedGroups,
    };
}
