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
import PasswordDialog from './PasswordDialog';
import { render, screen, waitFor } from '../../test-utils';
import { PASSWD_REQS } from '../..';

const testValidPassword = 'aA1!aaaaaaaa';

describe('PasswordDialog', () => {
    it('renders correctly', () => {
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(<PasswordDialog open={true} onSave={testOnSave} onClose={testOnClose} userId={testUserId} />);

        expect(screen.getByText('Change Password')).toBeInTheDocument();
        expect(screen.getByLabelText('Password')).toBeInTheDocument();
        expect(screen.getByLabelText('Confirmation Password')).toBeInTheDocument();
        expect(screen.queryByLabelText('Force Password Reset?')).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
    });

    it('displays Force Password Reset checkbox when showNeedsPasswordReset is true', () => {
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(
            <PasswordDialog
                open={true}
                onSave={testOnSave}
                onClose={testOnClose}
                userId={testUserId}
                showNeedsPasswordReset
            />
        );

        expect(screen.getByLabelText('Force Password Reset?')).toBeInTheDocument();
    });

    it('calls onClose when user clicks cancel', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(<PasswordDialog open={true} onSave={testOnSave} onClose={testOnClose} userId={testUserId} />);

        await user.click(screen.getByRole('button', { name: 'Cancel' }));

        expect(testOnClose).toHaveBeenCalled();
    });

    it('displays validation error messages when Save button is clicked and no form input is provided', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(<PasswordDialog open={true} onSave={testOnSave} onClose={testOnClose} userId={testUserId} />);

        await user.click(screen.getByRole('button', { name: 'Save' }));

        expect(await screen.findByText('Password Requirements')).toBeInTheDocument();

        for (const requirement of PASSWD_REQS) {
            expect(screen.getByText(requirement)).toBeInTheDocument();
        }

        expect(screen.getByText('Confirmation password is required')).toBeInTheDocument();

        expect(testOnSave).not.toHaveBeenCalled();
    });

    it('displays validation error messages when password does not meet requirements', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(<PasswordDialog open={true} onSave={testOnSave} onClose={testOnClose} userId={testUserId} />);

        await user.click(screen.getByRole('button', { name: 'Save' }));

        await user.type(screen.getByLabelText('Password'), 'aA1!');

        await user.type(screen.getByLabelText('Confirmation Password'), 'aA1!');

        await user.click(screen.getByRole('button', { name: 'Save' }));

        expect(await screen.findByText('Password Requirements')).toBeInTheDocument();

        expect(testOnSave).not.toHaveBeenCalled();
    });

    it('displays validation error messages when password does not match confirmation', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(<PasswordDialog open={true} onSave={testOnSave} onClose={testOnClose} userId={testUserId} />);

        await user.type(screen.getByLabelText('Password'), testValidPassword);

        await user.type(screen.getByLabelText('Confirmation Password'), testValidPassword + 'extracharacters');

        await user.click(screen.getByRole('button', { name: 'Save' }));

        expect(await screen.findByText('Password does not match')).toBeInTheDocument();

        expect(testOnSave).not.toHaveBeenCalled();
    });

    it('calls onSave when valid form inputs are provided', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSave = vi.fn();
        const testUserId = '1';

        render(<PasswordDialog open={true} onSave={testOnSave} onClose={testOnClose} userId={testUserId} />);

        await user.type(screen.getByLabelText('Password'), testValidPassword);

        await user.type(screen.getByLabelText('Confirmation Password'), testValidPassword);

        await user.click(screen.getByRole('button', { name: 'Save' }));

        await waitFor(() => expect(testOnSave).toHaveBeenCalled());

        expect(testOnSave).toHaveBeenCalledWith({
            needsPasswordReset: false,
            secret: testValidPassword,
            userId: testUserId,
        });
    });
});
