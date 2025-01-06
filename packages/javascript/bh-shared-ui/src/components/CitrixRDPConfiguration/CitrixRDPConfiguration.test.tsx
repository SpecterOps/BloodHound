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

import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../../test-utils';
import CitrixRDPConfiguration, { configurationData } from './CitrixRDPConfiguration';
import { dialogTitle } from './CitrixRDPConfirmDialog';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { ConfigurationKey } from 'js-client-library';

describe('CitrixRDPConfiguration', () => {
    const setInitialServerState = (savedConfigurationValue?: boolean) => {
        return {
            isCitrixRDPConfigurationEnabled: savedConfigurationValue || false,
        };
    };

    let serverState = setInitialServerState();

    const server = setupServer(
        rest.get(`/api/v2/config`, async (_req, res, ctx) => {
            return res(
                ctx.json({
                    data: [
                        {
                            key: ConfigurationKey.Citrix,
                            value: {
                                enabled: serverState.isCitrixRDPConfigurationEnabled,
                            },
                        },
                    ],
                })
            );
        }),
        rest.put(`/api/v2/config`, async (req, res, ctx) => {
            const body = await req.json();
            serverState.isCitrixRDPConfigurationEnabled = !!body['value']['enabled'];
            return res(ctx.json({ data: body }));
        })
    );
    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    describe('Initial render', () => {
        beforeEach(() => {
            serverState = setInitialServerState();
            render(<CitrixRDPConfiguration />);
        });
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
        beforeEach(() => {
            serverState = setInitialServerState();
            render(<CitrixRDPConfiguration />);
        });
        it('on clicking switch shows modal and when clicking cancel closes it and switch stays disabled', async () => {
            const panelSwitch = screen.getByRole('switch');
            const user = userEvent.setup();

            await user.click(panelSwitch);

            const panelDialogTitle = await screen.findByText(dialogTitle, { exact: false });
            const panelDialogDescription = screen.getByText(/analysis has been added with citrix configuration/i);

            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).not.toBeChecked();
            expect(panelDialogTitle).toBeInTheDocument();
            expect(panelDialogDescription).toBeInTheDocument();

            const cancelButton = screen.getByRole('button', { name: /cancel/i });

            await user.click(cancelButton);

            await waitFor(() => expect(panelDialogTitle).not.toBeInTheDocument());
            expect(panelDialogDescription).not.toBeInTheDocument();
            expect(panelSwitch).not.toBeChecked();
        });

        it('on clicking switch shows modal and when clicking confirm closes it and switch changes to enabled', async () => {
            const panelSwitch = screen.getByRole('switch');
            const user = userEvent.setup();

            await user.click(panelSwitch);

            const panelDialogTitle = await screen.findByText(dialogTitle, { exact: false });
            const panelDialogDescription = screen.getByText(/analysis has been added with citrix configuration/i);

            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).not.toBeChecked();
            expect(panelDialogTitle).toBeInTheDocument();
            expect(panelDialogDescription).toBeInTheDocument();

            const confirmButton = screen.getByRole('button', { name: /confirm/i });

            await user.click(confirmButton);

            await waitFor(() => expect(panelDialogTitle).not.toBeInTheDocument());
            expect(panelDialogDescription).not.toBeInTheDocument();
            expect(panelSwitch).toBeChecked();
        });
    });
    describe('Click on switch to disable', () => {
        beforeEach(() => {
            const savedConfigurationValue = true;
            serverState = setInitialServerState(savedConfigurationValue);
            render(<CitrixRDPConfiguration />);
        });
        it('on clicking switch shows modal and when clicking cancel closes it and switch stays enabled', async () => {
            const panelSwitch = screen.getByRole('switch');
            const user = userEvent.setup();

            await user.click(panelSwitch);

            const panelDialogTitle = await screen.findByText(dialogTitle, { exact: false });
            const panelDialogDescription = screen.getByText(/analysis has been removed with citrix configuration/i);

            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).toBeChecked();
            expect(panelDialogTitle).toBeInTheDocument();
            expect(panelDialogDescription).toBeInTheDocument();

            const cancelButton = screen.getByRole('button', { name: /cancel/i });

            await user.click(cancelButton);

            await waitFor(() => expect(panelDialogTitle).not.toBeInTheDocument());
            expect(panelDialogDescription).not.toBeInTheDocument();
            expect(panelSwitch).toBeChecked();
        });
        it('on clicking switch shows modal and when clicking confirm closes it and switch changes to disabled', async () => {
            const panelSwitch = screen.getByRole('switch');
            const user = userEvent.setup();

            await user.click(panelSwitch);

            const panelDialogTitle = await screen.findByText(dialogTitle, { exact: false });
            const panelDialogDescription = screen.getByText(/analysis has been removed with citrix configuration/i);

            expect(panelSwitch).toBeInTheDocument();
            expect(panelSwitch).toBeChecked();
            expect(panelDialogTitle).toBeInTheDocument();
            expect(panelDialogDescription).toBeInTheDocument();

            const confirmButton = screen.getByRole('button', { name: /confirm/i });

            await user.click(confirmButton);

            await waitFor(() => expect(panelDialogTitle).not.toBeInTheDocument());
            expect(panelDialogDescription).not.toBeInTheDocument();
            expect(panelSwitch).not.toBeChecked();
        });
    });
});
