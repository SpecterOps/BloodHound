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

import type { CollapseRule, LeafCollapseProposal } from './collapsePlan';

export const collapseSimilarLeaves: CollapseRule = {
    id: 'similar-leaves',
    propose(graph) {
        const groups = new Map<
            string,
            Omit<LeafCollapseProposal, 'members'> & { members: LeafCollapseProposal['members'][number][] }
        >();

        for (const [nodeId, node] of graph.nodes) {
            const incident = graph.incidentEdges.get(nodeId);
            if (incident?.length !== 1 || !node.kind || node.kind === 'HIDDEN') continue;

            const edgeId = incident[0];
            const edge = graph.edges.get(edgeId);
            if (!edge?.kind || edge.relationshipId === undefined || edge.source === edge.target) continue;

            const isSource = edge.source === nodeId;
            const anchorId = isSource ? edge.target : edge.source;
            const anchor = graph.nodes.get(anchorId);
            if (!anchor || anchor.kind === 'HIDDEN') continue;

            const direction = isSource ? 'toward-anchor' : 'away-from-anchor';
            const signature = JSON.stringify([
                anchorId,
                direction,
                edge.kind,
                node.kind,
                [...new Set(node.kinds)].sort(),
                [...new Set(node.systemTags)].sort(),
                node.isTierZero,
                node.isOwnedObject,
            ]);
            const member = { nodeId, edgeId };
            const group = groups.get(signature);
            if (group) {
                group.members.push(member);
            } else {
                groups.set(signature, {
                    kind: 'leaves',
                    key: [signature],
                    anchorId,
                    direction,
                    members: [member],
                });
            }
        }

        return [...groups.values()].filter((group) => group.members.length >= 2);
    },
};
