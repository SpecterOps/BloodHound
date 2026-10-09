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

import { EdgeType } from 'js-client-library';
import {
    ActiveDirectoryPathfindingEdges,
    ActiveDirectoryRelationshipKind,
    AzurePathfindingEdges,
    AzureRelationshipKind,
} from '../graphSchema';
import { useEdgeTypesQuery } from './useEdgeTypes';

export type EdgeTraversability = ReadonlyMap<string, boolean>;

/** Only explicitly traversable edge kinds return true; unknown kinds default to false. */
class EdgeTraversabilityMap extends Map<string, boolean> {
    override get(kind: string): boolean {
        return super.get(kind) === true;
    }
}

const builtinPathfindingEdges = new Set<string>([...ActiveDirectoryPathfindingEdges(), ...AzurePathfindingEdges()]);
const builtinTraversability: EdgeTraversability = new EdgeTraversabilityMap(
    [...Object.values(ActiveDirectoryRelationshipKind), ...Object.values(AzureRelationshipKind)].map((kind) => [
        kind,
        builtinPathfindingEdges.has(kind),
    ])
);

const selectTraversability = (edgeTypes: EdgeType[]): EdgeTraversability =>
    new EdgeTraversabilityMap([
        ...builtinTraversability,
        ...edgeTypes.map((edgeType): [string, boolean] => [edgeType.name, edgeType.is_traversable === true]),
    ]);

// Use the unfiltered schema, including built-in and non-traversable edge kinds.
export const useEdgeTraversability = () => {
    const query = useEdgeTypesQuery(selectTraversability);
    // Built-in rendering also works when the schema endpoint is disabled or unavailable.
    return { ...query, data: query.data ?? builtinTraversability };
};
