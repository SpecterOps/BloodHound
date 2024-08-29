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
import { render, screen, waitFor } from '../../test-utils';
import CitrixRDPConfiguration, { configurationData } from './CitrixRDPConfiguration';
import { dialogTitle } from './CitrixRDPConfirmDialog';

// To do: Test for initial switch value once getting configuration ( test for when it is on)
// To do: Test for checking correct dialog text when clicking to disable

describe('CitrixRDPConfiguration', () => {
    beforeEach(() => {
        render(<CitrixRDPConfiguration />);
    });

    describe('Initial render', () => {
        it('renders the component with all info and switch off', () => {
            const panelTitle = screen.getByText(configurationData.title);
            const panelDescription = screen.getByText(configurationData.description);
            const panelSwitch = screen.getByRole('switch');

            expect(panelTitle).toBeInTheDocument();
            expect(panelDescription).toBeInTheDocument();
            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).not.toBeChecked();
        });
    });

    describe('Click on switch to enable', () => {
        let panelSwitch: HTMLElement;
        let panelDialogTitle: HTMLElement;
        let panelDialogDescription: HTMLElement;
        const user = userEvent.setup();

        beforeEach(async () => {
            panelSwitch = screen.getByRole('switch');

            await user.click(panelSwitch);

            panelDialogTitle = screen.getByText(dialogTitle, { exact: false });
            panelDialogDescription = screen.getByText(/analysis has been added with citrix configuration/i);
        });

        it('on clicking switch shows modal and when clicking confirm closes it and switch stays enabled', async () => {
            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).toBeChecked();
            expect(panelDialogTitle).toBeInTheDocument();
            expect(panelDialogDescription).toBeInTheDocument();

            const confirmButton = screen.getByRole('button', { name: /confirm/i });

            await user.click(confirmButton);

            await waitFor(() => {
                expect(panelDialogTitle).not.toBeInTheDocument();
                expect(panelDialogDescription).not.toBeInTheDocument();
                expect(panelSwitch).toBeChecked();
            });
        });

        it('on clicking switch shows modal and when clicking cancel closes it and switch reverts to disabled', async () => {
            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).toBeChecked();
            expect(panelDialogTitle).toBeInTheDocument();
            expect(panelDialogDescription).toBeInTheDocument();

            const cancelButton = screen.getByRole('button', { name: /cancel/i });

            await user.click(cancelButton);

            await waitFor(() => {
                expect(panelDialogTitle).not.toBeInTheDocument();
                expect(panelDialogDescription).not.toBeInTheDocument();
                expect(panelSwitch).not.toBeChecked();
            });
        });
    });
});
