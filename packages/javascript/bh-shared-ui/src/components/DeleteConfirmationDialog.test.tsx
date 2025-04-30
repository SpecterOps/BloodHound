// Copyright 2025 Specter Ops, Inc.
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
import { render, screen, waitFor } from '../test-utils';
import DeleteConfirmationDialog from './DeleteConfirmationDialog';

describe('DeleteConfirmationDialog', () => {
    const user = userEvent.setup();
    const testOnClose = vi.fn();
    const testOnConfirm = vi.fn();

    beforeEach(async () => {
        render(
            <DeleteConfirmationDialog
                open={true}
                onConfirm={testOnConfirm}
                onClose={testOnClose}
                itemName='test-item'
                itemType='test-type'
            />
        );
        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
    });

    it('should display correctly', () => {
        expect(screen.queryByText(/delete test-item\?/i)).toBeInTheDocument();
        expect(
            screen.queryByText(
                /continuing onwards will delete test-item and all associated configurations and findings\./i
            )
        ).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /confirm/i })).toBeDisabled();
        expect(screen.getByTestId('confirmation-dialog_challenge-text')).toBeInTheDocument();
    });

    it('should fire Cancel once with false', async () => {
        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalledWith(false);
        expect(testOnClose).toHaveBeenCalledTimes(1);
    });

    it('should fire Confirm once with true after typing challenge text', async () => {
        await user.type(screen.getByTestId('confirmation-dialog_challenge-text'), 'delete this test-type');
        expect(screen.getByRole('button', { name: /confirm/i })).not.toBeDisabled();
        await user.click(screen.getByRole('button', { name: /confirm/i }));

        expect(testOnClose).toHaveBeenCalledWith(true);
        expect(testOnClose).toHaveBeenCalledTimes(1);
    });
});
