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

import { NodeKindRef, RelationshipKindRef, SourceKind } from 'js-client-library';
import { TagLabelPrefix } from '../constants';
import { useCustomNodeKinds } from './useCustomNodeKinds';
import { useSourceKindsQuery } from './useSourceKinds';

type KindNames = string[];

type KindObject = { name: string };
type KindObjects = KindObject[];

type EntityKinds = NodeKindRef[] | RelationshipKindRef[];

type KindList = KindNames | EntityKinds;

const isKindNames = (kinds: KindList): kinds is KindNames => {
    return !kinds[0] || typeof kinds[0] === 'string';
};

export const kindObjectsToKindNames = (kinds: KindObjects): KindNames => {
    return kinds.map((kind) => kind.name);
};

const getSourceKindNames = (sourceKinds: SourceKind[] | undefined): KindNames => {
    if (!sourceKinds) return [];

    return kindObjectsToKindNames(sourceKinds);
};

const filterTagsAndSourceKinds = (kinds: KindNames, sourceKindNames: KindNames) => {
    return kinds.filter((kind) => kind && !kind.startsWith(TagLabelPrefix) && !sourceKindNames.includes(kind));
};

export const usePrimaryKind = (kinds: KindList) => {
    const { data: sourceKinds } = useSourceKindsQuery();
    const { data: customNodeKinds } = useCustomNodeKinds();

    const sourceKindNames = getSourceKindNames(sourceKinds);

    // The frontend has no endpoint that surfaces a node kind's is_display_kind flag directly. During schema
    // reconciliation the backend creates a custom_node_kinds row for exactly those extension node kinds marked
    // is_display_kind (see upsertCustomIcons), and useCustomNodeKinds keys its result by those kind names.
    // The set of custom node kind names is therefore a faithful substitute for the set of is_display_kind kinds.
    const displayKindNames = customNodeKinds ? Object.keys(customNodeKinds) : [];

    const kindNames = isKindNames(kinds) ? kinds : kindObjectsToKindNames(kinds);

    const sourceAndTagFilteredKinds = filterTagsAndSourceKinds(kindNames, sourceKindNames);

    // A display kind (extension node kind flagged with is_display_kind) takes priority over any
    // non-display kind regardless of its position in the kinds array.
    const displayKind = sourceAndTagFilteredKinds.find((kind) => displayKindNames.includes(kind));

    const primaryKind = displayKind ?? sourceAndTagFilteredKinds[0];

    return primaryKind ? primaryKind : kindNames[0];
};
