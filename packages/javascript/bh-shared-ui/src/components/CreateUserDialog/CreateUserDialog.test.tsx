// Copyright 2023 Specter Ops, Inc.
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
import { ListSSOProvidersResponse, SAMLProviderInfo, SSOProvider, SSOProviderConfiguration } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { createAuthStateWithPermissions } from '../../mocks';
import { act, render, waitFor } from '../../test-utils';
import { Permission } from '../../utils';
import CreateUserDialog from './CreateUserDialog';

const testRoles = [
    { id: 1, name: 'Role 1' },
    { id: 2, name: 'Role 2' },
    { id: 3, name: 'Role 3' },
    { id: 4, name: 'Role 4' },
];

const testSSOProviders: SSOProvider[] = [
    {
        id: 1,
        name: 'saml-provider-1',
        slug: 'saml-provider-1',
        type: 'SAML',
        login_uri: '',
        callback_uri: '',
        created_at: '',
        updated_at: '',
        details: {} as SAMLProviderInfo,
        config: {} as SSOProviderConfiguration['config'],
    },
    {
        id: 2,
        name: 'saml-provider-2',
        slug: 'saml-provider-2',
        type: 'SAML',
        login_uri: '',
        callback_uri: '',
        created_at: '',
        updated_at: '',
        details: {} as SAMLProviderInfo,
        config: {} as SSOProviderConfiguration['config'],
    },
    {
        id: 3,
        name: 'saml-provider-3',
        slug: 'saml-provider-3',
        type: 'SAML',
        login_uri: '',
        callback_uri: '',
        created_at: '',
        updated_at: '',
        details: {} as SAMLProviderInfo,
        config: {} as SSOProviderConfiguration['config'],
    },
    {
        id: 4,
        name: 'saml-provider-4',
        slug: 'saml-provider-4',
        type: 'SAML',
        login_uri: '',
        callback_uri: '',
        created_at: '',
        updated_at: '',
        details: {} as SAMLProviderInfo,
        config: {} as SSOProviderConfiguration['config'],
    },
];

