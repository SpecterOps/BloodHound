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

import type { FlatGraphResponse, GraphEdge, GraphNode } from 'js-client-library';

export type SourceNode = Readonly<{
    id: string;
    kind: string;
    node: GraphNode;
    original: GraphNode | FlatGraphResponse[string];
}>;

export type SourceEdge = Readonly<{
    id: string;
    source: string;
    target: string;
    kind: string;
    edge: GraphEdge;
    original: GraphEdge | FlatGraphResponse[string];
}>;

export type SourceGraph = Readonly<{
    nodes: Readonly<Record<string, SourceNode>>;
    edges: Readonly<Record<string, SourceEdge>>;
    incidentEdges: Readonly<Record<string, readonly string[]>>;
    endpointEdges: Readonly<Record<string, readonly string[]>>;
}>;

export type SourceItemRef = Readonly<{
    kind: 'node' | 'edge';
    id: string;
}>;

export type NodeGroup = Readonly<{
    kind: 'nodes';
    id: string;
    ruleId: string;
    memberNodeIds: readonly string[];
}>;

export type RelationshipGroup = Readonly<{
    kind: 'relationships';
    id: string;
    ruleId: string;
    sourceNodeId: string;
    targetNodeId: string;
    memberEdgeIds: readonly string[];
}>;

export type GroupDefinition = NodeGroup | RelationshipGroup;

export type GroupingRule = Readonly<{
    id: string;
    discover: (source: SourceGraph) => readonly GroupDefinition[];
}>;

export type ProjectionState = Readonly<{
    expandedGroupIds: ReadonlySet<string>;
}>;

export type ProjectedNode =
    | Readonly<{
          id: string;
          kind: 'source-node';
          sourceNodeIds: readonly string[];
      }>
    | Readonly<{
          id: string;
          kind: 'node-group';
          groupId: string;
          sourceNodeIds: readonly string[];
      }>
    | Readonly<{
          id: string;
          kind: 'relationship-group';
          groupId: string;
          sourceEdgeIds: readonly string[];
      }>;

export type ProjectedEdge = Readonly<{
    id: string;
    source: string;
    target: string;
    kind: 'source-edge' | 'connector';
    sourceEdgeIds: readonly string[];
    groupId?: string;
}>;

export type ProjectionIndex = Readonly<{
    projectedToSource: Readonly<Record<string, readonly SourceItemRef[]>>;
    sourceNodeToVisible: Readonly<Record<string, string>>;
    sourceEdgeToVisible: Readonly<Record<string, readonly string[]>>;
    sourceNodeGroups: Readonly<Record<string, readonly string[]>>;
    sourceEdgeGroups: Readonly<Record<string, readonly string[]>>;
}>;

export type ProjectedGraph = Readonly<{
    nodes: Readonly<Record<string, ProjectedNode>>;
    edges: Readonly<Record<string, ProjectedEdge>>;
    index: ProjectionIndex;
}>;
