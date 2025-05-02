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

import { AssetGroupTagTypeTier } from 'js-client-library';
import { UseQueryResult } from 'react-query';
import { render, screen } from '../../../test-utils';
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
        },
    ],
} as unknown as UseQueryResult<{ name: string; id: number; counts: { members: number } }[], unknown>;

describe('List', async () => {
    it('shows a loading view when data is fetching', async () => {
        const testQuery = { isLoading: true, isError: false, data: [] } as unknown as UseQueryResult<
            { name: string; id: number; counts: { members: number } }[],
            unknown
        >;

        render(<DetailsList title='Selectors' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(screen.getAllByTestId('tier-management_details_selectors-list_loading-skeleton')).toHaveLength(3);
    });

    it('handles data fetching errors', async () => {
        const testQuery = { isLoading: false, isError: true, data: [] } as unknown as UseQueryResult<
            { name: string; id: number; counts: { members: number } }[],
            unknown
        >;

        render(<DetailsList title='Selectors' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByText('There was an error fetching this data')).toBeInTheDocument();
    });

    it('renders a sortable list for Selectors', async () => {
        render(<DetailsList title='Selectors' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByText('app-icon-sort-asc')).toBeInTheDocument();
        expect(screen.queryByTestId('tier-management_details_selectors-list_static-order')).not.toBeInTheDocument();
    });

    it('renders a sortable list for Labels', async () => {
        render(<DetailsList title='Labels' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByText('app-icon-sort-asc')).toBeInTheDocument();
        expect(screen.queryByTestId('tier-management_details_labels-list_static-order')).not.toBeInTheDocument();
    });

    it('renders a non sortable list for Tiers', async () => {
        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByTestId('tier-management_details_tiers-list_static-order')).toBeInTheDocument();
        expect(screen.queryByText('app-icon-sort-empty')).not.toBeInTheDocument();
    });

    it('handles rendering a selected item', async () => {
        render(<DetailsList title='Tiers' listQuery={testQuery} selected={'1'} onSelect={() => {}} />);

        expect(await screen.findByTestId('tier-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();
    });
});
