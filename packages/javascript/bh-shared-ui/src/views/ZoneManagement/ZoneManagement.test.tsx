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
import { Route, Routes } from 'react-router-dom';
import ZoneManagement from '.';
import { render, screen, waitFor } from '../../test-utils';

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

vi.mock('react-router-dom', async () => ({
    ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
    useNavigate: vi.fn,
}));

const server = setupServer(
    rest.get(`/api/v2/asset-group-tags`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    tags: [
                        {
                            id: 2,
                            type: 3,
                            kind_id: 179,
                            name: 'Owned',
                            description: 'Owned',
                            position: null,
                            require_certify: null,
                            analysis_enabled: null,
                        },
                        {
                            id: 1,
                            type: 1,
                            kind_id: 173,
                            name: 'Tier Zero',
                            description: 'Tier Zero description',
                            position: 1,
                            require_certify: false,
                            analysis_enabled: true,
                        },
                    ],
                },
            })
        );
    }),
    rest.get('/api/v2/features', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'tier_management_engine',
                        enabled: true,
                    },
                ],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Zone Management', async () => {
    const user = userEvent.setup();

    it('allows switching between the Tiers and Labels tabs', async () => {
        render(
            <Routes>
                <Route path='/zone-management/details/tier/:tierId/*' element={<ZoneManagement />} />
                <Route path='/' element={<ZoneManagement />} />
            </Routes>,
            { route: '/zone-management/details/tier/1' }
        );

        const labelTab = await screen.findByRole('tab', { name: /Labels/i });
        const tierTab = await screen.findByRole('tab', { name: /Tiers/i });

        expect(tierTab).toHaveAttribute('data-state', 'active');
        expect(labelTab).not.toHaveAttribute('data-state', 'active');

        // Switch to Labels tab
        await user.click(labelTab);

        waitFor(() => {
            expect(tierTab).not.toHaveAttribute('data-state', 'active');
            expect(labelTab).toHaveAttribute('data-state', 'active');
            expect(window.location.pathname).toBe('/zone-management/details/label/2');
        });

        // Switch back to Tiers
        await user.click(tierTab);

        waitFor(() => {
            expect(tierTab).toHaveAttribute('data-state', 'active');
            expect(labelTab).not.toHaveAttribute('data-state', 'active');
            expect(window.location.pathname).toBe('/zone-management/details/tier/1');
        });
    });
});
