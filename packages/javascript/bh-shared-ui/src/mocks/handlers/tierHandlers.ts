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
    rest.get('/api/v2/asset-group-labels/:assetGroupId/members', async (_req, res, ctx) => {
        return res(ctx.json({ data: { members: tierMocks.createSelectorNodes(10, 0) } }));
    }),

    // GET Members/Objects for Selector
    rest.get('/api/v2/asset-group-labels/:assetGroupId/selectors/:selectorId/members', async (req, res, ctx) => {
        const { selectorId } = req.params;
        return res(ctx.json({ data: { members: tierMocks.createSelectorNodes(10, parseInt(selectorId as string)) } }));
    }),
];

export default tierHandlers;
