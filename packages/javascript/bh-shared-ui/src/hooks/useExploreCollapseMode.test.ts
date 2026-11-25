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

import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { createCollapseGraph, planGraphCollapse } from '../graphCollapse';
import { useExploreCollapseMode } from './useExploreCollapseMode';

describe('useExploreCollapseMode', () => {
    it('starts a different result collapsed and resets when returning to the original result', () => {
        const original = planGraphCollapse(createCollapseGraph({}), []);
        const next = planGraphCollapse(createCollapseGraph({}), []);
        const { result, rerender } = renderHook(({ plan }) => useExploreCollapseMode(plan, 100), {
            initialProps: { plan: original },
        });

        expect(result.current[0]).toBe('collapsed');
        act(() => result.current[1]('expanded'));
        rerender({ plan: original });
        expect(result.current[0]).toBe('expanded');

        rerender({ plan: next });
        expect(result.current[0]).toBe('collapsed');
        rerender({ plan: original });
        expect(result.current[0]).toBe('collapsed');
    });

    it('starts a refreshed result collapsed even when the query reuses identical data', () => {
        const plan = planGraphCollapse(createCollapseGraph({}), []);
        const { result, rerender } = renderHook(({ updatedAt }) => useExploreCollapseMode(plan, updatedAt), {
            initialProps: { updatedAt: 100 },
        });

        act(() => result.current[1]('expanded'));
        rerender({ updatedAt: 100 });
        expect(result.current[0]).toBe('expanded');

        rerender({ updatedAt: 200 });
        expect(result.current[0]).toBe('collapsed');
    });
});
