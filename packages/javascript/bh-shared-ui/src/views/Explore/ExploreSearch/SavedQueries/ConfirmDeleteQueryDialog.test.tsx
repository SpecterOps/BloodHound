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
import { vi } from 'vitest';

import userEvent from '@testing-library/user-event';
import { render } from '../../../../test-utils';
import ConfirmDeleteQueryDialog from './ConfirmDeleteQueryDialog';

describe('ConfirmDeleteQueryDialog', () => {
    const testDeleteHandler = vi.fn();
    const testClose = vi.fn();
    it('renders a confirm dialg', async () => {
        const screen = render(
            <ConfirmDeleteQueryDialog
                open={true}
                queryId={123}
                deleteHandler={testDeleteHandler}
                handleClose={testClose}
            />
        );
        expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
    });
    it('fires click handlers', async () => {
        const user = userEvent.setup();
        const screen = render(
            <ConfirmDeleteQueryDialog
                open={true}
                queryId={123}
                deleteHandler={testDeleteHandler}
                handleClose={testClose}
            />
        );
        const testCancelBtn = screen.getByRole('button', { name: /cancel/i });
        const testConfirmBtn = screen.getByRole('button', { name: /confirm/i });

        expect(testCancelBtn).toBeInTheDocument();
        expect(testConfirmBtn).toBeInTheDocument();

        await user.click(testCancelBtn);
        expect(testClose).toBeCalled();

        await user.click(testConfirmBtn);
        expect(testDeleteHandler).toBeCalled();
    });
});
