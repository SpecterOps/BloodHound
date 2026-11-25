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

import { useCallback, useState } from 'react';
import type { CollapsePlan, CollapseViewOptions } from '../graphCollapse';

type CollapseMode = NonNullable<CollapseViewOptions['mode']>;

export function useExploreCollapseMode(plan: CollapsePlan | undefined, updatedAt: number) {
    const [state, setState] = useState<{ plan: CollapsePlan | undefined; updatedAt: number; mode: CollapseMode }>({
        plan,
        updatedAt,
        mode: 'collapsed',
    });

    const setMode = useCallback((mode: CollapseMode) => setState({ plan, updatedAt, mode }), [plan, updatedAt]);

    if (state.plan !== plan || state.updatedAt !== updatedAt) {
        setState({ plan, updatedAt, mode: 'collapsed' });
        return ['collapsed', setMode] as const;
    }

    return [state.mode, setMode] as const;
}
