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
import { ListSSOProvidersResponse, SAMLProviderInfo, SSOProvider, SSOProviderConfiguration } from 'js-client-library';
import { DefaultBodyType, MockedRequest, RestHandler, rest } from 'msw';

export const testAuthenticatedUser = {
    sso_provider_id: null,
    AuthSecret: {
        digest_method: 'argon2',
        expires_at: '2025-01-01T12:00:00Z',
        totp_activated: false,
        id: 31,
        created_at: '2024-01-01T12:00:00Z',
        updated_at: '2024-01-01T12:00:00Z',
        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    },
    roles: [
        {
            name: 'Administrator',
            description: 'Administrator',
            permissions: [
                {
                    authority: 'auth',
                    name: 'ManageUsers',
                },
            ],
            id: 4,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
    ],
    first_name: 'Test',
    last_name: 'Admin',
    email_address: 'test_admin@specterops.io',
    principal_name: 'test_admin',
    last_login: '0001-01-01T00:00:00Z',
    is_disabled: false,
    eula_accepted: true,
    id: '0',
    created_at: '2024-01-01T12:00:00Z',
    updated_at: '2024-01-01T12:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
};

export const testMarshallLaw = {
    sso_provider_id: 1,
    AuthSecret: {
        digest_method: 'argon2',
        expires_at: '2025-01-01T12:00:00Z',
        totp_activated: false,
        id: 31,
        created_at: '2024-01-01T12:00:00Z',
        updated_at: '2024-01-01T12:00:00Z',
        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    },
    roles: [
        {
            name: 'User',
            description: 'User',
            permissions: [
                {
                    authority: 'auth',
                    name: 'ManageSelf',
                },
            ],
            id: 4,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
    ],
    first_name: 'Marshall',
    last_name: 'Law',
    email_address: 'mlaw@specterops.io',
    principal_name: 'mlaw',
    last_login: '0001-01-01T00:00:00Z',
    is_disabled: false,
    eula_accepted: true,
    id: '1',
    created_at: '2024-01-01T12:00:00Z',
    updated_at: '2024-01-01T12:00:00Z',
    deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
};

const mfaUser = {
    sso_provider_id: null,
    AuthSecret: {
        digest_method: 'argon2',
        expires_at: '2026-01-08T22:00:00.893744Z',
        totp_activated: true,
        id: 2,
        created_at: '2025-10-10T22:00:00.896644Z',
        updated_at: '2025-10-10T22:35:42.15299Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
    roles: [
        {
            name: 'User',
            description: 'User',
            permissions: [
                {
                    authority: 'auth',
                    name: 'ManageSelf',
                },
            ],
            id: 4,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
    ],
    first_name: 'mfa',
    last_name: 'user',
    email_address: 'mfa@user.com',
    principal_name: 'mfa_user',
    last_login: '2025-10-10T22:34:54.189205Z',
    is_disabled: false,
    all_environments: true,
    environment_access_control: [],
    eula_accepted: true,
    id: '2',
    created_at: '2025-10-10T22:00:00.896184Z',
    updated_at: '2025-10-10T22:00:00.896184Z',
    deleted_at: {
        Time: '0001-01-01T00:00:00Z',
        Valid: false,
    },
};

export const testBloodHoundUsers = [testAuthenticatedUser, testMarshallLaw, mfaUser];

export const testSSOProviders: SSOProvider[] = [
    {
        name: 'saml-provider',
        slug: 'saml-provider',
        type: 'SAML',
        login_uri: '',
        callback_uri: '',
        id: 1,
        created_at: '2024-01-01T12:00:00Z',
        updated_at: '2024-01-01T12:00:00Z',
        details: {} as SAMLProviderInfo,
        config: {} as SSOProviderConfiguration['config'],
    },
];

export const testRoles = {
    roles: [
        {
            name: 'User',
            description: 'User',
            permissions: [
                {
                    authority: 'auth',
                    name: 'ManageSelf',
                },
            ],
            id: 3,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
        {
            name: 'Administrator',
            description: 'Administrator',
            permissions: [
                {
                    authority: 'auth',
                    name: 'ManageUsers',
                },
            ],
            id: 4,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
    ],
};

export const bloodHoundUsersHandlers: RestHandler<MockedRequest<DefaultBodyType>>[] = [
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: testAuthenticatedUser,
            })
        );
    }),
    rest.get('/api/v2/bloodhound-users', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    users: testBloodHoundUsers,
                },
            })
        );
    }),
    rest.get('/api/v2/bloodhound-users/1', (req, res, ctx) => {
        return res(
            ctx.json({
                data: testMarshallLaw,
            })
        );
    }),
    rest.get<any, any, ListSSOProvidersResponse>('/api/v2/sso-providers', (req, res, ctx) => {
        return res(
            ctx.json({
                data: testSSOProviders,
            })
        );
    }),
    rest.get('/api/v2/roles', (req, res, ctx) => {
        return res(ctx.json({ data: testRoles }));
    }),
    rest.patch('/api/v2/bloodhound-users/1', (req, res, ctx) => {
        return res(ctx.json({ data: { ...testMarshallLaw, sso_provider_id: null, AuthSecret: null } }));
    }),
];
