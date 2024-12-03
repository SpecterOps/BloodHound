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
import { OIDCProviderInfo, SAMLProviderInfo, SSOProvider, SSOProviderConfiguration } from 'js-client-library';
import { render, screen } from '../../test-utils';
import { SortOrder } from '../../utils';
import SSOProviderTable from './SSOProviderTable';

const samlProvider: SSOProvider = {
    id: 1,
    slug: 'gotham-saml',
    name: 'Gotham SAML',
    type: 'SAML',
    login_uri: '',
    callback_uri: '',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    details: {} as SAMLProviderInfo,
    config: {} as SSOProviderConfiguration['config'],
};

const oidcProvider: SSOProvider = {
    id: 2,
    slug: 'gotham-oidc',
    name: 'Gotham OIDC',
    type: 'OIDC',
    login_uri: '',
    callback_uri: '',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    details: {} as OIDCProviderInfo,
    config: {} as SSOProviderConfiguration['config'],
};

const ssoProviders = [samlProvider, oidcProvider];

describe('SSOProviderTable', () => {
    const onClickSSOProvider = vi.fn();
    const onDeleteSSOProvider = vi.fn();
    const onUpdateSSOProvider = vi.fn();

    it('should render', async () => {
        const onToggleTypeSortOrder = vi.fn();

        render(
            <SSOProviderTable
                ssoProviders={ssoProviders}
                loading={false}
                onClickSSOProvider={onClickSSOProvider}
                onDeleteSSOProvider={onDeleteSSOProvider}
                onUpdateSSOProvider={onUpdateSSOProvider}
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
        let typeSortOrder: SortOrder;
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
                onUpdateSSOProvider={onUpdateSSOProvider}
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
