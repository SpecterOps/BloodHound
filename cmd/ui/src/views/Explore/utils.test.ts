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

import { lightTheme } from 'bh-shared-ui';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';
import * as dagre from 'src/rendering/utils/dagre';
import { getNodeGlyphs, initGraph } from './utils';

const layoutDagreSpy = vi.spyOn(dagre, 'setDagreLayout');

const testNodes = {
    '1': {
        label: 'User 1',
        kind: 'User',
        kinds: ['User', 'Tag_Tier_Zero'],
        objectId: 'test-user-1',
        lastSeen: '',
        isTierZero: true,
        isOwnedObject: false,
    },
    '2': {
        label: 'Group 1',
        kind: 'Group',
        kinds: ['Group', 'Tag_Owned'],
        objectId: 'test-group-1',
        lastSeen: '',
        isTierZero: false,
        isOwnedObject: true,
    },
};

const testEdgesWithDuplicate = [
    {
        source: '1',
        target: '2',
        label: 'MemberOf',
        kind: 'MemberOf',
        lastSeen: '',
    },
    {
        source: '1',
        target: '2',
        label: 'MemberOf',
        kind: 'MemberOf',
        lastSeen: '',
    },
    {
        source: '2',
        target: '1',
        label: 'MemberOf',
        kind: 'MemberOf',
        lastSeen: '',
    },
    {
        source: '1',
        target: '2',
        label: 'GetChangesAll',
        kind: 'GetChangesAll',
        lastSeen: '',
    },
];

const pzFeatureFlagEnabled = true;

describe('Explore utils', () => {
    describe('initGraph', () => {
        it('uses sequentialLayout by default', () => {
            initGraph(
                { nodes: {}, edges: [] },
                {
                    theme: lightTheme,
                    hideNodes: false,
                    customIcons: {},
                    darkMode: false,
                    tagGlyphs: {},
                    pzFeatureFlagEnabled,
                }
            );

            expect(layoutDagreSpy).toBeCalled();
        });
        it('dedupes edges before adding them to the graph', () => {
            const graph = initGraph(
                { nodes: testNodes, edges: testEdgesWithDuplicate },
                {
                    theme: lightTheme,
                    hideNodes: false,
                    customIcons: {},
                    darkMode: false,
                    tagGlyphs: {},
                    pzFeatureFlagEnabled,
                }
            );

            expect(graph.edges().length).toEqual(3);
            expect(graph.hasEdge('1_MemberOf_2')).toBeTruthy();
            expect(graph.hasEdge('2_MemberOf_1')).toBeTruthy();
            expect(graph.hasEdge('1_GetChangesAll_2')).toBeTruthy();
        });
    });
});

describe('getNodeGlyphs', () => {
    const themedOptions = {
        labels: {
            labelColor: '#1D1B20',
            backgroundColor: '#F4F4F4',
            highlightedBackground: '#1A30FF',
            highlightedText: '#fff',
        },
        nodeBorderColor: '#1D1B20',
        glyph: {
            colors: {
                backgroundColor: '#1D1B20',
                color: '#FFFFFF',
            },
        },
    };

    const standardGlyphs = {
        TIER_ZERO: 'tierZero',
        TIER_ZERO_DARK: 'tierZeroDark',
        OWNED_OBJECT: 'owned',
        OWNED_OBJECT_DARK: 'ownedDark',
    };

    const tagGlyphs = { Tag_Tier_Zero: 'gem', owned: 'Tag_Owned', ownedGlyph: 'skull' };

    it('determines glyphs with the tier_managent_engine feature flag disabled', () => {
        const firstNodeGlyphs = getNodeGlyphs(
            testNodes[1],
            {
                theme: lightTheme,
                hideNodes: false,
                customIcons: {},
                darkMode: false,
                themedOptions,
                tagGlyphs,
                pzFeatureFlagEnabled: false,
            },
            standardGlyphs
        );

        expect(firstNodeGlyphs).toHaveLength(1);

        const firstNodeGlyph = firstNodeGlyphs[0];

        expect(firstNodeGlyph.location).toBe(GlyphLocation.TOP_RIGHT);
        expect(firstNodeGlyph.backgroundColor).toBe(themedOptions.glyph.colors.backgroundColor);
        expect(firstNodeGlyph.color).toBe(themedOptions.glyph.colors.color);
        expect(firstNodeGlyph.image).toBe('tierZero');

        const secondNodeGlyphs = getNodeGlyphs(
            testNodes[2],
            {
                theme: lightTheme,
                hideNodes: false,
                customIcons: {},
                darkMode: false,
                themedOptions,
                tagGlyphs,
                pzFeatureFlagEnabled: false,
            },
            standardGlyphs
        );

        expect(secondNodeGlyphs).toHaveLength(1);

        const secondNodeGlyph = secondNodeGlyphs[0];

        expect(secondNodeGlyph.location).toBe(GlyphLocation.BOTTOM_RIGHT);
        expect(secondNodeGlyph.backgroundColor).toBe(themedOptions.glyph.colors.backgroundColor);
        expect(secondNodeGlyph.color).toBe(themedOptions.glyph.colors.color);
        expect(secondNodeGlyph.image).toBe('owned');
    });

    it('determines glyphs with the tier_managent_engine feature flag enabled', () => {
        const firstNodeGlyphs = getNodeGlyphs(
            testNodes[1],
            {
                theme: lightTheme,
                hideNodes: false,
                customIcons: {},
                darkMode: false,
                themedOptions,
                tagGlyphs,
                pzFeatureFlagEnabled: true,
            },
            standardGlyphs
        );

        expect(firstNodeGlyphs).toHaveLength(1);

        const firstNodeGlyph = firstNodeGlyphs[0];

        expect(firstNodeGlyph.location).toBe(GlyphLocation.TOP_RIGHT);
        expect(firstNodeGlyph.backgroundColor).toBe(themedOptions.glyph.colors.backgroundColor);
        expect(firstNodeGlyph.color).toBe(themedOptions.glyph.colors.color);
        expect(firstNodeGlyph.image).toBe('gem');

        const secondNodeGlyphs = getNodeGlyphs(
            testNodes[2],
            {
                theme: lightTheme,
                hideNodes: false,
                customIcons: {},
                darkMode: false,
                themedOptions,
                tagGlyphs,
                pzFeatureFlagEnabled: true,
            },
            standardGlyphs
        );

        expect(secondNodeGlyphs).toHaveLength(1);

        const secondNodeGlyph = secondNodeGlyphs[0];

        expect(secondNodeGlyph.location).toBe(GlyphLocation.BOTTOM_RIGHT);
        expect(secondNodeGlyph.backgroundColor).toBe(themedOptions.glyph.colors.backgroundColor);
        expect(secondNodeGlyph.color).toBe(themedOptions.glyph.colors.color);
        expect(secondNodeGlyph.image).toBe('skull');
    });
});
