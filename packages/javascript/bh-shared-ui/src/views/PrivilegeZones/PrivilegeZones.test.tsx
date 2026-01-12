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
import { setupServer } from 'msw/node';
import { Route, Routes } from 'react-router-dom';
import PrivilegeZones from '.';
import zoneHandlers from '../../mocks/handlers/zoneHandlers';
import { detailsPath, labelsPath, privilegeZonesPath, zonesPath } from '../../routes';
import { render, screen, waitFor } from '../../test-utils';

vi.mock('react-router-dom', async () => ({
    ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
    useNavigate: vi.fn,
}));

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Zone Management', async () => {
    const user = userEvent.setup();

    it('allows switching between the Zones and Labels tabs', async () => {
        render(
            <Routes>
                <Route path={`/${privilegeZonesPath}/*`} element={<PrivilegeZones />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}` }
        );

        const labelTab = await screen.findByRole('tab', { name: /Labels/i });
        const zoneTab = await screen.findByRole('tab', { name: /Zones/i });
        const viewTitle = screen.getByText(/Zone Builder/i);
        expect(zoneTab).toHaveAttribute('data-state', 'active');
        expect(labelTab).not.toHaveAttribute('data-state', 'active');
        expect(viewTitle).toBeInTheDocument();

        // Switch to Labels tab
        await user.click(labelTab);

        waitFor(() => {
            expect(zoneTab).not.toHaveAttribute('data-state', 'active');
            expect(labelTab).toHaveAttribute('data-state', 'active');
            expect(window.location.pathname).toBe(`/${privilegeZonesPath}/${labelsPath}/2/${detailsPath}`);
            expect(viewTitle).toBeInTheDocument();
        });

        // Switch back to Zones
        await user.click(zoneTab);

        waitFor(() => {
            expect(zoneTab).toHaveAttribute('data-state', 'active');
            expect(labelTab).not.toHaveAttribute('data-state', 'active');
            expect(window.location.pathname).toBe(`/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}`);
            expect(viewTitle).toBeInTheDocument();
        });
    });
});
