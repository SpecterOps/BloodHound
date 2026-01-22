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
import UsersWithEnvironmentAccessControls from '.';
import { bloodHoundUsersHandlers, testAuthenticatedUser, testBloodHoundUsers, testSSOProviders } from '../../mocks';
import { render, screen, within } from '../../test-utils';

const server = setupServer(...bloodHoundUsersHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('UsersWithEnvironmentAccessControls', () => {
    test('The password reset dialog is opened when switching a user from SAML based authentication to username/password based authentication', async () => {
        render(<UsersWithEnvironmentAccessControls />);

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
        expect(within(testUserRow).getByText(`SSO: ${testSSOProviders[0].name}`)).toBeInTheDocument();
        expect(within(testUserRow).getByRole('button')).toBeInTheDocument();

        // open the update user dialog for Marshall
        await userEvent.click(within(testUserRow).getByRole('button', { name: 'Show user actions' }));
        await screen.findByRole('menuitem', { name: /update user/i, hidden: false });
        await userEvent.click(screen.getByRole('menuitem', { name: /update user/i, hidden: false }));
        expect(await screen.findByTestId('update-user-dialog')).toBeVisible();

        // update Marshall to use username/password based authentication and save the changes
        await userEvent.click(screen.getByLabelText('Authentication Method'));
        await userEvent.click(screen.getByRole('option', { name: 'Username / Password' }));
        await userEvent.click(screen.getByRole('button', { name: 'Save' }));

        // the update user dialog should close and the password reset dialog should open
        expect(screen.queryByTestId('update-user-dialog')).toBeNull();
        expect(await screen.findByTestId('password-dialog')).toBeVisible();

        // the force password reset option should be checked
        expect(screen.getByLabelText('Force Password Reset?')).toBeChecked();
    });

    it('disables the create user button and does not populate a table if the user lacks the permission', async () => {
        render(<UsersWithEnvironmentAccessControls />);

        expect(screen.getByTestId('manage-users_button-create-user')).toBeDisabled();

        const nameElement = screen.queryByText(testBloodHoundUsers[0].principal_name);
        expect(nameElement).toBeNull();

        const rows = screen.getAllByRole('row');
        // Only the header row renders even though there is a mock endpoint that serves data
        expect(rows).toHaveLength(1);
    });

    it('does not show the "Disable MFA" context menu option for users without MFA enabled', async () => {
        render(<UsersWithEnvironmentAccessControls />);

        const noMFARow = await screen.findByRole('row', { name: /test_admin/i });

        await userEvent.click(within(noMFARow).getByRole('button', { name: 'Show user actions' }));
        await screen.findByRole('menuitem', { name: /update user/i, hidden: false });
        expect(screen.queryByRole('menuitem', { name: /disable mfa/i, hidden: false })).not.toBeInTheDocument();
    });

    it('shows the "Disable MFA" context menu option for users with MFA enabled', async () => {
        render(<UsersWithEnvironmentAccessControls />);

        const withMFARow = await screen.findByRole('row', { name: /mfa_user/i });

        await userEvent.click(within(withMFARow).getByRole('button', { name: 'Show user actions' }));
        expect(screen.queryByRole('menuitem', { name: /disable mfa/i, hidden: false })).toBeInTheDocument();
    });

    it('requires a password to disable MFA for a user when logged in without SSO', async () => {
        render(<UsersWithEnvironmentAccessControls />);

        const withMFARow = await screen.findByRole('row', { name: /mfa_user/i });

        await userEvent.click(within(withMFARow).getByRole('button', { name: 'Show user actions' }));
        await userEvent.click(screen.getByRole('menuitem', { name: /disable mfa/i }));

        const dialog = screen.queryByRole('dialog', { name: /disable multi-factor authentication/i });
        const input = screen.queryByLabelText(/password/i);

        expect(dialog).toBeInTheDocument();
        expect(input).toBeInTheDocument();
    });

    it('hides the password field and removes the requirement when logged in with SSO', async () => {
        // Override logged in admin with a SSO provider value
        server.use(
            rest.get('/api/v2/self', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: { ...testAuthenticatedUser, sso_provider_id: 1 },
                    })
                );
            })
        );
        render(<UsersWithEnvironmentAccessControls />);

        const withMFARow = await screen.findByRole('row', { name: /mfa_user/i });

        await userEvent.click(within(withMFARow).getByRole('button', { name: 'Show user actions' }));
        await userEvent.click(screen.getByRole('menuitem', { name: /disable mfa/i }));

        const dialog = screen.queryByRole('dialog', { name: /disable multi-factor authentication/i });
        const input = screen.queryByLabelText(/password/i);

        expect(dialog).toBeInTheDocument();
        expect(input).not.toBeInTheDocument();
    });
});
