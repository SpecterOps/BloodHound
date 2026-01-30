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

import extend from 'lodash/extend';
import keys from 'lodash/keys';
import pick from 'lodash/pick';
import pickBy from 'lodash/pickBy';
import { mapToRange, toGraphLinkColor, toWidth } from 'src/ducks/graph/colors';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import { Index, Items } from 'src/utils';

const getLinks: (data: Items) => Items[] = (data) => {
    const relKeys = Object.keys(data).filter((nodeKey) => {
        return nodeKey.startsWith('rel_');
    });
    const links: Items[] = Object.values(pick(data, relKeys)) as Items[];
    return links;
};

const getLinksIndex: (data: Items) => Index<Items> = (data) => {
    const relKeys = Object.keys(data).filter((nodeKey) => {
        return nodeKey.startsWith('rel_');
    });
    const links: { [index: string]: Items } = pick(data, relKeys) as {
        [index: string]: Items;
    };
    return links;
};

const getNodesIndex: (data: Items) => Index<Node> = (data) => {
    const nodeKeys = Object.keys(data).filter((nodeKey) => {
        return !nodeKey.startsWith('rel_');
    });
    const nodes: { [index: string]: Node } = pick(data, nodeKeys) as {
        [index: string]: Node;
    };
    return nodes;
};

export type Combo = Index<Combo | Node>;

type CombineOptions = {
    level: number;
    properties: any;
};

const getCombined: (data: Items, combine: CombineOptions) => Combo = (data, combine) => {
    const { level, properties } = combine;
    const nodes: Index<Node> = getNodesIndex(data);
    if (level > 0) {
        const combineProps = [...properties];
        const combo: Combo = combineNodes(nodes, combineProps);

        // Flatten levels if necessary
        let flatten = properties.length - level;
        while (flatten > 0) {
            keys(combo).forEach((key) => {
                if (key.startsWith('_combonode_')) {
                    const subCombo: Combo = combo[key] as Combo;
                    delete combo[key];
                    keys(subCombo).forEach((subKey) => {
                        combo[subKey] = subCombo[subKey];
                    });
                }
            });
            flatten--;
        }

        return combo;
    } else {
        return { ...nodes } as Combo;
    }
};

const combineNodes: (nodes: Index<any>, properties: (string | number)[], parentPath?: string) => Combo = (
    nodes,
    properties,
    parentPath = '_combonode'
) => {
    const acc: Index<Index<Node>> = {};
    const unclassified: Index<Node> = {};
    const property: string | number = properties.pop()!;

    keys(nodes).forEach((id) => {
        const node = nodes[id];
        const val: string = node['data'][property];
        if (val !== undefined && val !== null) {
            const comboKey = parentPath + '_' + val;
            if (acc[comboKey] === undefined) {
                acc[comboKey] = {};
            }
            acc[comboKey][id] = node;
        } else {
            unclassified[id] = node;
        }
    });

    const result: Combo = {};
    keys(acc).forEach((comboKey) => {
        if (properties.length > 0) {
            const subCombo: Combo = combineNodes(acc[comboKey], [...properties], comboKey);
            result[comboKey] = subCombo;
        } else {
            result[comboKey] = acc[comboKey];
        }
    });

    extend(result, unclassified);

    return result;
};

const withLinkImact = (data: Items): void => {
    applyImpactPct(data);
};

const applyImpactPct = (data: Items): void => {
    const impact = (rel: any) => Math.max(0, Math.log10(Math.round(rel.data.composite_risk_impact_percent * 100)));
    const rels: Items[] = getLinks(data);
    const weights = rels.map(impact);
    weights.sort((a, b) => a - b);

    const sourceRange: [number, number] = [weights[0], weights[weights.length - 1]];

    rels.forEach((rel) => {
        const imp = impact(rel);
        rel.color = toGraphLinkColor(imp, sourceRange, 0.4);
        rel.width = toWidth(imp, sourceRange);
        rel.data.impact_weighted = mapToRange(imp, sourceRange, [0, 100]);
    });
};

const applyRelWidths = (data: Items, width: number): void => {
    const rels: Items[] = getLinks(data);
    rels.forEach((rel) => {
        rel.width = width;
    });
};

const handleLabels = (data: Items): void => {
    Object.keys(data)
        .filter((key) => !key.startsWith('rel_'))
        .forEach((key) => {
            delete data[key].label!.text;
        });
};

