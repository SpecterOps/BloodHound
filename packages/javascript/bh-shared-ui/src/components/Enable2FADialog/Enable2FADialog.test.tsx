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

import { render, screen, waitFor } from '../../test-utils';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';

import Enable2FADialog from './Enable2FADialog';

const testValidPassword = 'testValidPassword1!';
const testValidOtp = '123456';

const testUser = {
    id: 100,
    first_name: 'Test',
    last_name: 'User',
    email_address: 'testuser@example.com',
    roles: [
        {
            id: 1,
            name: 'Test Role',
        },
    ],
    saml_provider_id: null,
};

const server = setupServer(
    rest.get(`/api/v2/self`, (req, res, ctx) => {
        return res(ctx.json({}));
    }),
    rest.post(`/api/v2/bloodhound-users/${testUser.id}/mfa`, (req, res, ctx) => {
        return res(
            ctx.json({
                qr_code: '',
                totp_secret: '',
            })
        );
    }),
    rest.post(`/api/v2/bloodhound-users/${testUser.id}/mfa-activation`, (req, res, ctx) => {
        return res(
            ctx.json({
                status: 'activated',
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Enable2FADialog', () => {
    it('should display "Configure Multi-Factor Authentication" title', () => {
        const testOnCancel = vi.fn();
        const testOnClose = vi.fn();
        const testOnSavePassword = vi.fn(async () => {});
        const testOnSaveOTP = vi.fn(async () => {});
        const testOnSave = vi.fn();

        render(
            <Enable2FADialog
                open={true}
                onCancel={testOnCancel}
                onClose={testOnClose}
                onSavePassword={testOnSavePassword}
                onSaveOTP={testOnSaveOTP}
                onSave={testOnSave}
                TOTPSecret='12345'
                QRCode='12345'
            />
        );

        expect(screen.getByText('Configure Multi-Factor Authentication')).toBeInTheDocument();
        expect(screen.getAllByText('Password')).toHaveLength(2);
        expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Next' })).toBeInTheDocument();
    });

    it('onCancel is called when cancel is clicked', async () => {
        const user = userEvent.setup();
        const testOnCancel = vi.fn();
        const testOnClose = vi.fn();
        const testOnSavePassword = vi.fn(async () => {});
        const testOnSaveOTP = vi.fn(async () => {});
        const testOnSave = vi.fn();

        render(
            <Enable2FADialog
                open={true}
                onCancel={testOnCancel}
                onClose={testOnClose}
                onSavePassword={testOnSavePassword}
                onSaveOTP={testOnSaveOTP}
                onSave={testOnSave}
                TOTPSecret='12345'
                QRCode='12345'
            />
        );

        // click cancel
        await user.click(screen.getByRole('button', { name: 'Cancel' }));

        expect(testOnCancel).toHaveBeenCalledTimes(1);
    });

    it('should allow user to configure 2fa', async () => {
        const user = userEvent.setup();
        const testOnCancel = vi.fn();
        const testOnClose = vi.fn();
        const testOnSavePassword = vi.fn(async () => {});
        const testOnSaveOTP = vi.fn(async () => {});
        const testOnSave = vi.fn();

        render(
            <Enable2FADialog
                open={true}
                onCancel={testOnCancel}
                onClose={testOnClose}
                onSavePassword={testOnSavePassword}
                onSaveOTP={testOnSaveOTP}
                onSave={testOnSave}
                TOTPSecret='12345'
                QRCode='12345'
            />
        );

        // type password
        await user.type(screen.getByTestId('enable-2fa-dialog_input-password'), testValidPassword);

        // click next
        await user.click(screen.getByRole('button', { name: 'Next' }));

        await waitFor(() => {
            expect(screen.getByAltText('QR Code for Configuring Multi-Factor Authentication')).toBeInTheDocument();
        });
        expect(screen.getAllByText('One-Time Password')).toHaveLength(2);

        // type otp
        await user.type(screen.getByTestId('enable-2fa-dialog_input-one-time-password'), testValidOtp);

        // click next
        await user.click(screen.getByRole('button', { name: 'Next' }));

        expect(
            await screen.findByText("Next time you log in, you'll need to use your password and authentication code.")
        ).toBeInTheDocument();
    });
});
