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

import { AssetGroupTag, AssetGroupTagTypeLabel, AssetGroupTagTypeZone, ConfigurationKey } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { UseQueryResult } from 'react-query';
import { useParams } from 'react-router-dom';
import zoneHandlers from '../../../mocks/handlers/zoneHandlers';
import { detailsPath, privilegeZonesPath, zonesPath } from '../../../routes';
import { render, screen, within } from '../../../test-utils';
import { TagList } from './TagList';

const testQuery = {
    isLoading: false,
    isError: false,
    isSuccess: true,
    data: [
        {
            name: 'a',
            id: 1,
            counts: { selectors: 3, members: 2 },
            type: AssetGroupTagTypeZone,
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
            type: AssetGroupTagTypeZone,
            kind_id: 1,
            position: 2,
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
            type: AssetGroupTagTypeLabel,
            kind_id: 1,
            position: null,
            requireCertify: false,
            description: '',
            created: '',
            updated: '',
            deleted: false,
            analysis_enabled: false,
        },
    ],
} as unknown as UseQueryResult<AssetGroupTag[]>;

const configTrueResponse = {
    data: [
        {
            key: ConfigurationKey.Tiering,
            value: { multi_tier_analysis_enabled: true, tier_limit: 3, label_limit: 10 },
        },
    ],
};

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useParams: vi.fn(),
    };
});

describe('List', async () => {
    vi.mocked(useParams).mockReturnValue({ zoneId: '3', labelId: undefined });

    it('shows a loading view when data is fetching', async () => {
        const testQuery = { isLoading: true, isError: false, data: undefined } as unknown as UseQueryResult<
            AssetGroupTag[]
        >;

        render(<TagList title='Labels' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(screen.getAllByTestId('privilege-zones_labels-list_loading-skeleton')).toHaveLength(3);
    });

    it('handles data fetching errors', async () => {
        const testQuery = { isLoading: false, isError: true, data: [] } as unknown as UseQueryResult<AssetGroupTag[]>;

        render(<TagList title='Labels' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByText('There was an error fetching this data')).toBeInTheDocument();
    });

    it('renders a sortable list for Labels', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '3' });
        render(<TagList title='Labels' listQuery={testQuery} selected={'3'} onSelect={() => {}} />, {
            route: '/privilege-zones/labels/details',
        });

        expect(await screen.findByText('app-icon-sort-asc')).toBeInTheDocument();
        expect(screen.queryByTestId('privilege-zones_details_labels-list_static-order')).not.toBeInTheDocument();
    });

    it('renders a non sortable list for Zones', async () => {
        render(<TagList title='Zones' listQuery={testQuery} selected={'2'} onSelect={() => {}} />, {
            route: `/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}`,
        });

        expect(await screen.findByTestId('privilege-zones_details_zones-list_static-order')).toBeInTheDocument();
        expect(screen.queryByText('app-icon-sort-empty')).not.toBeInTheDocument();
    });

    it('does not render tier icon tooltip when multi tier analysis is disabled', async () => {
        const configFalseResponse = {
            data: [
                {
                    key: ConfigurationKey.Tiering,
                    value: { multi_tier_analysis_enabled: false, tier_limit: 3, label_limit: 10 },
                },
            ],
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configFalseResponse));
            })
        );

        render(<TagList title='Zones' listQuery={testQuery} selected={'2'} onSelect={() => {}} />, {
            route: '/privilege-zones/zones/details',
        });

        const listItem = await screen.findByTestId('privilege-zones_details_zones-list_item-2');
        expect(listItem).toBeInTheDocument();

        const icon = within(listItem).queryByTestId('analysis_disabled_icon');
        expect(icon).not.toBeInTheDocument();
    });

    it('renders tier icon tooltip when multi tier analysis is enabled but tier analysis is off', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '2', labelId: undefined });
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configTrueResponse));
            })
        );

        render(<TagList title='Zones' listQuery={testQuery} selected={'2'} onSelect={() => {}} />, {
            route: '/privilege-zones/zones/2/details',
        });

        const listItem = screen.getByTestId('privilege-zones_details_zones-list_item-2');
        expect(listItem).toBeInTheDocument();

        const icon = await within(listItem).findByTestId('analysis_disabled_icon');
        expect(icon).toBeInTheDocument();
    });

    it('does not render tier icon tooltip when multi tier analysis is enabled and tier analysis is on', async () => {
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configTrueResponse));
            })
        );

        render(<TagList title='Zones' listQuery={testQuery} selected={'2'} onSelect={() => {}} />, {
            route: '/privilege-zones/zones/2/details',
        });

        const listItem1 = screen.getByTestId('privilege-zones_details_zones-list_item-1');
        expect(listItem1).toBeInTheDocument();
        expect(await within(listItem1).queryByTestId('analysis_disabled_icon')).not.toBeInTheDocument();

        const listItem2 = screen.getByTestId('privilege-zones_details_zones-list_item-2');
        expect(listItem2).toBeInTheDocument();

        expect(await within(listItem2).findByTestId('analysis_disabled_icon')).toBeInTheDocument();
    });

    it('handles rendering a selected item', async () => {
        render(<TagList title='Zones' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByTestId('privilege-zones_details_zones-list_active-zones-item-1')).toBeInTheDocument();
    });
});
