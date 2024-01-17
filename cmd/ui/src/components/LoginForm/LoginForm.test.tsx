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

import LoginForm from './LoginForm';

describe('LoginForm', () => {
    it('should render', () => {
        const testOnSubmit = vi.fn();
        const testOnLoginViaSAML = vi.fn();

        render(<LoginForm onSubmit={testOnSubmit} onLoginViaSAML={testOnLoginViaSAML} />);

        expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /login$/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /login via sso/i })).toBeInTheDocument();
    });

    it('should call onSubmit with email address and password when login button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnLoginViaSAML = vi.fn();
        const testEmailAddress = 'user@example.com';
        const testPassword = 'password';

        render(<LoginForm onSubmit={testOnSubmit} onLoginViaSAML={testOnLoginViaSAML} />);

        await user.type(screen.getByLabelText(/email address/i), testEmailAddress);
        await user.type(screen.getByLabelText(/password/i), testPassword);
        await user.click(screen.getByRole('button', { name: /login$/i }));
        expect(testOnSubmit).toHaveBeenCalledWith(testEmailAddress, testPassword);
    });

    it('should call onLoginViaSAML when login via sso button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnLoginViaSAML = vi.fn();

        render(<LoginForm onSubmit={testOnSubmit} onLoginViaSAML={testOnLoginViaSAML} />);

        await user.click(screen.getByRole('button', { name: /login via sso/i }));
        expect(testOnLoginViaSAML).toHaveBeenCalled();
    });

    it('buttons are disabled while loading', () => {
        const testOnSubmit = vi.fn();
        const testOnLoginViaSAML = vi.fn();

        render(<LoginForm onSubmit={testOnSubmit} onLoginViaSAML={testOnLoginViaSAML} loading={true} />);

        expect(screen.getByRole('button', { name: /logging in$/i })).toBeDisabled();
        expect(screen.getByRole('button', { name: /login via sso/i })).toBeDisabled();
    });
});
