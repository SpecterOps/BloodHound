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

import { AssetGroupTag, AssetGroupTagTypeTier, ConfigurationKey } from 'js-client-library';
import { UseQueryResult } from 'react-query';
import { longWait, render, screen, within } from '../../../test-utils';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { DetailsList } from './DetailsList';

const testQuery = {
    isLoading: false,
    isError: false,
    isSuccess: true,
    data: [
        {
            name: 'a',
            id: 1,
            counts: { selectors: 3, members: 2 },
            type: AssetGroupTagTypeTier,
            kind_id: 1,
            position: 1,
            requireCertify: false,
            description: '',
            created: '',
            updated: '',
            deleted: false,
            analysis_enabled: true,
        },
        {
            name: 'b',
            id: 2,
            counts: { selectors: 3, members: 2 },
            type: AssetGroupTagTypeTier,
            kind_id: 1,
            position: 1,
            requireCertify: false,
            description: '',
            created: '',
            updated: '',
            deleted: false,
            analysis_enabled: false,
        },
        {
            name: 'c',
            id: 3,
            counts: { selectors: 3, members: 2 },
            type: AssetGroupTagTypeTier,
            kind_id: 1,
            position: 1,
            requireCertify: false,
            description: '',
            created: '',
            updated: '',
            deleted: false,
            analysis_enabled: false,
        },
    ],
} as unknown as UseQueryResult<AssetGroupTag[]>;

const configResponse = {
    data: [
        {
            key: ConfigurationKey.Tiering,
            value: { multi_tier_analysis_enabled: true, tier_limit: 3, label_limit: 10 },
        },
    ],
};

const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('List', async () => {
    it('shows a loading view when data is fetching', async () => {
        const testQuery = { isLoading: true, isError: false, data: [] } as unknown as UseQueryResult<AssetGroupTag[]>;

        render(<DetailsList title='Selectors' listQuery={testQuery} selected={'1'} onSelect={() => { }} />);

        expect(screen.getAllByTestId('zone-management_selectors-list_loading-skeleton')).toHaveLength(3);
    });

    it('handles data fetching errors', async () => {
        const testQuery = { isLoading: false, isError: true, data: [] } as unknown as UseQueryResult<AssetGroupTag[]>;

        render(<DetailsList title='Selectors' listQuery={testQuery} selected={'1'} onSelect={() => { }} />);

        expect(await screen.findByText('There was an error fetching this data')).toBeInTheDocument();
    });

    it('renders a sortable list for Selectors', async () => {
        render(<DetailsList title='Selectors' listQuery={testQuery} selected={'1'} onSelect={() => { }} />);

        expect(await screen.findByText('app-icon-sort-asc')).toBeInTheDocument();
        expect(screen.queryByTestId('zone-management_details_selectors-list_static-order')).not.toBeInTheDocument();
    });

    it('renders a sortable list for Labels', async () => {
        render(<DetailsList title='Labels' listQuery={testQuery} selected={'1'} onSelect={() => { }} />);

        expect(await screen.findByText('app-icon-sort-asc')).toBeInTheDocument();
        expect(screen.queryByTestId('zone-management_details_labels-list_static-order')).not.toBeInTheDocument();
    });

    it('renders a non sortable list for Tiers', async () => {
        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => { }} />);

        expect(await screen.findByTestId('zone-management_details_tiers-list_static-order')).toBeInTheDocument();
        expect(screen.queryByText('app-icon-sort-empty')).not.toBeInTheDocument();
    });

    it('does not render tier icon tooltip when multi tier analysis is disabled', async () => {
        const configRes = {
            data: [
                {
                    key: ConfigurationKey.Tiering,
                    value: { multi_tier_analysis_enabled: false, tier_limit: 3, label_limit: 10 },
                },
            ],
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configRes));
            })
        );

        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => { }} />)

        expect(await screen.findByTestId('zone-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();

        longWait(() => {
            expect(screen.queryByTestId('analysis_disabled_icon')).not.toBeInTheDocument();
        })
    });

    it('renders tier icon tooltip when multi tier analysis is enabled but tier analysis is off', async () => {
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            }),
        );

        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => { }} />)

        expect(await screen.findByTestId('zone-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();

        longWait(() => {
            expect(screen.getByTestId('analysis_disabled_icon')).toBeInTheDocument();
        })
    });

    it('does not render tier icon tooltip when multi tier analysis is enabled and tier analysis is on', async () => {
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            }),
        );

        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => { }} />)

        const listItem1 = await screen.findByTestId('zone-management_details_tiers-list_item-1');
        expect(listItem1).toBeInTheDocument();
        expect(within(listItem1).queryByTestId('analysis_disabled_icon')).not.toBeInTheDocument();

        const listItem2 = await screen.findByTestId('zone-management_details_tiers-list_item-2');
        expect(listItem2).toBeInTheDocument();
        longWait(() => {
            expect(within(listItem2).getByTestId('analysis_disabled_icon')).toBeInTheDocument();
        })
    });

    it('handles rendering a selected item', async () => {
        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => { }} />);

        expect(await screen.findByTestId('zone-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();
    });
});
