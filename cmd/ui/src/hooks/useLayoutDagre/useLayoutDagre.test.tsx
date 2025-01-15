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
    copySigmaNodesToGraphlibGraph,
    applyNodePositionsFromGraphlibGraph,
    NODE_DEFAULT_SIZE,
} from 'src/hooks/useLayoutDagre/useLayoutDagre';
import Graph from 'graphology';
import dagre from 'dagrejs';

const sigmaGraph = new Graph();
const graphlibGraph = new dagre.graphlib.Graph();

const testNode = 'test';
const testNode2 = 'test2';

describe('copySigmaNodesToGraphlibGraph', () => {
    test('Default label and sizing are used when graph of origin does not specify', () => {
        //No label or sizing are defined in the source graph
        sigmaGraph.addNode(testNode);

        copySigmaNodesToGraphlibGraph(sigmaGraph, graphlibGraph);

        expect(graphlibGraph.node(testNode).label).not.toBeUndefined();
        expect(graphlibGraph.node(testNode).label).toBe('');

        expect(graphlibGraph.node(testNode).height).not.toBeUndefined();
        expect(graphlibGraph.node(testNode).height).toBe(NODE_DEFAULT_SIZE);

        expect(graphlibGraph.node(testNode).width).not.toBeUndefined();
        expect(graphlibGraph.node(testNode).width).toBe(NODE_DEFAULT_SIZE);
    });

    test('Origin label and sizing are used when specified', () => {
        const size = 12;
        const label = 'testlabel';

        //Label and size defined in source graph
        sigmaGraph.addNode(testNode2, { label: label, size: size });

        copySigmaNodesToGraphlibGraph(sigmaGraph, graphlibGraph);

        expect(graphlibGraph.node(testNode2).height).toBe(size);
        expect(graphlibGraph.node(testNode2).width).toBe(size);
        expect(graphlibGraph.node(testNode2).label).toBe(label);
    });
});

describe('applyNodePositionsFromGraphlibGraph', () => {
    test('A warning is printed when node positions are not available to apply from', () => {
        console.warn = vi.fn(); //To keep test logs clean
        const warnLogSpy = vi.spyOn(console, 'warn');

        applyNodePositionsFromGraphlibGraph(sigmaGraph, graphlibGraph);
        expect(warnLogSpy).toBeCalledWith('incomplete information for applying graphlib node position to sigma');
        expect(warnLogSpy).toBeCalledTimes(2);
    });
});
