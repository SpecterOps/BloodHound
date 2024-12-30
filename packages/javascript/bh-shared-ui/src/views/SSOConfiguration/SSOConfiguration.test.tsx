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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../test-utils';
import SSOConfiguration from './SSOConfiguration';
import { ListRolesResponse, ListSSOProvidersResponse, Role, SAMLProviderInfo, SSOProvider } from 'js-client-library';

const testRoles = [
    { id: 1, name: 'Read-Only' },
    { id: 2, name: 'Power User' },
    { id: 3, name: 'Administrator' },
    { id: 4, name: 'Upload Only' },
] as Role[];

const initialSAMLProvider: SSOProvider = {
    id: 1,
    type: 'SAML',
    slug: 'test-idp-1',
    name: 'Test IDP 1',
    details: {
        idp_issuer_uri: 'http://test-idp-1:8081/metadata',
        idp_sso_uri: 'http://test-idp-1.localhost/sso',
        principal_attribute_mappings: null,
        sp_issuer_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1',
        sp_sso_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/sso',
        sp_metadata_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/metadata',
        sp_acs_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/acs',
    } as SAMLProviderInfo,
    login_uri: '',
    callback_uri: '',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    config: {
        auto_provision: { enabled: false, role_provision: false, default_role: 1 },
    },
};

const ssoProviders = [initialSAMLProvider];

const newSAMLProvider: SSOProvider = {
    id: 2,
    type: 'SAML',
    slug: 'test-idp-2',
    name: 'Test IDP 2',
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
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    config: {
        auto_provision: { enabled: false, role_provision: false, default_role: 1 },
    },
};

interface CreateSAMLProviderBody {
    name: string;
    metadata: File;
}

interface CreateSAMLProviderResponse {
    id: number;
    name: string;
    display_name: string;
    idp_issuer_uri: string;
    idp_sso_uri: string;
    principal_attribute_mappings: string[] | null;
    sp_issuer_uri: string;
    sp_sso_uri: string;
    sp_metadata_uri: string;
    sp_acs_uri: string;
    created_at: string;
    updated_at: string;
}

const server = setupServer(
    rest.get<any, any, ListRolesResponse>(`/api/v2/roles`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    roles: testRoles,
                },
            })
        );
    }),
    rest.get<any, any, ListSSOProvidersResponse>('/api/v2/sso-providers', (req, res, ctx) => {
        return res(
            ctx.json({
                data: ssoProviders,
            })
        );
    }),

    rest.post<CreateSAMLProviderBody, any, CreateSAMLProviderResponse>(
        '/api/v2/sso-providers/saml',
        (req, res, ctx) => {
            ssoProviders.push(newSAMLProvider);
            return res(ctx.json({ ...newSAMLProvider, ...(newSAMLProvider.details as SAMLProviderInfo) }));
        }
    )
);

beforeEach(() => {
    mockUseFeatureFlag.mockImplementation((flagKey: string) => {
        return {
            data: {
                key: flagKey,
                enabled: true, // flag enabled
            },
            isLoading: false,
            isError: false,
            error: null,
        };
    });
});
beforeAll(() => {
    server.listen();
});
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const mockUseFeatureFlag = vi.fn();

vi.mock('../../hooks/useFeatureFlags', () => {
    return {
        useFeatureFlag: (flagKey: string) => mockUseFeatureFlag(flagKey),
    };
});

describe('SSOConfiguration', async () => {
    it('should eventually render previously configured SSO providers', async () => {
        const user = userEvent.setup();

        render(<SSOConfiguration />);
        expect(await screen.findByText('SSO Configuration')).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLProvider.name)).toBeInTheDocument();
        expect(await screen.findByText('SAML')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: initialSAMLProvider.name }));

        const initialSAMLDetails = initialSAMLProvider.details as SAMLProviderInfo;

        expect(await screen.findByText(initialSAMLDetails.idp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLDetails.sp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLDetails.sp_acs_uri)).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLDetails.sp_metadata_uri)).toBeInTheDocument();
    });

    it('should allow user to create new SAML provder and then display it in the table', async () => {
        const user = userEvent.setup();
        const newSAMLProviderRequest: CreateSAMLProviderBody = {
            name: newSAMLProvider.slug,
            metadata: new File([], 'new-saml-provider.xml'),
        };

        render(<SSOConfiguration />);

        await user.click(screen.getByRole('button', { name: /create provider/i }));

        expect(await screen.findByRole('menu')).toBeInTheDocument();

        await user.click(screen.getByRole('menuitem', { name: /saml provider/i }));

        await user.type(screen.getByLabelText('SAML Provider Name'), newSAMLProviderRequest.name);
        await user.upload(screen.getByLabelText('Choose File'), newSAMLProviderRequest.metadata);

        await user.click(screen.getByRole('button', { name: 'Submit' }));

        expect(await screen.findByText(newSAMLProvider.name)).toBeInTheDocument();

        await user.click(await screen.findByText(newSAMLProvider.name));

        const newSAMLDetails = newSAMLProvider.details as SAMLProviderInfo;

        expect(await screen.findByText(newSAMLDetails.idp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(newSAMLDetails.sp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(newSAMLDetails.sp_acs_uri)).toBeInTheDocument();
        expect(await screen.findByText(newSAMLDetails.sp_metadata_uri)).toBeInTheDocument();
    });
});
