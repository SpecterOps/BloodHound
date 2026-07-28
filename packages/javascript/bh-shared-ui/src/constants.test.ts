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
import { darkPalette, graphSchema, lightPalette, themedComponents, typography } from './constants';
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

describe('legacy MUI typography bridge', () => {
    it('uses the DoodleUI font families', () => {
        expect(typography.fontFamily).toContain('Figtree');
        expect(typography.h1).toMatchObject({ fontFamily: expect.stringContaining('Nunito Sans') });
        expect(typography.h6).toMatchObject({ fontFamily: expect.stringContaining('Nunito Sans') });
    });

    it('mirrors the approved DoodleUI heading metrics', () => {
        expect(typography).toMatchObject({
            h1: { fontSize: '1.5rem', lineHeight: '1.75rem', fontWeight: 700, letterSpacing: 0 },
            h2: { fontSize: '1.375rem', lineHeight: '1.5rem', fontWeight: 700, letterSpacing: 0 },
            h3: { fontSize: '1.25rem', lineHeight: '1.375rem', fontWeight: 700, letterSpacing: 0 },
            h4: { fontSize: '1.25rem', lineHeight: '1.375rem', fontWeight: 600, letterSpacing: 0 },
            h5: { fontSize: '1.125rem', lineHeight: '1.25rem', fontWeight: 700, letterSpacing: '.25px' },
            h6: { fontSize: '1rem', lineHeight: '1.125rem', fontWeight: 600, letterSpacing: '.25px' },
        });
    });

    it('mirrors the approved DoodleUI body, subtitle, and caption metrics', () => {
        expect(typography).toMatchObject({
            body1: { fontSize: '1rem', lineHeight: '1.5rem', fontWeight: 400, letterSpacing: 0 },
            body2: { fontSize: '.875rem', lineHeight: '1.375rem', fontWeight: 400, letterSpacing: 0 },
            subtitle1: { fontSize: '.9375rem', lineHeight: '1.5rem', fontWeight: 500, letterSpacing: '.25px' },
            subtitle2: { fontSize: '.8125rem', lineHeight: '1.375rem', fontWeight: 500, letterSpacing: '.25px' },
            caption: { fontSize: '.75rem', lineHeight: '1.25rem', fontWeight: 400, letterSpacing: '.25px' },
        });
    });

    it('uses the semantic muted-text token for light-mode body and caption variants', () => {
        const typographyOverrides = themedComponents(lightPalette)?.MuiTypography?.styleOverrides;

        expect(typographyOverrides).toMatchObject({
            body1: { color: 'var(--text-muted)' },
            body2: { color: 'var(--text-muted)' },
            caption: { color: 'var(--text-muted)' },
        });
    });

    it('does not change legacy MUI typography colors in dark mode', () => {
        const typographyOverrides = themedComponents(darkPalette)?.MuiTypography?.styleOverrides;

        expect(typographyOverrides).toEqual({});
    });
});