const findNodes = (data: Items): Items[] => {
    const keys = Object.keys(data).filter((key) => {
        return !key.startsWith('rel_');
    });
    const nodes: Items[] = Object.values(pick(data, keys));
    return nodes;
};

const findRootId: (data: Items) => string | null = (data) => {
    const links: Index<Items> = getLinksIndex(data);
    const origins = Object.values(links).reduce<string[]>((acc, value) => {
        if (!acc.includes(value.id1)) {
            acc.push(value.id1);
        }
        return acc;
    }, []);
    const nodes: Index<Node> = getNodesIndex(data);
    const rootId = Object.keys(nodes).find((node) => !origins.includes(node));
    return rootId || null;
};

const findRootRelsIds: (data: Items) => string[] = (data) => {
    const rootId = findRootId(data);
    if (rootId != null) {
        const links: Index<Items> = getLinksIndex(data);
        const rootEdgesIndex = pickBy(links, (rel) => rel.id2 === rootId);
        if (rootEdgesIndex != null) {
            const rels: string[] = Object.keys(rootEdgesIndex);
            rels.sort((a, b) => {
                return data[b].data.impact_weighted - data[a].data.impact_weighted;
            });
            return rels;
        }
    }
    return [];
};

const findTierZeroNodeId: (data: Items) => string | null = (data) => {
    const nodes: Index<Items> = getNodesIndex(data);
    const tierZeroId = Object.keys(nodes).find((id) => nodes[id].data.level === 0);
    return tierZeroId || null;
};

const ICONS: { [id in GraphNodeTypes]: string } = {
    [GraphNodeTypes.AZBase]: 'fa-question',
    [GraphNodeTypes.AZApp]: 'fa-window-restore',
    [GraphNodeTypes.AZVMScaleSet]: 'fa-server',
    [GraphNodeTypes.AZRole]: 'fa-window-restore',
    [GraphNodeTypes.AZDevice]: 'fa-desktop',
    [GraphNodeTypes.AZFunctionApp]: 'fa-bolt',
    [GraphNodeTypes.AZGroup]: 'fa-users',
    [GraphNodeTypes.AZKeyVault]: 'fa-lock',
    [GraphNodeTypes.AZManagementGroup]: 'fa-cube',
    [GraphNodeTypes.AZResourceGroup]: 'fa-cube',
    [GraphNodeTypes.AZServicePrincipal]: 'fa-robot',
    [GraphNodeTypes.AZSubscription]: 'fa-key',
    [GraphNodeTypes.AZTenant]: 'fa-cloud',
    [GraphNodeTypes.AZUser]: 'fa-user',
    [GraphNodeTypes.AZVM]: 'fa-desktop',
    [GraphNodeTypes.AZManagedCluster]: 'fa-cubes',
    [GraphNodeTypes.AZContainerRegistry]: 'fa-box-open',
    [GraphNodeTypes.AZWebApp]: 'fa-object-group',
    [GraphNodeTypes.AZLogicApp]: 'fa-sitemap',
    [GraphNodeTypes.AZAutomationAccount]: 'fa-cog',
    [GraphNodeTypes.Base]: 'fa-question',
    [GraphNodeTypes.Computer]: 'fa-desktop',
    [GraphNodeTypes.Domain]: 'fa-globe',
    [GraphNodeTypes.GPO]: 'fa-th-list',
    [GraphNodeTypes.Group]: 'fa-users',
    [GraphNodeTypes.OU]: 'fa-sitemap',
    [GraphNodeTypes.User]: 'fa-user',
    [GraphNodeTypes.Container]: 'fa-box',
    [GraphNodeTypes.AIACA]: 'fa-box',
    [GraphNodeTypes.RootCA]: 'fa-landmark',
    [GraphNodeTypes.EnterpriseCA]: 'fa-box',
    [GraphNodeTypes.NTAuthStore]: 'fa-store',
    [GraphNodeTypes.CertTemplate]: 'fa-id-card',
    [GraphNodeTypes.IssuancePolicy]: 'fa-clipboard-check',
};

const setFontIcons = (data: Items): void => {
    findNodes(data).forEach((node) => {
        const nodeType: GraphNodeTypes = GraphNodeTypes[node.data!.nodetype as keyof typeof GraphNodeTypes];
        node.fontIcon!.text = ICONS[nodeType];
    });
};

export {
    applyRelWidths,
    findRootId,
    findRootRelsIds,
    findTierZeroNodeId,
    getCombined,
    getLinks,
    getLinksIndex,
    getNodesIndex,
    handleLabels,
    ICONS,
    setFontIcons,
    withLinkImact,
};
