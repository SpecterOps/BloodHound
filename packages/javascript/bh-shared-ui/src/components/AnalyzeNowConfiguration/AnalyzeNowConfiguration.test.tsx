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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../test-utils';
import AnalyzeNowConfiguration from './AnalyzeNowConfiguration';

const addNotificationMock = vi.fn();

vi.mock('../../providers', async () => {
    const actual = await vi.importActual('../../providers');
    return {
        ...actual,
        useNotifications: () => {
            return { addNotification: addNotificationMock };
        },
    };
});

describe('AnalyzeNowConfiguration', () => {
    const server = setupServer(
        rest.get(`/api/v2/datapipe/status`, async (_req, res, ctx) => {
            return res(
                ctx.json({
                    data: {
                        status: 'idle',
                    },
                })
            );
        })
    );

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('renders', () => {
        render(<AnalyzeNowConfiguration />);

        const title = screen.getByText(/Run Analysis Now/i);
        const button = screen.getByRole('button', { name: /Analyze Now/i });

        expect(title).toBeInTheDocument();
        expect(button).toBeInTheDocument();
    });

    it('Disable Analyze Now if datapipe status is different from "idle"', () => {
        server.use(
            rest.get(`/api/v2/datapipe/status`, async (_req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            status: 'ingesting',
                        },
                    })
                );
            })
        );
        render(<AnalyzeNowConfiguration />);

        const button = screen.getByRole('button', { name: /Analyze Now/i });

        expect(button).toBeInTheDocument();
        expect(button).toBeDisabled();
    });

    it('Displays a modal after clicking Analyze Now button and run analysis after clicking confirm', async () => {
        let analysisRequested = false;
        server.use(
            rest.put(`/api/v2/analysis`, (req, res, ctx) => {
                analysisRequested = true;
                return res(ctx.status(202));
            })
        );
        render(<AnalyzeNowConfiguration />);

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: /Analyze Now/i });

        await user.click(button);

        const modalDialog = await screen.getByText(
            /Analysis may take some time, during which your data will be in flux. Proceed with analysis?/i
        );
        const confirmButton = screen.getByRole('button', { name: /Confirm/i });
        const cancelButton = screen.getByRole('button', { name: /Cancel/i });

        expect(modalDialog).toBeInTheDocument();
        expect(confirmButton).toBeInTheDocument();
        expect(cancelButton).toBeInTheDocument();

        await user.click(confirmButton);
        expect(analysisRequested).toBeTruthy();
    });

    it('Displays a notification when an error happened when requesting analysis', async () => {
        server.use(
            rest.put(`/api/v2/analysis`, (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );
        console.error = vi.fn();
        render(<AnalyzeNowConfiguration />);

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: /Analyze Now/i });

        await user.click(button);

        const modalDialog = await screen.getByText(
            /Analysis may take some time, during which your data will be in flux. Proceed with analysis?/i
        );
        const confirmButton = screen.getByRole('button', { name: /Confirm/i });
        const cancelButton = screen.getByRole('button', { name: /Cancel/i });

        expect(modalDialog).toBeInTheDocument();
        expect(confirmButton).toBeInTheDocument();
        expect(cancelButton).toBeInTheDocument();

        await user.click(confirmButton);

        expect(addNotificationMock).toBeCalledWith('There was an error requesting analysis.');
    });

    it('Displays a notification when analysis has been requested successfully', async () => {
        server.use(
            rest.put(`/api/v2/analysis`, (req, res, ctx) => {
                return res(ctx.status(202));
            })
        );
        console.error = vi.fn();
        render(<AnalyzeNowConfiguration />);

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: /Analyze Now/i });

        await user.click(button);

        const modalDialog = await screen.getByText(
            /Analysis may take some time, during which your data will be in flux. Proceed with analysis?/i
        );
        const confirmButton = screen.getByRole('button', { name: /Confirm/i });
        const cancelButton = screen.getByRole('button', { name: /Cancel/i });

        expect(modalDialog).toBeInTheDocument();
        expect(confirmButton).toBeInTheDocument();
        expect(cancelButton).toBeInTheDocument();

        await user.click(confirmButton);

        expect(addNotificationMock).toBeCalledWith('Analysis requested successfully.');
    });
});
