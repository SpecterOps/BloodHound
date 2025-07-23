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
    AssetGroupTagSelector,
    AssetGroupTagTypes,
    SeedTypeCypher,
    SeedTypeObjectId,
} from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { UseQueryResult } from 'react-query';
import { render, screen } from '../../../test-utils';
import DynamicDetails from './DynamicDetails';

describe('DynamicDetails', () => {
    const server = setupServer(
        rest.get(`/api/v2/asset-group-tags/*`, async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {
                        total_count: 0,
                        counts: [],
                    },
                })
            );
        }),
        rest.get(`/api/v2/asset-group-tags`, async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {
                        total_count: 0,
                        counts: [],
                    },
                })
            );
        }),
        rest.get(`/api/v2/graphs/kinds`, async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [],
                })
            );
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
        })
    );

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('renders details for a selected tier', () => {
        const testTag = {
            isLoading: false,
            isError: false,
            isSuccess: true,
            data: {
                requireCertify: true,
                created_at: '2024-09-08T03:38:22.791Z',
                created_by: 'Franz.Smitham@yahoo.com',
                deleted_at: '2025-02-03T18:32:36.669Z',
                deleted_by: 'Vita.Hermann97@yahoo.com',
                description: 'pique International',
                id: 9,
                kind_id: 59514,
                name: 'Tier-8',
                updated_at: '2024-07-26T02:15:04.556Z',
                updated_by: 'Deontae34@hotmail.com',
                position: 0,
                type: 1 as AssetGroupTagTypes,
            },
        } as unknown as UseQueryResult<AssetGroupTag | undefined>;

        render(<DynamicDetails queryResult={testTag} />);

        expect(screen.getByText('Tier-8')).toBeInTheDocument();
        expect(screen.getByText('pique International')).toBeInTheDocument();
        expect(screen.getByText('Franz.Smitham@yahoo.com')).toBeInTheDocument();
        expect(screen.getByText('2024/07/25')).toBeInTheDocument();
    });

    it('renders details for a selected selector and is of type "Cypher"', () => {
        const testSelector = {
            isLoading: false,
            isError: false,
            isSuccess: true,
            data: {
                asset_group_tag_id: 9,
                allow_disable: false,
                auto_certify: true,
                created_at: '2025-02-12T16:24:18.633Z',
                created_by: 'Emery_Swift86@gmail.com',
                description: 'North',
                disabled_at: '2024-05-24T12:34:35.894Z',
                disabled_by: 'Travon27@gmail.com',
                id: 9,
                is_default: true,
                seeds: [{ type: SeedTypeCypher, value: '1', selector_id: 9 }],
                name: 'tier-0-selector-9',
                updated_at: '2024-11-25T11:34:45.894Z',
                updated_by: 'Demario_Corwin88@yahoo.com',
            },
        } as unknown as UseQueryResult<AssetGroupTag | undefined>;

        render(<DynamicDetails queryResult={testSelector} />);

        expect(screen.getByText('tier-0-selector-9')).toBeInTheDocument();
        expect(screen.getByText('North')).toBeInTheDocument();
        expect(screen.getByText('Emery_Swift86@gmail.com')).toBeInTheDocument();
        expect(screen.getByText('2024/11/25')).toBeInTheDocument();
        expect(screen.getByText('Cypher')).toBeInTheDocument();
    });

    it('renders details for a selected selector and is of type "Object"', () => {
        const testSelectorSeedTypeObjectID = {
            isLoading: false,
            isError: false,
            isSuccess: true,
            data: {
                asset_group_tag_id: 9,
                allow_disable: false,
                id: 1,
                auto_certify: true,
                seeds: [{ type: SeedTypeObjectId, value: '1', selector_id: 1 }],
                created_at: '2025-02-12T16:24:18.633Z',
                created_by: 'Emery_Swift86@gmail.com',
                description: 'North',
                disabled_at: '2024-05-24T12:34:35.894Z',
                disabled_by: 'Travon27@gmail.com',
                is_default: true,
                name: 'tier-0-selector-9',
                updated_at: '2024-11-25T11:34:45.894Z',
                updated_by: 'Demario_Corwin88@yahoo.com',
            },
        } as unknown as UseQueryResult<AssetGroupTagSelector | undefined>;

        render(<DynamicDetails queryResult={testSelectorSeedTypeObjectID} />);

        expect(screen.getByText('tier-0-selector-9')).toBeInTheDocument();
        expect(screen.getByText('North')).toBeInTheDocument();
        expect(screen.getByText('Emery_Swift86@gmail.com')).toBeInTheDocument();
        expect(screen.getByText('2024/11/25')).toBeInTheDocument();
        expect(screen.getByText('Object ID')).toBeInTheDocument();
    });
});
