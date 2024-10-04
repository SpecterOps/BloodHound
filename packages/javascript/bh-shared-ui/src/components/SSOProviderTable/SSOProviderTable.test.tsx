// Copyright 2024 Specter Ops, Inc.
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

import userEvent from '@testing-library/user-event';
import { render, screen } from '../../test-utils';

import SSOProviderTable from './SSOProviderTable';

interface ListSSOProvidersResponse {
    id: number;
    slug: string;
    type: string;
    name: string;
    created_at: string;
    updated_at: string;
    deleted_at: {
        Time: string;
        Valid: boolean;
    };
}

const samlProvider: ListSSOProvidersResponse = {
    id: 1,
    slug: 'gotham-saml',
    name: 'Gotham SAML',
    type: 'SAML',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
};

const oidcProvider: ListSSOProvidersResponse = {
    id: 1,
    slug: 'gotham-oidc',
    name: 'Gotham OIDC',
    type: 'OIDC',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
};

const ssoProviders = [samlProvider, oidcProvider];

describe('SSOProviderTable', () => {
    const onClickSSOProvider = vi.fn();
    const onDeleteSSOProvider = vi.fn();

    it('should render', async () => {
        const onToggleTypeSortOrder = vi.fn();

        render(
            <SSOProviderTable
                ssoProviders={ssoProviders}
                loading={false}
                onClickSSOProvider={onClickSSOProvider}
                onDeleteSSOProvider={onDeleteSSOProvider}
                onToggleTypeSortOrder={onToggleTypeSortOrder}
            />
        );

        expect(await screen.findByText(samlProvider.name)).toBeInTheDocument();
        expect(await screen.findByText('SAML')).toBeInTheDocument();
        expect(await screen.findByText(oidcProvider.name)).toBeInTheDocument();
        expect(await screen.findByText('OIDC')).toBeInTheDocument();
    });

    it('should sort by type', async () => {
        const user = userEvent.setup();
        var typeSortOrder: 'asc' | 'desc' | undefined;
        const onToggleTypeSortOrder = () => {
            if (!typeSortOrder || typeSortOrder === 'desc') {
                typeSortOrder = 'asc';
            } else {
                typeSortOrder = 'desc';
            }
        };

        render(
            <SSOProviderTable
                ssoProviders={ssoProviders}
                loading={false}
                onClickSSOProvider={onClickSSOProvider}
                onDeleteSSOProvider={onDeleteSSOProvider}
                onToggleTypeSortOrder={onToggleTypeSortOrder}
                typeSortOrder={typeSortOrder}
            />
        );

        await user.click(screen.getByRole('button', { name: 'Type' }));
        expect(typeSortOrder).toBe('asc');

        await user.click(screen.getByRole('button', { name: 'Type' }));
        expect(typeSortOrder).toBe('desc');
    });
});
