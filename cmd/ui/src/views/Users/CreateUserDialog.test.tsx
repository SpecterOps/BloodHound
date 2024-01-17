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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor } from 'src/test-utils';
import CreateUserDialog from './CreateUserDialog';

const testRoles = [
    { id: 1, name: 'Role 1' },
    { id: 2, name: 'Role 2' },
    { id: 3, name: 'Role 3' },
    { id: 4, name: 'Role 4' },
];

const testSAMLProviders = [
    { id: 1, name: 'saml-provider-1' },
    { id: 2, name: 'saml-provider-2' },
    { id: 3, name: 'saml-provider-3' },
    { id: 4, name: 'saml-provider-4' },
];

const server = setupServer(
    rest.get(`/api/v2/roles`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    roles: testRoles,
                },
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
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('CreateUserDialog', () => {
    type SetupOptions = {
        renderErrors?: boolean;
        renderLoading?: boolean;
    };

    const setup = (options?: SetupOptions) => {
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

        render(
            <CreateUserDialog
                open={true}
                onClose={testOnClose}
                onSave={testOnSave}
                isLoading={options?.renderLoading || false}
                error={options?.renderErrors}
            />
        );

        return {
            user,
            testUser,
            testOnClose,
            testOnSave,
        };
    };

    it('should render a create user form', async () => {
        setup();

        expect(screen.getByText('Create User')).toBeInTheDocument();

        expect(await screen.findByLabelText('Email Address')).toBeInTheDocument();

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

    it('should call onClose when Close button is clicked', async () => {
        const { user, testOnClose } = setup();

        const cancelButton = await screen.findByRole('button', { name: 'Cancel' });

        await user.click(cancelButton);

        expect(testOnClose).toHaveBeenCalled();
    });

    it('should not call onSave when Save button is clicked and form input is invalid', async () => {
        const { user, testOnSave } = setup();

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
        const { user, testUser, testOnSave } = setup();

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
        const { user } = setup();

        await user.click(await screen.findByLabelText('Role'));

        for (const role of testRoles) {
            expect(await screen.findByRole('option', { name: role.name })).toBeInTheDocument();
        }
    });

    it('should display all available SAML providers', async () => {
        const { user } = setup();

        await user.click(await screen.findByLabelText('Authentication Method'));

        await user.click(await screen.findByRole('option', { name: 'SAML' }));

        expect(screen.queryByLabelText('Initial Password')).not.toBeInTheDocument();

        expect(screen.queryByLabelText('Force Password Reset?')).not.toBeInTheDocument();

        expect(screen.getByLabelText('SAML Provider')).toBeInTheDocument();

        await user.click(screen.getByLabelText('SAML Provider'));

        for (const SAMLProvider of testSAMLProviders) {
            expect(await screen.findByRole('option', { name: SAMLProvider.name })).toBeInTheDocument();
        }
    });

    it('should disable Cancel and Save buttons while isLoading is true', async () => {
        setup({ renderLoading: true });

        expect(await screen.findByRole('button', { name: 'Cancel' })).toBeDisabled();

        expect(await screen.findByRole('button', { name: 'Save' })).toBeDisabled();
    });

    it('should display error message when error prop is provided', async () => {
        setup({ renderErrors: true });

        expect(await screen.findByText('An unexpected error occurred. Please try again.')).toBeInTheDocument();
    });

    it('should clear out the saml provider id from submission data when the authentication method is changed', async () => {
        const { user, testUser, testOnSave } = setup();

        const saveButton = await screen.findByRole('button', { name: 'Save' });

        await user.type(screen.getByLabelText('Email Address'), testUser.emailAddress);
        await user.type(screen.getByLabelText('Principal Name'), testUser.principalName);
        await user.type(screen.getByLabelText('First Name'), testUser.firstName);
        await user.type(screen.getByLabelText('Last Name'), testUser.lastName);

        await user.click(await screen.findByLabelText('Authentication Method'));
        await user.click(await screen.findByRole('option', { name: 'SAML' }));

        await user.click(screen.getByLabelText('SAML Provider'));
        await user.click(await screen.findByRole('option', { name: testSAMLProviders[0].name }));

        await user.click(await screen.findByLabelText('Authentication Method'));
        await user.click(await screen.findByRole('option', { name: 'Username / Password' }));
        await user.type(screen.getByLabelText('Initial Password'), testUser.password);

        await user.click(saveButton);

        expect(testOnSave).toBeCalledWith(expect.objectContaining({ SAMLProviderId: '' }));
    });
});
