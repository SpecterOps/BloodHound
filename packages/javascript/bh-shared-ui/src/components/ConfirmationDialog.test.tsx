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
import ConfirmationDialog from './ConfirmationDialog';

describe('ConfirmationDialog', () => {
    const user = userEvent.setup();
    const testOnClose = vi.fn();
    const testOnConfirm = vi.fn();

    beforeEach(async () => {
        render(
            <ConfirmationDialog
                open={true}
                onConfirm={testOnConfirm}
                onClose={testOnClose}
                text='text-test'
                title='title-test'
            />
        );
        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
    });

    it('should display correctly', () => {
        expect(screen.queryByText('text-test')).toBeInTheDocument();
        expect(screen.queryByText('title-test')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
    });

    it('should fire Cancel once with false', async () => {
        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalledTimes(1);
        expect(testOnConfirm).toHaveBeenCalledTimes(0);
    });

    it('should fire Confirm once with true', async () => {
        await user.click(screen.getByRole('button', { name: /confirm/i }));

        expect(testOnClose).toHaveBeenCalledTimes(0);
        expect(testOnConfirm).toHaveBeenCalledTimes(1);
    });
});
