// Copyright 2026 Specter Ops, Inc.
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
import { CustomRulesKey, DefaultRulesKey, DisabledRulesKey, RulesKey } from 'js-client-library';
import { setupServer } from 'msw/node';
import { PrivilegeZonesContext, defaultPrivilegeZoneCtxValue } from '..';
import * as usePZParams from '../../../hooks/usePZParams/usePZPathParams';
import { mockPZPathParams } from '../../../mocks/factories/privilegeZones';
import { zoneHandlers } from '../../../mocks/handlers';
import { render, screen } from '../../../test-utils';
import PZDetailsTabProvider from './SelectedDetailsTabs/SelectedDetailsTabsProvider';
import TagSelector from './TagSelector';

vi.mock('../../../hooks/useSelectedTag', () => ({
    useSelectedTagPathParams: () => ({
        counts: {
            [RulesKey]: 6,
            [CustomRulesKey]: 2,
            [DefaultRulesKey]: 2,
            [DisabledRulesKey]: 2,
        },
    }),
}));

const usePZPathParamsSpy = vi.spyOn(usePZParams, 'usePZPathParams');

const server = setupServer(...zoneHandlers);

const PrivilegeZonesProviderEmptySelectors: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    return (
        <PrivilegeZonesContext.Provider
            value={{ ...defaultPrivilegeZoneCtxValue, EnvironmentSelector: () => <></>, InfoHeader: () => <></> }}>
            {children}
        </PrivilegeZonesContext.Provider>
    );
};

const PrivilegeZonesProviderWithSelectors: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    return (
        <PrivilegeZonesContext.Provider
            value={{
                ...defaultPrivilegeZoneCtxValue,
                EnvironmentSelector: () => <></>,
                InfoHeader: () => <></>,
                ZoneSelector: () => <span>ZoneSelector</span>,
                LabelSelector: () => <span>LabelSelector</span>,
            }}>
            {children}
        </PrivilegeZonesContext.Provider>
    );
};

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('TagSelector', () => {
    it('renders Tier Zero text by default when no Zone Selector is available and on a Zone page', async () => {
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams });
        render(
            <PrivilegeZonesProviderEmptySelectors>
                <PZDetailsTabProvider>
                    <TagSelector />
                </PZDetailsTabProvider>
            </PrivilegeZonesProviderEmptySelectors>
        );

        expect(await screen.findByText('Tier Zero')).toBeInTheDocument();
    });

    it('renders Owned text by default when no Label Selector is available and on a Label page', async () => {
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams, isLabelPage: true, isZonePage: false });
        render(
            <PrivilegeZonesProviderEmptySelectors>
                <PZDetailsTabProvider>
                    <TagSelector />
                </PZDetailsTabProvider>
            </PrivilegeZonesProviderEmptySelectors>
        );

        expect(await screen.findByText('Owned')).toBeInTheDocument();
    });

    it('renders the ZoneSelector when available and on a Zone page', async () => {
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams });
        render(
            <PrivilegeZonesProviderWithSelectors>
                <PZDetailsTabProvider>
                    <TagSelector />
                </PZDetailsTabProvider>
            </PrivilegeZonesProviderWithSelectors>
        );

        expect(await screen.findByText('ZoneSelector')).toBeInTheDocument();
    });

    it('renders the LabelSelector when available and on a Label page', async () => {
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams, isLabelPage: true, isZonePage: false });
        render(
            <PrivilegeZonesProviderWithSelectors>
                <PZDetailsTabProvider>
                    <TagSelector />
                </PZDetailsTabProvider>
            </PrivilegeZonesProviderWithSelectors>
        );

        expect(await screen.findByText('LabelSelector')).toBeInTheDocument();
    });
});
