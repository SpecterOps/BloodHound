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
import MakeSSOConfiguration from './SSOConfiguration';

const initialSAMLProvider: CreateSAMLProviderResponse = {
    id: 1,
    name: 'test-idp-1',
    display_name: 'Test IDP 1',
    idp_issuer_uri: 'http://test-idp-1:8081/metadata',
    idp_sso_uri: 'http://test-idp-1.localhost/sso',
    principal_attribute_mappings: null,
    sp_issuer_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1',
    sp_sso_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/sso',
    sp_metadata_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/metadata',
    sp_acs_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-1/acs',
    created_at: '2022-02-24T23:38:41.420271Z',
    updated_at: '2022-02-24T23:38:41.420271Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
};

const samlProviders = [initialSAMLProvider];
let autoId = 1;

const getNewSAMLProvider = (): CreateSAMLProviderResponse => {
    const now = new Date().toISOString();

    const newSAMLProvider: CreateSAMLProviderResponse = {
        id: ++autoId,
        name: 'test-idp-2',
        display_name: 'Test IDP 2',
        idp_issuer_uri: 'http://test-idp-2:8081/metadata',
        idp_sso_uri: 'http://test-idp-2.localhost/sso',
        principal_attribute_mappings: null,
        sp_issuer_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2',
        sp_sso_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2/sso',
        sp_metadata_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2/metadata',
        sp_acs_uri: 'http://bloodhound.localhost/api/v2/login/saml/test-idp-2/acs',
        created_at: now,
        updated_at: now,
        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    };
    return newSAMLProvider;
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
    deleted_at: {
        Time: string;
        Valid: boolean;
    };
}

const server = setupServer(
    rest.get<{ saml_providers: CreateSAMLProviderResponse[] }>('/api/v2/saml', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    saml_providers: samlProviders,
                },
            })
        );
    }),

    rest.post<CreateSAMLProviderBody, any, CreateSAMLProviderResponse>('/api/v2/saml/providers', (req, res, ctx) => {
        const newSAMLProvider = getNewSAMLProvider();
        samlProviders.push(newSAMLProvider);
        return res(ctx.json(newSAMLProvider));
    })
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
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const mockUseFeatureFlag = vi.fn();

vi.mock('../../hooks/useFeatureFlags', () => {
    return {
        useFeatureFlag: (flagKey: string) => mockUseFeatureFlag(flagKey),
    };
});

describe('SSOConfiguration', async () => {
    const addSnackbar = vi.fn();
    const useAppDispatch = vi.fn();
    const SSOConfiguration = MakeSSOConfiguration(addSnackbar, useAppDispatch);

    it('should eventually render previously configured SSO providers', async () => {
        const user = userEvent.setup();

        render(<SSOConfiguration />);
        expect(await screen.findByText('SSO Configuration')).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLProvider.name)).toBeInTheDocument();
        expect(await screen.findByText('SAML')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'test-idp-1' }));

        expect(await screen.findByText(initialSAMLProvider.idp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLProvider.sp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLProvider.sp_acs_uri)).toBeInTheDocument();
        expect(await screen.findByText(initialSAMLProvider.sp_metadata_uri)).toBeInTheDocument();
    });

    it('should allow user to create new SAML provder and then display it in the table', async () => {
        const user = userEvent.setup();
        const newSAMLProvider = getNewSAMLProvider();
        const newSAMLProviderRequest: CreateSAMLProviderBody = {
            name: newSAMLProvider.name,
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

        expect(await screen.findByText(newSAMLProvider.idp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(newSAMLProvider.sp_sso_uri)).toBeInTheDocument();
        expect(await screen.findByText(newSAMLProvider.sp_acs_uri)).toBeInTheDocument();
        expect(await screen.findByText(newSAMLProvider.sp_metadata_uri)).toBeInTheDocument();
    });
});
