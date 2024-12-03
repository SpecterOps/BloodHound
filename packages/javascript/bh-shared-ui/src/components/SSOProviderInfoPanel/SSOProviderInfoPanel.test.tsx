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

import { OIDCProviderInfo, SAMLProviderInfo, SSOProvider } from 'js-client-library';
import { render, screen } from '../../test-utils';

import SSOProviderInfoPanel from './SSOProviderInfoPanel';

const samlProvider: SSOProvider = {
    id: 1,
    slug: 'gotham-saml',
    name: 'Gotham SAML',
    type: 'SAML',
    details: {
        idp_issuer_uri: 'http://test-idp-2:8081/metadata',
        idp_sso_uri: 'http://test-idp-2.localhost/sso',
        principal_attribute_mappings: null,
        sp_issuer_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2',
        sp_sso_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2/sso',
        sp_metadata_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2/metadata',
        sp_acs_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2/acs',
    } as SAMLProviderInfo,
    login_uri: '',
    callback_uri: '',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    config: {
        auto_provision: { enabled: false, role_provision: false, default_role: 0 },
    },
};

const oidcProvider: SSOProvider = {
    id: 1,
    slug: 'gotham-oidc',
    name: 'Gotham OIDC',
    type: 'OIDC',
    details: {
        issuer: 'http://bloodhound.localhost/test-idp-2',
        client_id: 'gotham-oidc',
    } as OIDCProviderInfo,
    login_uri: '',
    callback_uri: 'http://bloodhound.localhost/api/v2/sso/test-idp-2/callback',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    config: {
        auto_provision: { enabled: true, role_provision: true, default_role: 1 },
    },
};

describe('SSOProviderTable', () => {
    it('should render saml info provider', async () => {
        render(<SSOProviderInfoPanel ssoProvider={samlProvider} />);

        const samlInfo = samlProvider.details as SAMLProviderInfo;

        expect(await screen.findByText(samlInfo.idp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(samlInfo.sp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(samlInfo.sp_acs_uri)).toBeInTheDocument();
        expect(await screen.findByText(samlInfo.sp_metadata_uri)).toBeInTheDocument();

        expect(await screen.findByText('Automatically create new users on login')).toBeInTheDocument();
        // This provider has IDP provisioning disabled which should hide these 2 fields
        expect(screen.queryByText('Allow SSO provider to manage roles for new users')).not.toBeInTheDocument();
        expect(screen.queryByText('Default role when creating new users')).not.toBeInTheDocument();

        expect(
            screen.getByRole('button', { name: `Download ${samlProvider.name} SP Certificate` })
        ).toBeInTheDocument();
    });

    it('should render oidc info provider', async () => {
        render(<SSOProviderInfoPanel ssoProvider={oidcProvider} />);

        const oidcInfo = oidcProvider.details as OIDCProviderInfo;

        expect(await screen.findByText(oidcInfo.issuer)).toBeInTheDocument();
        expect(await screen.findByText(oidcInfo.client_id)).toBeInTheDocument();
        expect(await screen.findByText(oidcProvider.callback_uri)).toBeInTheDocument();

        expect(await screen.findByText('Automatically create new users on login')).toBeInTheDocument();
        expect(await screen.findByText('Allow SSO provider to manage roles for new users')).toBeInTheDocument();
        expect(await screen.findByText('Default role when creating new users')).toBeInTheDocument();
    });
});