const server = setupServer(
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions([Permission.AUTH_MANAGE_USERS]).user,
            })
        );
    }),
    rest.get(`/api/v2/roles`, (req, res, ctx) => {
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
                data: testSSOProviders,
            })
        );
    }),
    rest.get('/api/v2/available-domains', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('CreateUserDialog', () => {
    type SetupOptions = {
        renderErrors?: boolean;
        renderLoading?: boolean;
        renderShowEnvironmentAccessControls?: boolean;
    };

    const setup = async (options?: SetupOptions) => {
        return await act(() => {
            const user = userEvent.setup();
            const testOnClose = vi.fn();
            const testOnSave = vi.fn(() => Promise.resolve({ data: {} }));
            const testUser = {
                emailAddress: 'testuser@example.com',
                principalName: 'testuser',
                firstName: 'Test',
                lastName: 'User',
                password: 'adminAdmin1!',
                forcePasswordReset: false,
                role: testRoles[0],
            };

            const screen = render(
                <CreateUserDialog
                    error={options?.renderErrors || false}
                    isLoading={options?.renderLoading || false}
                    onClose={testOnClose}
                    onSave={testOnSave}
                    showEnvironmentAccessControls={options?.renderShowEnvironmentAccessControls || false}
                />
            );

            const openDialog = async () => await user.click(screen.getByTestId('manage-users_button-create-user'));

            return {
                screen,
                openDialog,
                user,
                testUser,
                testOnClose,
                testOnSave,
            };
        });
    };

    it('should render a create user form', async () => {
        const { screen, openDialog } = await setup();
        await openDialog();

        expect(await screen.findByText('Email Address')).toBeInTheDocument();

        expect(screen.getByLabelText('Principal Name')).toBeInTheDocument();

        expect(screen.getByLabelText('First Name')).toBeInTheDocument();

        expect(screen.getByLabelText('Last Name')).toBeInTheDocument();

        expect(screen.getByLabelText('Authentication Method')).toBeInTheDocument();

        expect(screen.getByLabelText('Initial Password')).toBeInTheDocument();

        expect(screen.getByLabelText('Force Password Reset?')).toBeInTheDocument();

        expect(screen.getByLabelText('Role')).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
    });

    it('should not call onSave when Save button is clicked and form input is invalid', async () => {
        const { screen, openDialog, user, testOnSave } = await setup();
        await openDialog();

        const saveButton = await screen.findByRole('button', { name: 'Save' });

        await user.click(saveButton);

        expect(await screen.findByText('Email Address is required')).toBeInTheDocument();

        expect(screen.getByText('Principal Name is required')).toBeInTheDocument();

        expect(screen.getByText('First Name is required')).toBeInTheDocument();

        expect(screen.getByText('Last Name is required')).toBeInTheDocument();

        expect(screen.getByText('Password is required')).toBeInTheDocument();

        expect(testOnSave).not.toHaveBeenCalled();
    });

    it('should call onSave when Save button is clicked and form input is valid', async () => {
        const { screen, openDialog, user, testOnSave, testUser } = await setup();
        await openDialog();

        const saveButton = await screen.findByRole('button', { name: 'Save' });

        await user.type(screen.getByLabelText('Email Address'), testUser.emailAddress);

        await user.type(screen.getByLabelText('Principal Name'), testUser.principalName);

        await user.type(screen.getByLabelText('First Name'), testUser.firstName);

        await user.type(screen.getByLabelText('Last Name'), testUser.lastName);

        await user.type(screen.getByLabelText('Initial Password'), testUser.password);

        await user.click(saveButton);

        await waitFor(() => expect(testOnSave).toHaveBeenCalled());
    });

    it('should display all available roles', async () => {
        const { screen, openDialog, user } = await setup();
        await openDialog();

        await user.click(await screen.findByLabelText('Role'));

        for (const role of testRoles) {
            expect(await screen.findByRole('option', { name: role.name })).toBeInTheDocument();
        }
    });

    it('should display all available SSO providers', async () => {
        const { screen, openDialog, user } = await setup();
        await openDialog();

        await user.click(await screen.findByLabelText('Authentication Method'));

        await user.click(await screen.findByRole('option', { name: 'Single Sign-On (SSO)' }));

        expect(screen.queryByLabelText('Initial Password')).not.toBeInTheDocument();

        expect(screen.queryByLabelText('Force Password Reset?')).not.toBeInTheDocument();

        expect(screen.getByLabelText('SSO Provider')).toBeInTheDocument();

        await user.click(screen.getByLabelText('SSO Provider'));

        for (const SSOProvider of testSSOProviders) {
            expect(await screen.findByRole('option', { name: SSOProvider.name })).toBeInTheDocument();
        }
    });

    it('should disable Cancel and Save buttons while isLoading is true', async () => {
        const { screen, openDialog } = await setup({ renderLoading: true });
        await openDialog();

        expect(await screen.findByRole('button', { name: 'Cancel' })).toBeDisabled();

        expect(await screen.findByRole('button', { name: 'Save' })).toBeDisabled();
    });

    it('should display error message when error prop is provided', async () => {
        const { screen, openDialog } = await setup({ renderErrors: true });
        await openDialog();

        expect(await screen.findByText('An unexpected error occurred. Please try again.')).toBeInTheDocument();
    });

    it('should clear out the SSO Provider id from submission data when the authentication method is changed', async () => {
        const { screen, openDialog, user, testUser, testOnSave } = await setup();
        await openDialog();

        const saveButton = await screen.findByRole('button', { name: 'Save' });

        await user.type(screen.getByLabelText('Email Address'), testUser.emailAddress);
        await user.type(screen.getByLabelText('Principal Name'), testUser.principalName);
        await user.type(screen.getByLabelText('First Name'), testUser.firstName);
        await user.type(screen.getByLabelText('Last Name'), testUser.lastName);

        await user.click(await screen.findByLabelText('Authentication Method'));
        await user.click(await screen.findByRole('option', { name: 'Single Sign-On (SSO)' }));

        await user.click(screen.getByLabelText('SSO Provider'));
        await user.click(await screen.findByRole('option', { name: testSSOProviders[0].name }));

        await user.click(await screen.findByLabelText('Authentication Method'));
        await user.click(await screen.findByRole('option', { name: 'Username / Password' }));
        await user.type(screen.getByLabelText('Initial Password'), testUser.password);

        await user.click(saveButton);

        expect(testOnSave).toBeCalledWith(expect.objectContaining({ sso_provider_id: undefined }));
    });
});
