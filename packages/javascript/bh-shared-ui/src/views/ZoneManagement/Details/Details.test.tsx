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
import { zoneHandlers } from '../../../mocks/handlers';
import {
    ROUTE_PRIVILEGE_ZONES_ROOT,
    zonePath,
    ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS,
    ROUTE_PRIVILEGE_ZONES_ZONE_OBJECT_DETAILS,
    ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_OBJECT_DETAILS,
    detailsPath,
    privilegeZonesPath,
    selectorPath,
    memberPath,
} from '../../../routes';
import { longWait, render, screen, within } from '../../../test-utils';
import Details from './Details';

vi.mock('../../../hooks/useMeasure', () => ({
    useMeasure: vi.fn().mockImplementation(() => [200, 200]),
}));

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

vi.mock('react-router-dom', async () => ({
    ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
    useNavigate: vi.fn,
}));

describe('Details', async () => {
    const user = userEvent.setup();

    it('renders', async () => {
        render(
            <Routes>
                <Route path={ROUTE_PRIVILEGE_ZONES_ROOT + ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonePath}/1/${detailsPath}` }
        );

        expect(await screen.findByText(/Zones/)).toBeInTheDocument();

        expect(await screen.findByTestId('zone-management_details_zones-list')).toBeInTheDocument();
        expect(await screen.findByTestId('zone-management_details_selectors-list')).toBeInTheDocument();
        expect(await screen.findByTestId('zone-management_details_members-list')).toBeInTheDocument();
    });

    it('has Tier Zero zone selected by default and no selectors or objects are selected', async () => {
        render(
            <Routes>
                <Route path={ROUTE_PRIVILEGE_ZONES_ROOT + ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonePath}/1/${detailsPath}` }
        );

        const selectors = await screen.findByTestId('zone-management_details_selectors-list');
        const selectorsListItems = await within(selectors).findAllByRole('listitem');

        const objects = await screen.findByTestId('zone-management_details_members-list');
        const objectsListItems = await within(objects).findAllByRole('listitem');

        longWait(() => {
            expect(screen.getByTestId('zone-management_details_zones-list')).toBeInTheDocument();
            // The Tier Zero zone is selected by default
            expect(screen.getByTestId('zone-management_details_zones-list_active-zones-item-1')).toBeInTheDocument();
        });

        // No selector is selected to begin with
        selectorsListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });

        // No object is selected to begin with
        objectsListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });
    });

    it('handles object selection when a zone is already selected', async () => {
        render(
            <Routes>
                <Route path={ROUTE_PRIVILEGE_ZONES_ROOT + ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonePath}/1/${detailsPath}` }
        );

        longWait(async () => {
            const object5 = await screen.findByText('tier-0-object-5');
            await user.click(object5);
        });

        const selectors = await screen.findByTestId('zone-management_details_selectors-list');
        const selectorsListItems = await within(selectors).findAllByRole('listitem');

        // After selecting an object, the edit action is not viable and thus the button is not rendered
        expect(screen.queryByRole('button', { name: /Edit/ })).toBeNull();

        longWait(() => {
            // The Tier Zero tier is still selected when selecting an object that is within it
            expect(screen.getByTestId('zone-management_details_zones-list_active-zones-item-1')).toBeInTheDocument();
        });

        // No selector is selected
        selectorsListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });

        longWait(async () => {
            expect(await screen.findByTestId('zone-management_details_members-list_active-members-item-5'));
        });
    });

    it('handles selector selection when a zone and object are already selected', async () => {
        render(
            <Routes>
                <Route
                    path={ROUTE_PRIVILEGE_ZONES_ROOT + ROUTE_PRIVILEGE_ZONES_ZONE_OBJECT_DETAILS}
                    element={<Details />}
                />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonePath}/1/${memberPath}/7/${detailsPath}` }
        );

        const selector7 = await screen.findByText('tier-0-selector-7');

        const objects = await screen.findByTestId('zone-management_details_members-list');
        const objectsListItems = within(objects).getAllByRole('listitem');
        // Selecting a selector will enable the Edit button from a disabled state

        await user.click(selector7);

        longWait(async () => {
            // The selector now displays as selected
            expect(
                await screen.findByTestId('zone-management_details_selectors-list_active-selectors-item-7')
            ).toBeInTheDocument();

            // Selecting a selector after having an Object selected will deselect that Object
            objectsListItems.forEach((li) => {
                expect(li.childNodes).toHaveLength(1);
            });

            // The Tier Zero zone is still selected when selecting a selector that is within it
            expect(screen.getByTestId('zone-management_details_zones-list_active-zones-item-1')).toBeInTheDocument();

            expect(await screen.findByRole('button', { name: /Edit/ })).toBeInTheDocument();
        });
    });

    it('will deselect both the selected selector and selected object when a different zone is selected', async () => {
        render(
            <Routes>
                <Route
                    path={ROUTE_PRIVILEGE_ZONES_ROOT + ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_OBJECT_DETAILS}
                    element={<Details />}
                />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonePath}/1/${selectorPath}/7/${memberPath}/7/${detailsPath}` }
        );

        const selectors = await screen.findByTestId('zone-management_details_selectors-list');
        let selectorsListItems = await within(selectors).findAllByRole('listitem');

        const objects = await screen.findByTestId('zone-management_details_members-list');
        let objectsListItems = await within(objects).findAllByRole('listitem');

        longWait(async () => {
            expect(screen.getByText('Zone-2')).toBeInTheDocument();
            await user.click(screen.getByText('Zone-2'));

            // This list rerenders with different list items so we have to grab those again
            selectorsListItems = await within(selectors).findAllByRole('listitem');
            selectorsListItems.forEach((li) => {
                expect(li.childNodes).toHaveLength(1);
            });

            // This list rerenders with different list items so we have to grab those again
            objectsListItems = await within(objects).findAllByRole('listitem');
            objectsListItems.forEach((li) => {
                expect(li.childNodes).toHaveLength(1);
            });

            expect(
                await screen.findByTestId('zone-management_details_zones-list_active-zones-item-3')
            ).toBeInTheDocument();
        });
    });
});
