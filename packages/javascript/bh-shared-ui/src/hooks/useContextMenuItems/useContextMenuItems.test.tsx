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

import { type UseQueryResult } from 'react-query';
import { renderHook } from '../../test-utils';
import { type MousePosition } from '../../types';
import { Permission } from '../../utils';
import * as epHook from '../useExploreParams';
import * as esiHook from '../useExploreSelectedItem';
import { type ItemResponse } from '../useGraphItem';
import * as pHook from '../usePermissions';
import * as cmiHook from './useContextMenuItems';

const position: MousePosition = { mouseX: 0, mouseY: 0 };

const mockUseExploreSelectedItem = vi.spyOn(esiHook, 'useExploreSelectedItem');
const mockUseExploreParams = vi.spyOn(epHook, 'useExploreParams');
const mockUsePermissions = vi.spyOn(pHook, 'usePermissions');

const nodeQuery = {
    data: {
        objectId: 'abc',
    },
} as UseQueryResult<ItemResponse, unknown>;

const edgeQuery = {
    data: {
        id: '123_MemberOf_456',
        source: 'edge_source',
    },
} as UseQueryResult<ItemResponse, unknown>;

const setup = async ({
    exploreSearchTab = '',
    permission = false,
    selectedItemQuery,
}: {
    selectedItemQuery: UseQueryResult<ItemResponse, unknown>;
    exploreSearchTab?: string;
    permission?: boolean;
}) => {
    mockUseExploreSelectedItem.mockReturnValue({ selectedItemQuery } as any);
    mockUseExploreParams.mockReturnValue({ exploreSearchTab } as any);
    mockUsePermissions.mockReturnValue({ checkPermission: (p: Permission) => !!p && permission } as any);
};

describe('useContextMenuItems', () => {
    it('tests if selected item is a node', async () => {
        await setup({ selectedItemQuery: nodeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(position));
        const { asEdgeItem, asNodeItem, selectedItemQuery } = result.current;

        expect(asEdgeItem(selectedItemQuery)).toBeUndefined();
        expect(asNodeItem(selectedItemQuery)).toBe(nodeQuery.data);
    });

    it('tests if selected item is an edge on pathfinding tab', async () => {
        await setup({ exploreSearchTab: 'pathfinding', selectedItemQuery: edgeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(position));
        const { asEdgeItem, asNodeItem, selectedItemQuery } = result.current;

        expect(asNodeItem(selectedItemQuery)).toBeUndefined();
        expect(asEdgeItem(selectedItemQuery)).toBe(edgeQuery.data);
    });

    it('tests if selected item is an edge not on pathfinding tab', async () => {
        await setup({ exploreSearchTab: 'node', selectedItemQuery: edgeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(position));
        const { asEdgeItem, asNodeItem, selectedItemQuery } = result.current;

        expect(asNodeItem(selectedItemQuery)).toBeUndefined();
        expect(asEdgeItem(selectedItemQuery)).toBeUndefined();
    });

    it('calculates the menu position', async () => {
        await setup({ exploreSearchTab: 'node', selectedItemQuery: edgeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(position));
        const { menuPosition } = result.current;

        expect(menuPosition).toEqual({ left: 56, top: 0 });
    });

    it('returns a null menu position', async () => {
        await setup({ exploreSearchTab: 'node', selectedItemQuery: edgeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(null));
        const { menuPosition } = result.current;

        expect(menuPosition).toBe(null);
    });

    it('returns permissions check for asset group allowed', async () => {
        await setup({ permission: true, selectedItemQuery: nodeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(position));
        const { isAssetGroupEnabled } = result.current;

        expect(isAssetGroupEnabled).toBeTruthy();
    });

    it('returns permissions check for asset group not allowed', async () => {
        await setup({ permission: false, selectedItemQuery: nodeQuery });
        const { result } = renderHook(() => cmiHook.useContextMenuItems(position));
        const { isAssetGroupEnabled } = result.current;

        expect(isAssetGroupEnabled).toBeFalsy();
    });
});
