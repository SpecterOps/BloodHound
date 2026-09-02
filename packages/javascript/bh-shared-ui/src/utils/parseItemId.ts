// Copyright 2025 Specter Ops, Inc.
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

import { escapeCypherString } from './cypher';

export interface ParsedQueryItem {
    itemType: 'edge' | 'node' | 'none';
    cypherQuery: string;
    id: string;
}

export const parseItemId = (itemId: string): ParsedQueryItem => {
    let match = itemId.match(/^node_(.+)$/);
    if (match) {
        return {
            itemType: 'node',
            id: match[1],
            cypherQuery: `MATCH (n) WHERE n.objectid = ${escapeCypherString(match[1])} RETURN n LIMIT 1`,
        };
    }

    // `edge_||:<sourceObjectId>||:<edgeType>||:<targetObjectId>` for relationship findings
    match = itemId.match(/^edge_\|\|:(.+)\|\|:(.+)\|\|:(.+)$/);
    if (match) {
        const [, sourceObjectId, edgeType, targetObjectId] = match;
        return {
            itemType: 'edge',
            id: '',
            cypherQuery: `MATCH p=(s)-[r:${edgeType}]->(t) WHERE s.objectid = ${escapeCypherString(sourceObjectId)} AND t.objectid = ${escapeCypherString(targetObjectId)}  RETURN p LIMIT 1`,
        };
    }

    return {
        itemType: 'none',
        id: '',
        cypherQuery: '',
    };
};

// Some constants and helper functions useful for handling the ID formats that parseItemId can work with
export const NODE_ID_PREFIX = 'node_';
export const EDGE_ID_PREFIX = 'edge_';
export const REL_ID_PREFIX = 'rel_';
export const EDGE_ID_SEPARATOR = '||:';
export const REL_ID_SEPARATOR = '_';

export const createNodeItemId = (objectId: string): string => {
    return NODE_ID_PREFIX + objectId;
};

export const createEdgeItemId = (sourceObjectId: string, edgeType: string, targetObjectId: string): string => {
    return [EDGE_ID_PREFIX, sourceObjectId, edgeType, targetObjectId].join(EDGE_ID_SEPARATOR);
};

export const createRelItemId = (sourceGraphId: string, edgeType: string, targetGraphId: string): string => {
    return REL_ID_PREFIX + [sourceGraphId, edgeType, targetGraphId].join(REL_ID_SEPARATOR);
};
