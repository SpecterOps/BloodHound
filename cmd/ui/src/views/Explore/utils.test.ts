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

import { Theme } from '@mui/material';
import * as dagre from 'src/rendering/utils/dagre';
import { initGraph } from './utils';

const layoutDagreSpy = vi.spyOn(dagre, 'setDagreLayout');

const testNodes = {
    '1': {
        label: 'User 1',
        kind: 'User',
        kinds: ['User'],
        objectId: 'test-user-1',
        lastSeen: '',
        isTierZero: false,
        isOwnedObject: false,
    },
    '2': {
        label: 'Group 1',
        kind: 'Group',
        kinds: ['Group'],
        objectId: 'test-group-1',
        lastSeen: '',
        isTierZero: false,
        isOwnedObject: false,
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

describe('Explore utils', () => {
    describe('initGraph', () => {
        const mockTheme = {
            palette: {
                color: { primary: '', links: '' },
                neutral: { primary: '', secondary: '' },
                common: { black: '', white: '' },
            },
        };
        it('uses sequentialLayout by default', () => {
            initGraph(
                { nodes: {}, edges: [] },
                {
                    theme: mockTheme as Theme,
                    hideNodes: false,
                    customIcons: {},
                    darkMode: false,
                    tagGlyphs: {},
                }
            );

            expect(layoutDagreSpy).toBeCalled();
        });
        it('dedupes edges before adding them to the graph', () => {
            const graph = initGraph(
                { nodes: testNodes, edges: testEdgesWithDuplicate },
                {
                    theme: mockTheme as Theme,
                    hideNodes: false,
                    customIcons: {},
                    darkMode: false,
                    tagGlyphs: {},
                }
            );

            expect(graph.edges().length).toEqual(3);
            expect(graph.hasEdge('1_MemberOf_2')).toBeTruthy();
            expect(graph.hasEdge('2_MemberOf_1')).toBeTruthy();
            expect(graph.hasEdge('1_GetChangesAll_2')).toBeTruthy();
        });
    });
});
