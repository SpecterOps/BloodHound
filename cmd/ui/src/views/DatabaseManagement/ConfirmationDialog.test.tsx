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

import { render, screen } from 'src/test-utils';
import ConfirmationDialog from './ConfirmationDialog';
import userEvent from '@testing-library/user-event';

describe('Confirmation Dialog', () => {
    const handleClose = vi.fn();
    const handleDelete = vi.fn();

    beforeEach(() => {
        render(<ConfirmationDialog open handleClose={handleClose} handleDelete={handleDelete} />);
    });

    afterEach(() => {
        vi.resetAllMocks();
    });

    it('renders', () => {
        const title = screen.getByRole('heading', { name: /confirm deleting data/i });
        const textField = screen.getByRole('textbox');

        expect(title).toBeInTheDocument();
        expect(textField).toBeInTheDocument();
    });

    it('displays an error message when user tries to submit without typing the message', async () => {
        const user = userEvent.setup();

        const confirmButton = screen.getByRole('button', { name: /confirm/i });
        await user.click(confirmButton);

        const errMessage = screen.getByText(/please input the phrase prior to clicking confirm/i);
        expect(errMessage).toBeInTheDocument();
    });

    it('removes the error message when user tries fully types the message', async () => {
        const user = userEvent.setup();

        // user types some of the phrase
        const textField = screen.getByRole('textbox');
        await user.type(textField, 'please');

        const confirmButton = screen.getByRole('button', { name: /confirm/i });
        await user.click(confirmButton);

        const errMessage = screen.getByText(/please input the phrase prior to clicking confirm/i);
        expect(errMessage).toBeInTheDocument();

        // user types all of the phrase
        await user.type(textField, ' delete my data');
        expect(errMessage).not.toBeInTheDocument();
    });

    it('handles a submission when the user has typed the phrase', async () => {
        const user = userEvent.setup();

        const textField = screen.getByRole('textbox');
        await user.type(textField, 'please delete my data');

        const confirmButton = screen.getByRole('button', { name: /confirm/i });
        await user.click(confirmButton);

        expect(handleClose).toHaveBeenCalledTimes(1);
        expect(handleDelete).toHaveBeenCalledTimes(1);
    });
});
