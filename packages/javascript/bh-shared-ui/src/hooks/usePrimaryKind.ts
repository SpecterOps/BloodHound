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
    const sourceKindNames = getSourceKindNames(sourceKinds);

    const kindNames = isKindNames(kinds) ? kinds : kindObjectsToKindNames(kinds);

    const sourceAndTagFilteredKinds = filterTagsAndSourceKinds(kindNames, sourceKindNames);

    const primaryKind = sourceAndTagFilteredKinds[0];

    return primaryKind ? primaryKind : kindNames[0];
};
