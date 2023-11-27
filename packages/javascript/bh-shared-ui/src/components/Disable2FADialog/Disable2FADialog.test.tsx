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

import Disable2FADialog from './Disable2FADialog';

const testValidPassword = 'testValidPassword1!';

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
        return res(ctx.json({ data: testUser }));
    }),
    rest.delete(`/api/v2/bloodhound-users/${testUser.id}/mfa`, (req, res, ctx) => {
        return res();
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Enable2FADialog', () => {
    const user = userEvent.setup();
    const testOnCancel = vi.fn();
    const testOnClose = vi.fn();
    const testOnSave = vi.fn();
    const testSetSecret = vi.fn();

    beforeEach(() => {
        render(
            <Disable2FADialog
                open={true}
                onCancel={testOnCancel}
                onClose={testOnClose}
                onSave={testOnSave}
                secret=''
                onSecretChange={testSetSecret}
                contentText=''
            />
        );
    });

    it('should display "Disable Multi-Factor Authentication?" title', () => {
        expect(screen.getByText('Disable Multi-Factor Authentication?')).toBeInTheDocument();
    });

    it('should display "Password" input', () => {
        expect(screen.getByLabelText('Password')).toBeInTheDocument();
    });

    it('should display "Cancel" button', () => {
        expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
    });

    it('should display "Disable Multi-Factor Authentication" button', () => {
        expect(
            screen.getByRole('button', {
                name: 'Disable Multi-Factor Authentication',
            })
        ).toBeInTheDocument();
    });

    it('displays security warning ', () => {
        expect(screen.queryByTestId('ReportProblemOutlinedIcon')).toBeInTheDocument();
    });

    describe('user clicks "Cancel" button', () => {
        beforeEach(async () => {
            await user.click(screen.getByRole('button', { name: 'Cancel' }));
        });

        it('should call "onCancel one time"', () => {
            expect(testOnCancel).toHaveBeenCalledTimes(1);
        });
    });

    describe('user enters valid password', () => {
        beforeEach(async () => {
            await user.type(screen.getByLabelText('Password'), testValidPassword);
        });

        // TODO: it('should not display a validation error', () => {})

        describe('user clicks "Disable Multi-Factor Authentication" button', () => {
            beforeEach(async () => {
                await user.click(
                    screen.getByRole('button', {
                        name: 'Disable Multi-Factor Authentication',
                    })
                );
            });

            it('should call "onSave" one time', async () => {
                await waitFor(() => {
                    expect(testOnSave).toHaveBeenCalledTimes(1);
                });
            });
        });
    });
});
