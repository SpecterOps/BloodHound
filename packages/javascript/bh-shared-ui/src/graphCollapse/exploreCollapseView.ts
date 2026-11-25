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

import { NODE_ID_PREFIX, REL_ID_PREFIX, createRelItemId } from '../utils/parseItemId';
import { projectGraphCollapse, type CollapsePlan, type CollapseViewOptions } from './collapsePlan';

function getObjectId(plan: CollapsePlan, nodeId: string): string | undefined {
    const original = plan.graph.nodes.get(nodeId)?.original;
    if (!original) return;
    return 'objectId' in original ? original.objectId : original.data?.objectid;
}

/** Resolve a selection to a group so the canvas can reuse its layout for ordinary selections. */
export function getExploreCollapseGroupId(plan?: CollapsePlan, selectedId?: string | null): string | undefined {
    if (!plan || !selectedId) return;

    const directGroupId = plan.groupByNodeId.get(selectedId) ?? plan.groupByEdgeId.get(selectedId);
    if (directGroupId !== undefined || plan.graph.edges.has(selectedId)) return directGroupId;

    if (selectedId.startsWith(NODE_ID_PREFIX)) {
        const objectId = selectedId.slice(NODE_ID_PREFIX.length);
        for (const [nodeId, groupId] of plan.groupByNodeId) {
            if (getObjectId(plan, nodeId) === objectId) return groupId;
        }
    } else if (selectedId.startsWith(REL_ID_PREFIX)) {
        for (const [edgeId, groupId] of plan.groupByEdgeId) {
            const edge = plan.graph.edges.get(edgeId);
            if (
                edge &&
                (edge.sourceKey === selectedId || createRelItemId(edge.source, edge.kind, edge.target) === selectedId)
            )
                return groupId;
        }
    }
}

export function projectExploreCollapse(
    plan: CollapsePlan,
    selectedGroupId?: string,
    protectedObjectIds: readonly (string | null | undefined)[] = [],
    options: CollapseViewOptions = {}
) {
    if (options.mode === 'expanded') return projectGraphCollapse(plan, options);

    const expandedGroupIds = new Set(options.expandedGroupIds);
    if (selectedGroupId !== undefined) expandedGroupIds.add(selectedGroupId);

    const objectIds = new Set(protectedObjectIds.filter(Boolean));
    if (objectIds.size > 0) {
        for (const [nodeId, groupId] of plan.groupByNodeId) {
            if (objectIds.has(getObjectId(plan, nodeId))) expandedGroupIds.add(groupId);
        }
    }

    return projectGraphCollapse(plan, { ...options, expandedGroupIds });
}

export function getCollapseSummaryId(
    groupId: string,
    itemType: 'node' | 'edge',
    isReserved: (id: string) => boolean
): string {
    const baseId = `collapse-${itemType}:${groupId}`;
    let id = baseId;
    for (let suffix = 1; isReserved(id); suffix += 1) id = `${baseId}:${suffix}`;
    return id;
}
