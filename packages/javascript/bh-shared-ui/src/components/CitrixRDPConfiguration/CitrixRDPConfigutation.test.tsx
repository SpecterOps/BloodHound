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
import { render, screen } from '../../test-utils';
import CitrixRDPConfiguration, { configurationData } from './CitrixRDPConfiguration';
import { dialogTitle } from './CitrixRDPConfirmDialog';

describe('CitrixRDPConfiguration', () => {
    beforeEach(() => {
        render(<CitrixRDPConfiguration />);
    });

    it('should render the component with all info and switch off', () => {
        const panelTitle = screen.getByText(configurationData.title);
        const panelDescription = screen.getByText(configurationData.description);
        const panelSwitchLabel = screen.getByLabelText(/off/i);
        const panelSwitch = screen.getByRole('switch');

        expect(panelTitle).toBeInTheDocument();
        expect(panelDescription).toBeInTheDocument();
        expect(panelSwitch).toBeInTheDocument();
        expect(panelSwitchLabel).toBeInTheDocument();
        expect(panelSwitch).not.toBeChecked();
    });
    it('when clicking on switch to on its shows modal and when clicking on confirm stays on', async () => {
        const panelSwitch = screen.getByRole('switch');
        const user = userEvent.setup();

        await user.click(panelSwitch);

        const panelDialogTitle = screen.getByText(dialogTitle, { exact: false });
        const panelDialogDescription = screen.getByText(/Analysis has been added with Citrix Configuration/i);

        expect(panelSwitch).toBeChecked();
        expect(panelDialogTitle).toBeInTheDocument();
        expect(panelDialogDescription).toBeInTheDocument();
    });
    it('when clicking on switch to on its shows modal and when clicking on cancel it returns to off', async () => {
        const panelSwitch = screen.getByRole('switch');
        const user = userEvent.setup();

        await user.click(panelSwitch);

        const panelDialogTitle = screen.getByText(dialogTitle, { exact: false });
        const panelDialogDescription = screen.getByText(/Analysis has been removed with Citrix Configuration/i);

        expect(panelSwitch).not.toBeChecked();
        expect(panelDialogTitle).toBeInTheDocument();
        expect(panelDialogDescription).toBeInTheDocument();
    });
});
