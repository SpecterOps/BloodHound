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

import { render, screen, within, waitFor } from '../../test-utils';
import userEvent from '@testing-library/user-event';

import UserProfile from './UserProfile';

import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
    rest.get(`/api/v2/self`, (req, res, ctx) => {
        return res();
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('UserProfile with SAML User', () => {
    const testSAMLUser = {
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
        saml_provider_id: 'test-idp-1',
    };

    beforeEach(async () => {
        server.use(
            rest.get(`/api/v2/self`, (req, res, ctx) => {
                return res(ctx.json({ data: testSAMLUser }));
            })
        );

        render(<UserProfile />);

        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
    });

    test('The reset password and two-factor authentication options should not appear in profile page if the user is configured for SAML login', () => {
        const resetPasswordButton = screen.queryByText('Reset Password');
        const twoFactorAuthToggle = screen.queryByRole('checkbox', {
            name: 'Multi-Factor Authentication Enabled',
        });

        expect(resetPasswordButton).not.toBeInTheDocument();
        expect(twoFactorAuthToggle).not.toBeInTheDocument();
    });

    test('The API key management option should appear in profile page if the user is configured for SAML login', () => {
        expect(screen.getByRole('button', { name: 'API Key Management' })).toBeInTheDocument();
    });
});

describe('UserProfile', () => {
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

    beforeEach(async () => {
        server.use(
            rest.get(`/api/v2/self`, (req, res, ctx) => {
                return res(ctx.json({ data: testUser }));
            })
        );

        render(<UserProfile />);

        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
    });

    it('should display the "My Profile" title', () => {
        expect(screen.getByText('My Profile')).toBeInTheDocument();
    });

    it('should display the "User Information" header', () => {
        expect(screen.getByText('User Information')).toBeInTheDocument();
    });

    it("should display the logged in user's email address", () => {
        expect(screen.getByText('Email')).toBeInTheDocument();
        expect(screen.getByText(testUser.email_address)).toBeInTheDocument();
    });

    it("should display the logged in user's full name", () => {
        expect(screen.getByText('Name')).toBeInTheDocument();
        expect(screen.getByText(`${testUser.first_name} ${testUser.last_name}`)).toBeInTheDocument();
    });

    it("should display the logged in user's role", () => {
        expect(screen.getByText('Email')).toBeInTheDocument();
        expect(screen.getByText(testUser.roles[0].name)).toBeInTheDocument();
    });

    it('should display the "Authentication" header', () => {
        expect(screen.getByText('Authentication')).toBeInTheDocument();
    });

    it('should display a "Reset Password" button', () => {
        expect(screen.getByRole('button', { name: 'Reset Password' })).toBeInTheDocument();
    });

    describe('"Reset Password" button is clicked', () => {
        const user = userEvent.setup();
        beforeEach(async () => {
            await user.click(screen.getByText('Reset Password'));
        });

        it('should display a "Change Password" modal', () => {
            const modal = screen.getByRole('dialog');
            expect(modal).toBeInTheDocument();
            expect(within(modal).getByText('Change Password')).toBeInTheDocument();
        });
    });

    it('should display a toggle switch to enable multi-factor authentication', () => {
        expect(
            screen.getByRole('checkbox', {
                name: 'Multi-Factor Authentication Enabled',
            })
        ).toBeInTheDocument();
    });

    describe('"Multi-Factor Authentication Enabled" switch is enabled', () => {
        const user = userEvent.setup();
        beforeEach(async () => {
            await user.click(screen.getByLabelText('Multi-Factor Authentication Enabled'));
        });

        it('should display a "Configure Multi-Factor Authentication" modal', () => {
            const modal = screen.getByRole('dialog');
            expect(modal).toBeInTheDocument();
            expect(within(modal).getByText('Configure Multi-Factor Authentication')).toBeInTheDocument();
        });
    });
});
