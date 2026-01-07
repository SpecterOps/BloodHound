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
import { Route, Routes, useParams } from 'react-router-dom';
import { zoneHandlers } from '../../../mocks/handlers';
import {
    ROUTE_PRIVILEGE_ZONES,
    ROUTE_PZ_ZONE_OBJECT_DETAILS,
    ROUTE_PZ_ZONE_RULE_OBJECT_DETAILS,
    detailsPath,
    objectsPath,
    privilegeZonesPath,
    rulesPath,
    zonesPath,
} from '../../../routes';
import { render, screen, waitFor, within } from '../../../test-utils';
import Details from './Details';

vi.mock('../../../hooks/useMeasure', () => ({
    useMeasure: vi.fn().mockImplementation(() => [200, 200]),
}));

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useParams: vi.fn(),
        useNavigate: vi.fn,
    };
});

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Details', async () => {
    const user = userEvent.setup();

    it('renders', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });
        render(
            <Routes>
                <Route path={`/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}`} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}` }
        );

        const zonesList = await screen.findByTestId('privilege-zones_details_zones-list');
        expect(within(zonesList).getByText('Zones')).toBeInTheDocument();

        expect(await screen.findByTestId('privilege-zones_details_zones-list')).toBeInTheDocument();
        expect(await screen.findByTestId('privilege-zones_details_rules-list')).toBeInTheDocument();
        expect(await screen.findByTestId('privilege-zones_details_members-list')).toBeInTheDocument();
    });

    it('has Tier Zero zone selected by default and no rules or objects are selected', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });
        render(
            <Routes>
                <Route path={`/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}`} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}` }
        );

        const rules = await screen.findByTestId('privilege-zones_details_rules-list');
        const rulesListItems = await within(rules).findAllByTestId('rule-row');

        const objects = await screen.findByTestId('privilege-zones_details_members-list');
        const objectsListItems = await within(objects).findAllByTestId('member-row');

        await waitFor(() => {
            expect(screen.getByTestId('privilege-zones_details_zones-list_static-order')).toBeInTheDocument();
            // The Tier Zero zone is selected by default
            expect(screen.getByTestId('privilege-zones_details_tags-list_active-tags-item-1')).toBeInTheDocument();
        });

        // No rule is selected to begin with
        rulesListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });

        // No object is selected to begin with
        objectsListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });
    });

    it('handles object selection when a zone is already selected', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined, memberId: '5' });
        render(
            <Routes>
                <Route path={`/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}`} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}` }
        );

        const objects = await screen.findByTestId('privilege-zones_details_members-list');
        const objectsListItems = await within(objects).findAllByTestId('sort-button');
        expect(objectsListItems.length).toBeGreaterThan(0);

        const object5 = await screen.findByText('tag-0-object-5');
        await user.click(object5);

        await waitFor(async () => {
            expect(await screen.findByTestId('privilege-zones_details_members-list_active-members-item-5'));
        });
    });

    it('handles rule selection', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined, ruleId: '7' });
        render(
            <Routes>
                <Route path={ROUTE_PRIVILEGE_ZONES + ROUTE_PZ_ZONE_OBJECT_DETAILS} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${objectsPath}/7/${detailsPath}` }
        );

        const rules = await screen.findByTestId('privilege-zones_details_rules-list');
        await within(rules).findAllByTestId('sort-button');
        const rule7 = await within(rules).findByText('tag-0-rule-7');

        await user.click(rule7);

        // The rule displays as selected
        expect(await screen.findByTestId('privilege-zones_details_rules-list_active-rules-item-7')).toBeInTheDocument();
    });

    it('will deselect both the selected rule and selected object when a different zone is selected', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '3', labelId: undefined });
        render(
            <Routes>
                <Route path={ROUTE_PRIVILEGE_ZONES + ROUTE_PZ_ZONE_RULE_OBJECT_DETAILS} element={<Details />} />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${rulesPath}/7/${objectsPath}/7/${detailsPath}` }
        );

        const rules = await screen.findByTestId('privilege-zones_details_rules-list');
        const rulesListItems = await within(rules).findAllByTestId('sort-button');
        rulesListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });

        const objects = await screen.findByTestId('privilege-zones_details_members-list');
        const objectsListItems = await within(objects).findAllByTestId('sort-button');
        objectsListItems.forEach((li) => {
            expect(li.childNodes).toHaveLength(1);
        });

        const zones = await screen.findByTestId('privilege-zones_details_zones-list');
        await within(zones).findAllByTestId('privilege-zones_details_zones-list_static-order');
        const zone = await within(zones).findByText('tag-2');
        expect(zone).toBeInTheDocument();
        await user.click(zone);

        expect(await screen.findByTestId('privilege-zones_details_tags-list_active-tags-item-3')).toBeInTheDocument();
    });
});
