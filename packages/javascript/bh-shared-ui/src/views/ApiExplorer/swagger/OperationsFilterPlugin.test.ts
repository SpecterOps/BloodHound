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

import { fromJS } from 'immutable';
import { compareTerm, opsFilter } from './OperationsFilterPlugin';

const buildTaggedOps = (data: Record<string, { operations: { path: string }[] }>) => fromJS(data);

describe('compareTerm', () => {
    it('should return true when value matches term', () => {
        expect(compareTerm('auth', 'auth')).toEqual(true);
    });

    it('should return true when value partially matches term', () => {
        expect(compareTerm('auth', 'au')).toEqual(true);
    });

    it('should ignore case when filtering', () => {
        expect(compareTerm('user', 'USER')).toEqual(true);
        expect(compareTerm('USER', 'user')).toEqual(true);
    });

    it('should return false when there is no match', () => {
        expect(compareTerm('user', 'auth')).toEqual(false);
    });
});

describe('opsFilter', () => {
    const taggedOps = buildTaggedOps({
        Auth: { operations: [{ path: '/api/v1/login' }, { path: '/api/v1/logout' }] },
        Users: { operations: [{ path: '/api/v1/users' }, { path: '/api/v1/users/{id}' }] },
        Graph: { operations: [{ path: '/api/v1/graph/nodes' }] },
    });

    it('should include all operations when the tag matches the term', () => {
        const result = opsFilter(taggedOps, 'Auth');
        expect(result.has('Auth')).toBe(true);
        expect(result.getIn(['Auth', 'operations']).size).toBe(2);
    });

    it('should exclude tags and operations that do not match the term', () => {
        const result = opsFilter(taggedOps, 'Auth');
        expect(result.has('Users')).toBe(false);
        expect(result.has('Graph')).toBe(false);
    });

    it('should include a tag with only the matching operations when the path matches', () => {
        const result = opsFilter(taggedOps, '/api/v1/users/{id}');
        expect(result.has('Users')).toBe(true);
        expect(result.getIn(['Users', 'operations']).size).toBe(1);
        expect(result.getIn(['Users', 'operations', 0, 'path'])).toBe('/api/v1/users/{id}');
    });

    it('should exclude operations within a tag when only the path matches but not all paths', () => {
        const result = opsFilter(taggedOps, 'login');
        expect(result.has('Auth')).toBe(true);
        expect(result.getIn(['Auth', 'operations']).size).toBe(1);
        expect(result.getIn(['Auth', 'operations', 0, 'path'])).toBe('/api/v1/login');
    });

    it('should be case-insensitive when matching by path', () => {
        const result = opsFilter(taggedOps, 'GRAPH');
        expect(result.has('Graph')).toBe(true);
        expect(result.getIn(['Graph', 'operations']).size).toBe(1);
    });

    it('should return an empty map when no tags or paths match', () => {
        const result = opsFilter(taggedOps, 'nonexistent');
        expect(result.size).toBe(0);
    });
});
