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
import { render } from '../../../../test-utils';
import ConfirmUpdateQueryDialog from './ConfirmUpdateQueryDialog';

const testDialogContent = 'lorem ipsum';

describe('ConfirmUpdateQueryDialog', () => {
    const setup = async () => {
        const testHandleApply = vi.fn();
        const testHandleCancel = vi.fn();

        const screen = render(
            <ConfirmUpdateQueryDialog
                open={true}
                handleApply={testHandleApply}
                handleCancel={testHandleCancel}
                dialogContent={testDialogContent}
            />
        );
        const user = userEvent.setup();
        return { screen, user, testHandleApply, testHandleCancel };
    };

    it('should render with the correct content', async () => {
        const { screen } = await setup();
        expect(screen.getByText(testDialogContent)).toBeInTheDocument();
    });

    it('should handle click events on apply and cancel', async () => {
        const { screen, user, testHandleApply, testHandleCancel } = await setup();
        expect(screen.getByText(testDialogContent)).toBeInTheDocument();

        const cancel = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancel);
        expect(testHandleCancel).toBeCalled();

        const apply = screen.getByRole('button', { name: /ok/i });
        await user.click(apply);
        expect(testHandleApply).toBeCalled();
    });

    it('should not render if open === false', async () => {
        const screen = render(
            <ConfirmUpdateQueryDialog
                open={false}
                handleApply={vi.fn()}
                handleCancel={vi.fn()}
                dialogContent={testDialogContent}
            />
        );
        const testContent = screen.queryByText(testDialogContent);
        expect(testContent).toBeNull();
    });
});
