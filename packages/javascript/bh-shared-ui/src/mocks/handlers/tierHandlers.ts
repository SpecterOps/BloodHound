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

import { DefaultBodyType, MockedRequest, rest, RestHandler } from 'msw';
import * as tierMocks from '../factories/tierManagement';

const tierHandlers: RestHandler<MockedRequest<DefaultBodyType>>[] = [
    rest.get('/api/v2/features', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        id: 17,
                        key: 'tier_management_engine',
                        name: 'Tier Management Engine',
                        description: 'Updates the managed assets selector engine and the asset management page.',
                        enabled: true,
                        user_updatable: true,
                    },
                ],
            })
        );
    }),

    // GET Labels
    rest.get('/api/v2/asset-group-labels', async (_req, res, ctx) => {
        return res(ctx.json({ data: { asset_group_labels: tierMocks.createAssetGroupLabels() } }));
    }),

    // GET Selectors
    rest.get('/api/v2/asset-group-labels/:assetGroupId/selectors', async (req, res, ctx) => {
        const { assetGroupId } = req.params;
        return res(ctx.json({ data: { selectors: tierMocks.createSelectors(10, parseInt(assetGroupId as string)) } }));
    }),

    // GET Members/Objects for Label
    rest.get('/api/v2/asset-group-labels/:assetGroupId/members', async (req, res, ctx) => {
        const total = 3000;
        const url = new URL(req.url);
        const { assetGroupId, selectorId } = req.params;
        const skip = url.searchParams.get('skip');
        const limit = url.searchParams.get('limit');

        return res(
            ctx.json({
                data: {
                    members: tierMocks.createSelectorNodes(
                        parseInt(assetGroupId as string),
                        parseInt(selectorId as string),
                        parseInt(skip as string),
                        parseInt(limit as string),
                        total
                    ),
                },
                skip: skip,
                limit: limit,
                count: total,
            })
        );
    }),

    // GET Members/Objects for Selector
    rest.get('/api/v2/asset-group-labels/:assetGroupId/selectors/:selectorId/members*', async (req, res, ctx) => {
        const total = 2000;
        const { assetGroupId, selectorId } = req.params;
        const url = new URL(req.url);
        const skip = url.searchParams.get('skip');
        const limit = url.searchParams.get('limit');
        return res(
            ctx.json({
                data: {
                    members: tierMocks.createSelectorNodes(
                        parseInt(assetGroupId as string),
                        parseInt(selectorId as string),
                        parseInt(skip as string),
                        parseInt(limit as string),
                        total
                    ),
                },
                skip: skip,
                limit: limit,
                count: total,
            })
        );
    }),

    // GET Selectors for Object/Member
    rest.get('/api/v2/asset-group-labels/:assetGroupId/members/:memberId', async (req, res, ctx) => {
        const total = 1057;
        const { assetGroupId, memberId } = req.params;
        const url = new URL(req.url);
        const skip = url.searchParams.get('skip');
        const limit = url.searchParams.get('limit');

        return res(
            ctx.json({
                data: {
                    member: tierMocks.createAssetGroupMemberInfo(
                        parseInt(assetGroupId as string),
                        parseInt(memberId as string)
                    ),
                },
                skip: skip,
                limit: limit,
                count: total,
            })
        );
    }),

    // GET object counts
    rest.get('/api/v2/asset-group-labels/:assetGroupId/members/counts', async (req, res, ctx) => {
        const { assetGroupId } = req.params;
        return res(
            ctx.json({ data: { counts: tierMocks.createAssetGroupMembersCount(parseInt(assetGroupId as string)) } })
        );
    }),
];

export default tierHandlers;
