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

import { graphSchema } from './constants';
import { ActiveDirectoryKindProperties, AzureKindProperties, CommonKindProperties } from './graphSchema';

describe('graphSchema', () => {
    it('returns default empty labels and relationshipTypes when called with no arguments', () => {
        const result = graphSchema({ nodes: undefined, edges: undefined });

        expect(result.labels).toEqual([]);
        expect(result.relationshipTypes).toEqual([]);
    });

    it('returns propertyKeys from all three kind property enums', () => {
        const result = graphSchema({ nodes: undefined, edges: undefined });

        const expectedPropertyKeys = [
            ...Object.values(CommonKindProperties),
            ...Object.values(ActiveDirectoryKindProperties),
            ...Object.values(AzureKindProperties),
        ];

        expect(result.propertyKeys).toEqual(expectedPropertyKeys);
    });

    it('prefixes each node_kind with a colon for labels', () => {
        const result = graphSchema({
            nodes: ['User', 'Computer', 'Group'],
            edges: [],
        });

        expect(result.labels).toEqual([':User', ':Computer', ':Group']);
    });

    it('prefixes each edge_kind with a colon for relationshipTypes', () => {
        const result = graphSchema({
            nodes: [],
            edges: ['MemberOf', 'HasSession', 'AdminTo'],
        });

        expect(result.relationshipTypes).toEqual([':MemberOf', ':HasSession', ':AdminTo']);
    });

    it('handles both node_kinds and edge_kinds together', () => {
        const result = graphSchema({
            nodes: ['User', 'Domain'],
            edges: ['MemberOf', 'Contains'],
        });

        expect(result.labels).toEqual([':User', ':Domain']);
        expect(result.relationshipTypes).toEqual([':MemberOf', ':Contains']);
    });

    it('returns empty arrays when kinds data has empty arrays', () => {
        const result = graphSchema({
            nodes: [],
            edges: [],
        });

        expect(result.labels).toEqual([]);
        expect(result.relationshipTypes).toEqual([]);
    });
});
