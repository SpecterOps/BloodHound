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
import { MultiDirectedGraph } from 'graphology';
import * as layoutDagre from 'src/hooks/useLayoutDagre/useLayoutDagre';
import { initGraph } from './utils';

const layoutDagreSpy = vi.spyOn(layoutDagre, 'layoutDagre');

describe('Explore utils', () => {
    describe('initGraph', () => {
        const mockTheme = {
            palette: {
                color: { primary: '', links: '' },
                neutral: { primary: '', secondary: '' },
                common: { black: '', white: '' },
            },
        };
        it('calls sequentialLayout as the default graph layout', () => {
            const graph = new MultiDirectedGraph();
            initGraph(graph, { nodes: {}, edges: [] }, mockTheme as Theme, false);

            expect(layoutDagreSpy).toBeCalled();
        });
    });
});
