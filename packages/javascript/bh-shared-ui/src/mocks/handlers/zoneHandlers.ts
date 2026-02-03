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

import { ConfigurationKey } from 'js-client-library';
import { rest } from 'msw';
import * as tierMocks from '../factories/privilegeZones';

const zoneHandlers = [
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
                    {
                        id: 17,
                        key: 'dark_mode',
                        name: 'Dark Mode',
                        description: 'Best mode 😎',
                        enabled: true,
                        user_updatable: false,
                    },
                ],
            })
        );
    }),
    // GET Kinds
    rest.get('/api/v2/graphs/kinds', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {},
            })
        );
    }),

    rest.get('/api/v2/available-domains', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),

    rest.get(`/api/v2/custom-nodes`, async (_req, res, ctx) => {
        return res(ctx.json({ data: [] }));
    }),

    rest.get('/api/v2/users/*', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    props: {},
                },
            })
        );
    }),

    rest.get('/api/v2/config', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: ConfigurationKey.Tiering,
                        value: { multi_tier_analysis_enabled: true, tier_limit: 3, label_limit: 10 },
                    },
                ],
            })
        );
    }),

    // GET Tags
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tags: tierMocks.createAssetGroupTags(5) } }));
    }),

    // GET Tag
    rest.get('/api/v2/asset-group-tags/:tagId', async (req, res, ctx) => {
        const { tagId } = req.params;

        return res(ctx.json({ data: { tag: tierMocks.createAssetGroupTag(parseInt(tagId as string)) } }));
    }),

    // POST Tag
    rest.post('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tag: { id: 777 } } }));
    }),

    // PATCH Tag
    rest.patch('/api/v2/asset-group-tags/:tagId', async (_, res, ctx) => {
        return res(ctx.json({ data: { tag: { id: 777 } } }));
    }),

    // DELETE Tag
    rest.delete('/api/v2/asset-group-tags/:tagId', async (_, res, ctx) => {
        return res(ctx.status(500, 'get rekt'));
    }),

    // GET Selectors
    rest.get('/api/v2/asset-group-tags/:tagId/selectors', async (req, res, ctx) => {
        const { tagId } = req.params;
        return res(ctx.json({ data: { selectors: tierMocks.createRules(10, parseInt(tagId as string)) } }));
    }),

    // GET Selector
    rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (req, res, ctx) => {
        const { tagId, selectorId } = req.params;
        return res(
            ctx.json({
                data: { selector: tierMocks.createRule(parseInt(tagId as string), parseInt(selectorId as string)) },
            })
        );
    }),

    // POST Selector
    rest.post('/api/v2/asset-group-tags/:tagId/selectors', async (_, res, ctx) => {
        return res(ctx.status(200));
    }),

    // PATCH Selector
    rest.patch('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (_, res, ctx) => {
        return res(ctx.status(200));
    }),

    // DELETE Selector
    rest.delete('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', async (_, res, ctx) => {
        return res(ctx.status(500, 'get rekt'));
    }),

    // GET Members/Objects for Tag
    rest.get('/api/v2/asset-group-tags/:tagId/members', async (req, res, ctx) => {
        const total = 3000;
        const url = new URL(req.url);
        const { tagId, selectorId } = req.params;
        const skip = url.searchParams.get('skip');
        const limit = url.searchParams.get('limit');

        return res(
            ctx.json({
                data: {
                    members: tierMocks.createObjects(
                        parseInt(tagId as string),
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
    rest.get('/api/v2/asset-group-tags/:tagId/selectors/:selectorId/members*', async (req, res, ctx) => {
        const total = 2000;
        const { tagId, selectorId } = req.params;
        const url = new URL(req.url);
        const skip = url.searchParams.get('skip');
        const limit = url.searchParams.get('limit');
        return res(
            ctx.json({
                data: {
                    members: tierMocks.createObjects(
                        parseInt(tagId as string),
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

    // GET Member counts
    rest.get('/api/v2/asset-group-tags/:tagId/members/counts', async (_, res, ctx) => {
        return res(ctx.json({ data: tierMocks.createAssetGroupMembersCount() }));
    }),

    // GET Selectors for Object/Member
    rest.get('/api/v2/asset-group-tags/:tagId/members/:memberId', async (req, res, ctx) => {
        const { tagId, memberId } = req.params;

        return res(
            ctx.json({
                data: {
                    member: tierMocks.createAssetGroupMemberInfo(tagId as string, memberId as string),
                },
            })
        );
    }),
    rest.get('/api/v2/self', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {},
            })
        );
    }),
    rest.get(`/api/v2/roles`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    roles: [],
                },
            })
        );
    }),
];

export default zoneHandlers;
