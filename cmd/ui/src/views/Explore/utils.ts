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

import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryKindPropertiesToDisplay,
    AzureKindProperties,
    AzureKindPropertiesToDisplay,
    CommonKindProperties,
    CommonKindPropertiesToDisplay,
    EntityField,
    EntityPropertyKind,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { GraphEdges, GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import startCase from 'lodash/startCase';
import { ZERO_VALUE_API_DATE } from 'src/constants';
import { GlyphKind } from 'bh-shared-ui';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';
import { EdgeDirection, EdgeParams, NodeParams } from 'src/utils';
import { NODE_ICON, GLYPHS, UNKNOWN_ICON } from './svgIcons';

export const formatObjectInfoFields = (props: any): EntityField[] => {
    let mappedFields: EntityField[] = [];
    const propKeys = Object.keys(props || {});

    for (let i = 0; i < propKeys.length; i++) {
        const value = props[propKeys[i]];
        // Don't display empty fields or fields with zero date values
        if (
            value === undefined ||
            value === '' ||
            value === ZERO_VALUE_API_DATE ||
            (typeof value === 'object' && isEmpty(value))
        )
            continue;

        const { kind, isKnownProperty } = validateProperty(propKeys[i]);

        if (isKnownProperty) {
            mappedFields.push({
                kind: kind,
                label: getFieldLabel(kind!, propKeys[i]),
                value: value,
                keyprop: propKeys[i],
            });
        } else {
            mappedFields.push({
                kind: kind,
                label: `${startCase(propKeys[i])}:`,
                value: value,
                keyprop: propKeys[i],
            });
        }
    }

    mappedFields = mappedFields.sort((a, b) => {
        return a.label!.localeCompare(b.label!);
    });

    return mappedFields;
};

const isActiveDirectoryProperty = (enumValue: ActiveDirectoryKindProperties): boolean => {
    return Object.values(ActiveDirectoryKindProperties).includes(enumValue);
};

const isAzureProperty = (enumValue: AzureKindProperties): boolean => {
    return Object.values(AzureKindProperties).includes(enumValue);
};

const isCommonProperty = (enumValue: CommonKindProperties): boolean => {
    return Object.values(CommonKindProperties).includes(enumValue);
};

export type ValidatedProperty = {
    isKnownProperty: boolean;
    kind: EntityPropertyKind;
};

export const validateProperty = (enumValue: string): ValidatedProperty => {
    if (isActiveDirectoryProperty(enumValue as ActiveDirectoryKindProperties))
        return { isKnownProperty: true, kind: 'ad' };
    if (isAzureProperty(enumValue as AzureKindProperties)) return { isKnownProperty: true, kind: 'az' };
    if (isCommonProperty(enumValue as CommonKindProperties)) return { isKnownProperty: true, kind: 'cm' };
    return { isKnownProperty: false, kind: null };
};

const getFieldLabel = (kind: string, key: string): string => {
    let label: string;

    switch (kind) {
        case 'ad':
            label = ActiveDirectoryKindPropertiesToDisplay(key as ActiveDirectoryKindProperties)!;
            break;
        case 'az':
            label = AzureKindPropertiesToDisplay(key as AzureKindProperties)!;
            break;
        case 'cm':
            label = CommonKindPropertiesToDisplay(key as CommonKindProperties)!;
            break;
        default:
            label = key;
    }

    return `${label}:`;
};

export const initGraphNodes = (graph: MultiDirectedGraph, nodes: GraphNodes, nodeSize: number) => {
    Object.keys(nodes).forEach((key: string) => {
        const node = nodes[key];
        // Set default node parameters
        const nodeParams: Partial<NodeParams> = {
            color: '#FFFFFF',
            type: 'combined',
            label: node.label,
            forceLabel: true,
        };

        const icon = NODE_ICON[node.kind] || UNKNOWN_ICON;
        nodeParams.color = icon.color;
        nodeParams.image = icon.url || '';

        // Tier zero nodes should be marked with a gem glyph
        if (node.isTierZero) {
            const glyph = GLYPHS[GlyphKind.TIER_ZERO];
            nodeParams.type = 'glyphs';
            nodeParams.glyphs = [
                {
                    location: GlyphLocation.TOP_RIGHT,
                    image: glyph.url || '',
                    backgroundColor: glyph.color,
                },
            ];
        }

        graph.addNode(key, {
            size: nodeSize,
            borderColor: '#000000',
            ...nodeParams,
        });
    });
};

export const initGraphEdges = (graph: MultiDirectedGraph, edges: GraphEdges) => {
    // Group edges with the same start and end nodes into arrays. Should be grouped regardless of direction
    const groupedEdges = edges.reduce<Record<string, GraphEdges>>((groups, edge) => {
        const identifiers = [edge.source, edge.target].sort();
        const id = `${identifiers[0]}_${identifiers[1]}`;

        if (!groups[id]) {
            groups[id] = [];
        }
        groups[id].push(edge);

        return groups;
    }, {});

    // Loop through our group arrays
    for (const group in groupedEdges) {
        const groupSize = groupedEdges[group].length;

        for (const [i, edge] of groupedEdges[group].entries()) {
            const key = `${edge.source}_${edge.kind}_${edge.target}`;

            // Set default values for single edges
            const edgeParams: Partial<EdgeParams> = {
                size: 3,
                type: 'arrow',
                label: edge.label,
                color: '#000000C0',
                groupPosition: 0,
                groupSize: 1,
                exploreGraphId: edge.exploreGraphId || key,
                forceLabel: true,
            };

            // Groups with odd-numbered totals should have a straight edge first, then curve the rest
            const edgeShouldBeCurved = groupSize > 1;

            // Handle edge groups that have a mix of directions that edges travel between source and target.
            // We can use the value of the enum to indicate which direction the curve should bend
            const groupStart = group.split('_')[0];
            const edgeDirection = groupStart === edge.source ? EdgeDirection.FORWARDS : EdgeDirection.BACKWARDS;

            if (edgeShouldBeCurved) {
                edgeParams.type = 'curved';
                edgeParams.groupPosition = i;
                edgeParams.groupSize = groupSize;
                edgeParams.direction = edgeDirection;
            }

            graph.addEdgeWithKey(key, edge.source, edge.target, edgeParams);
        }
    }
};
