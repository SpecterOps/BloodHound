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
import {
    AssetGroupTag,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypeZone,
} from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../../test-utils';
import { SortOrder } from '../../types';
import { apiClient } from '../../utils';
import * as agtHook from './useAssetGroupTags';

const handlers = [
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    tags: [
                        { position: 1, id: 42, type: AssetGroupTagTypeZone },
                        { position: 2, id: 23, type: AssetGroupTagTypeZone },
                        { position: 7, id: 1, type: AssetGroupTagTypeZone },
                        { position: 3, id: 2, type: AssetGroupTagTypeZone },
                        { position: 777, id: 3, type: AssetGroupTagTypeZone },
                    ],
                },
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/:tagId/members', async (_, res, ctx) => {
        return res(ctx.status(200));
    }),

    rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId/members', async (_, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get('/api/v2/features', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'tier_management_engine',
                        enabled: true,
                    },
                ],
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('the useAssetGroupTags utilities', () => {
    it('enables returning a list of tags', async () => {
        const { result } = renderHook(() => agtHook.useTagsQuery());

        await waitFor(() => {
            expect(result.current.data).toHaveLength(5);
        });

        expect(result.current.data[0].position).toBe(1);
        expect(result.current.data[1].position).toBe(2);
        expect(result.current.data[2].position).toBe(7);
        expect(result.current.data[3].position).toBe(3);
        expect(result.current.data[4].position).toBe(777);
    });

    it('enables pulling an ordered list of tags by position ascending', async () => {
        const { result } = renderHook(() => agtHook.useOrderedTags());

        await waitFor(() => {
            expect(result.current.orderedTags).toHaveLength(5);
        });

        expect(result.current.orderedTags[0].position).toBe(1);
        expect(result.current.orderedTags[1].position).toBe(2);
        expect(result.current.orderedTags[2].position).toBe(3);
        expect(result.current.orderedTags[3].position).toBe(7);
        expect(result.current.orderedTags[4].position).toBe(777);
    });

    it('enables correctly returning the tag associated with Tier Zero (position value of 1) from the list of tags', async () => {
        server.use(
            rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            tags: [
                                { position: 2, id: 23, type: AssetGroupTagTypeZone },
                                { position: 7, id: 1, type: AssetGroupTagTypeZone },
                                { position: 3, id: 2, type: AssetGroupTagTypeZone },
                                { position: 777, id: 3, type: AssetGroupTagTypeZone },
                                { position: 1, id: 42, type: AssetGroupTagTypeZone },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => agtHook.useHighestPrivilegeTag());

        await waitFor(() => {
            expect(result.current.tag.position).toBe(1);
        });
    });

    it('enables filtering out for tags that are treated as Labels (including Owned)', async () => {
        server.use(
            rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            tags: [
                                { position: 1, id: 42, type: AssetGroupTagTypeZone },
                                { position: null, id: 2, type: AssetGroupTagTypeLabel },
                                { position: null, id: 3, type: AssetGroupTagTypeLabel },
                                { position: null, id: 4, type: AssetGroupTagTypeOwned },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => agtHook.useLabels());

        await waitFor(() => {
            expect(result.current).toHaveLength(3);
        });

        expect(result.current.filter((tag: AssetGroupTag) => tag.position !== null)).toHaveLength(0);
    });

    test('tag members refetches on sort change', async () => {
        const tagMembersSpy = vi.spyOn(apiClient, 'getAssetGroupTagMembers');

        const { rerender } = renderHook((sortOrder: SortOrder) => agtHook.useTagMembersInfiniteQuery(1, sortOrder));

        await waitFor(() => {
            expect(tagMembersSpy).toHaveBeenCalledTimes(1);
        });

        rerender('desc');

        await waitFor(() => {
            expect(tagMembersSpy).toHaveBeenCalledTimes(2);
        });
    });

    test('tag members refetches on environment change', async () => {
        const tagMembersSpy = vi.spyOn(apiClient, 'getAssetGroupTagMembers');

        const { rerender } = renderHook((environments: string[]) =>
            agtHook.useTagMembersInfiniteQuery(1, 'asc', environments)
        );

        await waitFor(() => {
            expect(tagMembersSpy).toHaveBeenCalledTimes(1);
        });

        rerender(['1']);

        await waitFor(() => {
            expect(tagMembersSpy).toHaveBeenCalledTimes(2);
        });
    });

    test('selector members refetches on sort change', async () => {
        const selectorMembersSpy = vi.spyOn(apiClient, 'getAssetGroupTagSelectorMembers');

        const { rerender } = renderHook((sortOrder: SortOrder) => agtHook.useRuleMembersInfiniteQuery(1, 1, sortOrder));

        await waitFor(() => {
            expect(selectorMembersSpy).toHaveBeenCalledTimes(1);
        });

        rerender('desc');

        await waitFor(() => {
            expect(selectorMembersSpy).toHaveBeenCalledTimes(2);
        });
    });

    test('selector members refetches on envrionment change', async () => {
        const selectorMembersSpy = vi.spyOn(apiClient, 'getAssetGroupTagSelectorMembers');

        const { rerender } = renderHook((environments: string[]) =>
            agtHook.useRuleMembersInfiniteQuery(1, 1, 'asc', environments)
        );

        await waitFor(() => {
            expect(selectorMembersSpy).toHaveBeenCalledTimes(1);
        });

        rerender(['1']);

        await waitFor(() => {
            expect(selectorMembersSpy).toHaveBeenCalledTimes(2);
        });
    });
});
