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
                <Route path='/zone-management/details/tier/:tierId' element={<Details />} />
            </Routes>,
            { route: '/zone-management/details/tier/1' }
        );

        expect(await screen.findByText(/Tiers/)).toBeInTheDocument();

        expect(await screen.findByTestId('zone-management_details_tiers-list')).toBeInTheDocument();
        expect(await screen.findByTestId('zone-management_details_selectors-list')).toBeInTheDocument();
        expect(await screen.findByTestId('zone-management_details_members-list')).toBeInTheDocument();
    });

    it('has Tier Zero tier selected by default and no selectors or objects are selected', async () => {
        render(
            <Routes>
                <Route path='/zone-management/details/tier/:tierId' element={<Details />} />
            </Routes>,
            { route: '/zone-management/details/tier/1' }
        );

        const selectors = await screen.findByTestId('zone-management_details_selectors-list');
        const selectorsListItems = await within(selectors).findAllByRole('listitem');

        const objects = await screen.findByTestId('zone-management_details_members-list');
        const objectsListItems = await within(objects).findAllByRole('listitem');

        longWait(() => {
            expect(screen.getByTestId('zone-management_details_tiers-list')).toBeInTheDocument();
            // The Tier Zero tier is selected by default
            expect(screen.getByTestId('zone-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();
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

    it('handles object selection when a tier is already selected', async () => {
        render(
            <Routes>
                <Route path='/zone-management/details/tier/:tierId' element={<Details />} />
            </Routes>,
            { route: '/zone-management/details/tier/1' }
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
            expect(screen.getByTestId('zone-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();
        });

        // No selector is selected
        selectorsListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });

        longWait(async () => {
            expect(await screen.findByTestId('zone-management_details_members-list_active-members-item-5'));
        });
    });

    it('handles selector selection when a tier and object are already selected', async () => {
        render(
            <Routes>
                <Route path='/zone-management/details/tier/:tierId/member/:memberId' element={<Details />} />
            </Routes>,
            { route: '/zone-management/details/tier/1/member/7' }
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

            // The Tier Zero tier is still selected when selecting a selector that is within it
            expect(screen.getByTestId('zone-management_details_tiers-list_active-tiers-item-1')).toBeInTheDocument();

            expect(await screen.findByRole('button', { name: /Edit/ })).toBeInTheDocument();
        });
    });

    it('will deselect both the selected selector and selected object when a different tier is selected', async () => {
        render(
            <Routes>
                <Route
                    path='/zone-management/details/tier/:tierId/selector/:selectorId/member/:memberId'
                    element={<Details />}
                />
            </Routes>,
            { route: '/zone-management/details/tier/1/selector/7/member/7' }
        );

        const selectors = await screen.findByTestId('zone-management_details_selectors-list');
        let selectorsListItems = await within(selectors).findAllByRole('listitem');

        const objects = await screen.findByTestId('zone-management_details_members-list');
        let objectsListItems = await within(objects).findAllByRole('listitem');

        longWait(async () => {
            expect(screen.getByText('Tier-2')).toBeInTheDocument();
            await user.click(screen.getByText('Tier-2'));

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
                await screen.findByTestId('zone-management_details_tiers-list_active-tiers-item-3')
            ).toBeInTheDocument();
        });
    });
});
