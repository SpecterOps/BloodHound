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
import { render, screen } from 'src/test-utils';

import PasswordResetForm from './PasswordResetForm';

describe('PasswordResetForm', () => {
    it('should render', () => {
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();

        render(<PasswordResetForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        expect(screen.getByText(/your account password has expired/i)).toBeInTheDocument();
        expect(screen.getByText(/please provide a new password for this account to continue/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/^password/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /reset password$/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /return to login/i })).toBeInTheDocument();
    });

    it('should call onSubmit with password when reset password button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();
        const testPassword = 'password';
        const testConfirmPassword = 'password';

        render(<PasswordResetForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        await user.type(screen.getByLabelText(/^password/i), testPassword);
        await user.type(screen.getByLabelText(/confirm password/i), testConfirmPassword);
        await user.click(screen.getByRole('button', { name: /reset password$/i }));
        expect(testOnSubmit).toHaveBeenCalledWith(testPassword);
    });

    it('should not call onSubmit when password does not match confirmation password when reset password button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();
        const testPassword = 'password';
        const testConfirmPassword = 'drowssap'; // does not match testPassword

        render(<PasswordResetForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        await user.type(screen.getByLabelText(/^password/i), testPassword);
        await user.type(screen.getByLabelText(/confirm password/i), testConfirmPassword);
        await user.click(screen.getByRole('button', { name: /reset password$/i }));
        expect(screen.getByText(/password does not match/i)).toBeInTheDocument();
        expect(testOnSubmit).not.toHaveBeenCalled();
    });

    it('should call onCancel when return to login button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();

        render(<PasswordResetForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        await user.click(screen.getByRole('button', { name: /return to login/i }));
        expect(testOnCancel).toHaveBeenCalled();
    });

    it('buttons are disabled while loading', () => {
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();

        render(<PasswordResetForm onSubmit={testOnSubmit} onCancel={testOnCancel} loading={true} />);

        expect(screen.getByRole('button', { name: /resetting password$/i })).toBeDisabled();
        expect(screen.getByRole('button', { name: /return to login/i })).toBeDisabled();
    });
});
