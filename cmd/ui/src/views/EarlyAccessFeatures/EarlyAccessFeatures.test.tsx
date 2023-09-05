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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { Flag } from 'src/hooks/useFeatureFlags';
import { act, render, screen } from 'src/test-utils';
import EarlyAccessFeatures from '.';

const testFeatureFlags: Flag[] = [
    {
        id: 1,
        name: 'feature-flag-1',
        key: 'feature-flag-1',
        description: 'description-1',
        enabled: false,
        user_updatable: true,
    },
    {
        id: 2,
        name: 'feature-flag-2',
        key: 'feature-flag-2',
        description: 'description-2',
        enabled: false,
        user_updatable: true,
    },
    {
        id: 3,
        name: 'feature-flag-3',
        key: 'feature-flag-3',
        description: 'description-3',
        enabled: false,
        user_updatable: true,
    },
];

const server = setupServer(
    rest.get(`/api/v2/features`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: testFeatureFlags,
            })
        );
    })
);

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => ({
    ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
    useNavigate: () => mockNavigate,
}));

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EarlyAccessFeatures', () => {
    it('displays a warning dialog when mounted', async () => {
        await act(async () => {
            render(<EarlyAccessFeatures />);
        });

        expect(screen.getByRole('dialog')).toBeInTheDocument();

        expect(screen.getByRole('dialog')).toBeVisible();

        expect(screen.getByText('Heads up!')).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Take me back' })).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'I understand, show me the new stuff!' })).toBeInTheDocument();
    });

    it('navigates to previous history entry when warning dialog is cancelled', async () => {
        const user = userEvent.setup();

        render(<EarlyAccessFeatures />);

        // Close (cancel) warning dialog
        await user.click(screen.getByRole('button', { name: 'Take me back' }));

        expect(mockNavigate).toHaveBeenCalled();
    });

    it('eventually displays a list of early access features when warning dialog is accepted', async () => {
        render(<EarlyAccessFeatures />);
        const user = userEvent.setup();

        // Close (accept) warning dialog
        await user.click(screen.getByRole('button', { name: 'I understand, show me the new stuff!' }));

        expect(screen.getByRole('dialog')).not.toBeVisible();

        expect(screen.getByText('Early Access Features')).toBeInTheDocument();

        for (const featureFlag of testFeatureFlags) {
            expect(await screen.findByText(featureFlag.name)).toBeInTheDocument();
            expect(await screen.findByText(featureFlag.description)).toBeInTheDocument();
        }
    });

    it('displays a polite message when there are no user updatable features available', async () => {
        server.use(
            rest.get(`/api/v2/features`, (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: testFeatureFlags.map((flag) => ({ ...flag, user_updatable: false })),
                    })
                );
            })
        );

        render(<EarlyAccessFeatures />);
        const user = userEvent.setup();

        // Close (accept) warning dialog
        await user.click(screen.getByRole('button', { name: 'I understand, show me the new stuff!' }));

        expect(screen.getByRole('dialog')).not.toBeVisible();

        expect(screen.getByText('Early Access Features')).toBeInTheDocument();

        expect(await screen.findByText('No Early Access Features Available')).toBeInTheDocument();

        expect(
            await screen.findByText(
                'There are no early access features available at this time. Please check back later.'
            )
        ).toBeInTheDocument();
    });

    it('displays an error when unable to fetch data', async () => {
        console.error = vi.fn();
        server.use(
            rest.get(`/api/v2/features`, (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );
        render(<EarlyAccessFeatures />);
        const user = userEvent.setup();

        // Close (accept) warning dialog
        await user.click(screen.getByRole('button', { name: 'I understand, show me the new stuff!' }));

        expect(screen.getByRole('dialog')).not.toBeVisible();

        expect(screen.getByText('Early Access Features')).toBeInTheDocument();

        expect(await screen.findByText('Could Not Display Early Access Features')).toBeInTheDocument();
    });
});
