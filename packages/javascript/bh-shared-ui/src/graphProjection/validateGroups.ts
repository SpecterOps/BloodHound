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

import type { GroupDefinition, SourceGraph } from './types';

export type GroupValidationResult =
    | { valid: true; groups: readonly GroupDefinition[] }
    | { valid: false; diagnostics: readonly string[] };

export function validateGroups(source: SourceGraph, groups: readonly GroupDefinition[]): GroupValidationResult {
    const ids = new Set<string>();
    const nodeMembers = new Set<string>();
    const edgeMembers = new Set<string>();
    const diagnostics: string[] = [];
    for (const group of groups) {
        if (ids.has(group.id)) {
            diagnostics.push(`Duplicate group: ${group.id}`);
        }
        ids.add(group.id);
        const memberIds = group.kind === 'nodes' ? group.memberNodeIds : group.memberEdgeIds;
        const occupiedMembers = group.kind === 'nodes' ? nodeMembers : edgeMembers;
        if (memberIds.length < 2) {
            diagnostics.push(`Insufficient membership: ${group.id}`);
        }

        for (const memberId of memberIds) {
            if (occupiedMembers.has(memberId)) {
                diagnostics.push(`Overlapping ${group.kind} membership: ${memberId}`);
            }
            occupiedMembers.add(memberId);

            if (group.kind === 'nodes') {
                if (!Object.hasOwn(source.nodes, memberId)) {
                    diagnostics.push(`Missing node: ${memberId}`);
                }
            } else {
                const edge = source.edges[memberId];
                if (
                    !edge ||
                    edge.source !== group.sourceNodeId ||
                    edge.target !== group.targetNodeId ||
                    !Object.hasOwn(source.nodes, edge.source) ||
                    !Object.hasOwn(source.nodes, edge.target)
                ) {
                    diagnostics.push(`Invalid relationship membership: ${memberId}`);
                }
            }
        }
    }

    if (diagnostics.length > 0) {
        return { valid: false, diagnostics: diagnostics.sort() };
    }
    return { valid: true, groups };
}
