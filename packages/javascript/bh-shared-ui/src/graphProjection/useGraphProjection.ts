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

import { useEffect, useMemo, useState } from 'react';
import { buildProjection } from './buildProjection';
import { effectivelyExpandedGroups } from './provenance';
import type { GroupingRule, SourceGraph, SourceItemRef } from './types';
import { validateGroups } from './validateGroups';

type UseGraphProjectionOptions = {
    source?: SourceGraph;
    queryKey: unknown;
    rules: readonly GroupingRule[];
    selectedSourceItem?: SourceItemRef;
};

type ExpansionState = {
    queryIdentity: string | undefined;
    expanded: ReadonlySet<string>;
};

export function useGraphProjection({ source, queryKey, rules, selectedSourceItem }: UseGraphProjectionOptions) {
    const queryIdentity = JSON.stringify(queryKey);
    const validation = useMemo(
        () =>
            source
                ? validateGroups(
                      source,
                      rules.flatMap((rule) => rule.discover(source))
                  )
                : undefined,
        [source, rules]
    );
    const groups = useMemo(() => (validation?.valid ? validation.groups : []), [validation]);
    const [state, setState] = useState<ExpansionState>({ queryIdentity, expanded: new Set() });
    const availableGroupIds = useMemo(() => new Set(groups.map((group) => group.id)), [groups]);
    const expandedGroupIds = useMemo(
        () =>
            state.queryIdentity === queryIdentity
                ? new Set([...state.expanded].filter((id) => availableGroupIds.has(id)))
                : new Set<string>(),
        [state, queryIdentity, availableGroupIds]
    );

    // Reset during render so a new query never briefly shows the previous query's expansion state.
    // Also prune group IDs that no longer exist in the current source.
    if (state.queryIdentity !== queryIdentity || (source && expandedGroupIds.size !== state.expanded.size)) {
        setState({ queryIdentity, expanded: expandedGroupIds });
    }

    useEffect(() => {
        if (validation && !validation.valid) {
            console.error('Invalid graph projection groups', validation.diagnostics);
        }
    }, [validation]);

    const selectedKind = selectedSourceItem?.kind;
    const selectedId = selectedSourceItem?.id;
    const revealed = useMemo<readonly SourceItemRef[]>(
        () => (selectedKind && selectedId !== undefined ? [{ kind: selectedKind, id: selectedId }] : []),
        [selectedKind, selectedId]
    );

    const graph = useMemo(
        () => (source ? buildProjection(source, groups, { expandedGroupIds }, revealed) : undefined),
        [source, groups, expandedGroupIds, revealed]
    );
    const effectivelyExpandedGroupIds = source
        ? effectivelyExpandedGroups(source, groups, expandedGroupIds, revealed)
        : new Set<string>();

    return {
        graph,
        groups,
        canExpand: groups.some((group) => !effectivelyExpandedGroupIds.has(group.id)),
        canCollapse: groups.some((group) => effectivelyExpandedGroupIds.has(group.id)),
        expandGroup: (id: string) => {
            if (availableGroupIds.has(id)) {
                setState({ queryIdentity, expanded: new Set([...expandedGroupIds, id]) });
            }
        },
        expandAll: () => setState({ queryIdentity, expanded: availableGroupIds }),
        collapseAll: () => setState({ queryIdentity, expanded: new Set() }),
    };
}
