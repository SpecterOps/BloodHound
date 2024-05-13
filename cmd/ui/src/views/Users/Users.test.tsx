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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, within } from 'src/test-utils';
import Users from '.';
import userEvent from '@testing-library/user-event';

const testAuthenticatedUser = {
    saml_provider_id: null,
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
            permissions: [],
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

const testMarshallLaw = {
    saml_provider_id: 1,
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
            permissions: [],
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

const testBloodHoundUsers = [testAuthenticatedUser, testMarshallLaw];

const testSAMLProviders = [
    {
        name: 'saml-provider',
        display_name: 'saml-provider',
        idp_issuer_uri: 'urn:saml-provider.com',
        idp_sso_uri: 'https://saml-provider.com/saml',
        principal_attribute_mappings: null,
        sp_issuer_uri: 'https://test.bloodhoundenterprise.io/api/v1/login/saml/saml-provider',
        sp_sso_uri: 'https://test.bloodhoundenterprise.io/api/v1/login/saml/saml-provider/sso',
        sp_metadata_uri: 'https://test.bloodhoundenterprise.io/api/v1/login/saml/saml-provider/metadata',
        sp_acs_uri: 'https://test.bloodhoundenterprise.io/api/v1/login/saml/saml-provider/acs',
        id: 1,
        created_at: '2024-01-01T12:00:00Z',
        updated_at: '2024-01-01T12:00:00Z',
        deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
    },
];

const testRoles = {
    roles: [
        {
            name: 'User',
            description: 'User',
            permissions: [],
            id: 3,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
        {
            name: 'Administrator',
            description: 'Administrator',
            permissions: [],
            id: 4,
            created_at: '2024-01-01T12:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
            deleted_at: { Time: '0001-01-01T00:00:00Z', Valid: false },
        },
    ],
};

const server = setupServer(
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
    rest.get('/api/v2/saml', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    saml_providers: testSAMLProviders,
                },
            })
        );
    }),
    rest.get('/api/v2/roles', (req, res, ctx) => {
        return res(ctx.json({ data: testRoles }));
    }),
    rest.patch('/api/v2/bloodhound-users/1', (req, res, ctx) => {
        return res(ctx.json({ data: { ...testMarshallLaw, saml_provider_id: null, AuthSecret: null } }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Users', () => {
    it('The password reset dialog is opened when switching a user from SAML based authentication to username/password based authentication', async () => {
        render(<Users />);

        expect(screen.getByText('Manage Users')).toBeInTheDocument();

        // wait for the table data to load
        await screen.findByText(testBloodHoundUsers[0].principal_name);

        // this table row contains the data for "Marshall Law" aka testBloodHoundUsers[1]
        const testUserRow = screen.getAllByRole('row')[2];

        expect(within(testUserRow).getByText(testBloodHoundUsers[1].principal_name)).toBeInTheDocument();
        expect(within(testUserRow).getByText(testBloodHoundUsers[1].email_address)).toBeInTheDocument();
        expect(
            within(testUserRow).getByText(`${testBloodHoundUsers[1].first_name} ${testBloodHoundUsers[1].last_name}`)
        ).toBeInTheDocument();
        expect(within(testUserRow).getByText('2024-01-01 04:00 PST (GMT-0800)')).toBeInTheDocument();
        expect(within(testUserRow).getByText('User')).toBeInTheDocument();
        expect(within(testUserRow).getByText('Active')).toBeInTheDocument();
        expect(within(testUserRow).getByText(`SAML: ${testSAMLProviders[0].name}`)).toBeInTheDocument();
        expect(within(testUserRow).getByRole('button')).toBeInTheDocument();

        // open the update user dialog for Marshall
        await userEvent.click(within(testUserRow).getByRole('button', { name: 'bars' }));
        await screen.findByRole('menuitem', { name: /update user/i, hidden: false });
        await userEvent.click(screen.getByRole('menuitem', { name: /update user/i, hidden: false }));
        expect(await screen.findByTestId('update-user-dialog')).toBeVisible();

        // update Marshall to use username/password based authentication and save the changes
        await userEvent.click(screen.getByLabelText('Authentication Method'));
        await userEvent.click(screen.getByRole('option', { name: 'Username / Password' }));
        await userEvent.click(screen.getByRole('button', { name: 'Save' }));

        // the update user dialog should close and the password reset dialog should open
        expect(await screen.findByTestId('update-user-dialog')).not.toBeVisible();
        expect(await screen.findByTestId('password-dialog')).toBeVisible();

        // the force password reset option should be checked
        expect(screen.getByLabelText('Force Password Reset?')).toBeChecked();
    });
});
