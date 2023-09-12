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

import AbstractGraph, { Attributes } from 'graphology-types';
import dagre from 'dagrejs';
import { getEdgeDataFromKey } from 'src/ducks/graph/utils';

export const NODE_DEFAULT_SIZE = 10;

export enum RankDirection {
    TOP_BOTTOM = 'TB',
    BOTTOM_TOP = 'BT',
    LEFT_RIGHT = 'LR',
    RIGHT_LEFT = 'RL',
}

enum Align {
    UP_LEFT = 'UL',
    UP_RIGHT = 'UR',
    DOWN_LEFT = 'DL',
    DOWN_RIGHT = 'DR',
}

enum RankerAlgorithms {
    NETWORK_SIMPLEX = 'network-simplex',
    TIGHT_TREE = 'tight-tree',
    LONGEST_PATH = 'longest-path',
}

enum LabelPositiion {
    LEFT = 'l',
    CENTER = 'c',
    RIGHT = 'r',
}

type DagreGraphAttributes = {
    rankdir: RankDirection;
    align?: Align | undefined;
    nodesep?: number;
    edgesep?: number;
    ranksep: number;
    marginx?: number;
    marginy?: number;
    acyclicer?: 'greedy' | undefined;
    ranker?: RankerAlgorithms;
};

type DagreNodeAttributes = {
    width: number;
    height: number;
};

type DagreEdgeAttributes = {
    minlen: number;
    weight: number;
    width: number;
    height: number;
    labelpos: LabelPositiion;
    labeloffset: number;
};

export type DagreAttributes = {
    graph: DagreGraphAttributes;
    node: DagreNodeAttributes;
    edge: DagreEdgeAttributes;
};

export const copySigmaNodesToGraphlibGraph = (
    sigmaGraph: AbstractGraph<Attributes, Attributes, Attributes>,
    graphlibGraph: any
): void => {
    sigmaGraph.forEachNode((node: string) => {
        const { label, size } = sigmaGraph.getNodeAttributes(node);
        graphlibGraph.setNode(node, {
            label: label || '',
            width: size || NODE_DEFAULT_SIZE,
            height: size || NODE_DEFAULT_SIZE,
        });
    });
};

const copySigmaEdgesToGraphlibGraph = (
    sigmaGraph: AbstractGraph<Attributes, Attributes, Attributes>,
    graphlibGraph: any
): void => {
    sigmaGraph.forEachEdge((edge: string) => {
        const edgeData = getEdgeDataFromKey(edge);
        if (edgeData !== null) graphlibGraph.setEdge(edgeData.source, edgeData.target, edgeData.label);
    });
};

const sigmaGraphToGraphlibGraph = (
    sigmaGraph: AbstractGraph<Attributes, Attributes, Attributes>,
    graphlibGraph: any
): void => {
    copySigmaNodesToGraphlibGraph(sigmaGraph, graphlibGraph);
    copySigmaEdgesToGraphlibGraph(sigmaGraph, graphlibGraph);
};

export const applyNodePositionsFromGraphlibGraph = (
    sigmaGraph: AbstractGraph<Attributes, Attributes, Attributes>,
    graphlibGraph: any
): void => {
    graphlibGraph.nodes().forEach((node: any) => {
        const { x, y } = graphlibGraph.node(node);

        if (x && y && node !== 'ReadWrite') {
            sigmaGraph.updateNodeAttribute(node, 'x', () => x);
            sigmaGraph.updateNodeAttribute(node, 'y', () => y);
        } else console.warn('incomplete information for applying graphlib node position to sigma');
    });
};

type AtLeast<T, K extends keyof T> = Partial<T> & Pick<T, K>;
export type useLayoutDagreProps = AtLeast<DagreAttributes, 'graph'>;

export const layoutDagre = (
    attributes: useLayoutDagreProps,
    graph: AbstractGraph<Attributes, Attributes, Attributes> | undefined
): { assign: () => void } => {
    if (!graph || !graph.size) return { assign: () => {} };

    const assign = (): void => {
        //Initialize an empty graph in the graphlib format to be able to run dagre layout on
        const graphlibGraph = new dagre.graphlib.Graph({ directed: true, multigraph: true });
        graphlibGraph.setGraph({});
        graphlibGraph.setDefaultEdgeLabel(() => ({}));
        graphlibGraph.setDefaultNodeLabel(() => ({}));

        //Apply the current sigma graph information onto the graphlib graph
        sigmaGraphToGraphlibGraph(graph, graphlibGraph);

        const graphlibGraphGraph = graphlibGraph.graph();
        graphlibGraphGraph.rankdir = attributes.graph.rankdir;
        graphlibGraphGraph.ranksep = attributes.graph.ranksep;

        dagre.layout(graphlibGraph);

        //Extract the layout positions from the graphlib graph into the sigma graph for rendering
        applyNodePositionsFromGraphlibGraph(graph, graphlibGraph);
    };

    return { assign: assign };
};

export default layoutDagre;
