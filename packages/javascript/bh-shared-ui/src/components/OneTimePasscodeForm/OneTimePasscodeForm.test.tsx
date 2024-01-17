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
import { act, render, screen } from '../../test-utils';

import OneTimePasscodeForm from './OneTimePasscodeForm';

describe('OneTimePasscodeForm', () => {
    it('should render', () => {
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();

        render(<OneTimePasscodeForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        expect(screen.getByText(/multi-factor authentication enabled/i)).toBeInTheDocument();
        expect(screen.getByText(/provide the 6 digit code from your authenticator app/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/6-digit code/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /check code/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /return to login/i })).toBeInTheDocument();
    });

    it('should call onSubmit with one time passcode when check code button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();
        const testOneTimePasscode = '123456';

        render(<OneTimePasscodeForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        await act(async () => {
            await user.type(screen.getByLabelText(/6-digit code/i), testOneTimePasscode);
            await user.click(screen.getByRole('button', { name: /check code/i }));
        });
        expect(testOnSubmit).toHaveBeenCalledWith(testOneTimePasscode);
    });

    it('should call onCancel when return to login button clicked', async () => {
        const user = userEvent.setup();
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();

        render(<OneTimePasscodeForm onSubmit={testOnSubmit} onCancel={testOnCancel} />);

        await act(async () => {
            await user.click(screen.getByRole('button', { name: /return to login/i }));
        });
        expect(testOnCancel).toHaveBeenCalled();
    });

    it('buttons are disabled while loading', () => {
        const testOnSubmit = vi.fn();
        const testOnCancel = vi.fn();

        render(<OneTimePasscodeForm onSubmit={testOnSubmit} onCancel={testOnCancel} loading={true} />);

        expect(screen.getByRole('button', { name: /checking code$/i })).toBeDisabled();
        expect(screen.getByRole('button', { name: /return to login/i })).toBeDisabled();
    });
});
