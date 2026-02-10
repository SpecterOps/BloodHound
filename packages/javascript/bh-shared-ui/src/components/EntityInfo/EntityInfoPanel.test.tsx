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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../test-utils';

import { AzureNodeKind } from '../../graphSchema';
import { useRoleBasedFiltering } from '../../hooks';
import { allSections } from '../../utils';
import { ObjectInfoPanelContext } from '../../views';
import EntityInfoPanel, { EntityInfoPanelProps } from './EntityInfoPanel';

const objectId = 'fake-object-id';
const azKeyVaultSections: any = allSections[AzureNodeKind.KeyVault]!(objectId);

vi.mock('../../hooks/useRoleBasedFiltering');
const mockUseRoleBasedFiltering = vi.mocked(useRoleBasedFiltering);

const testProps: EntityInfoPanelProps = {
    DataTable: { ...azKeyVaultSections[0] },
};

const setIsObjectInfoPanelOpen = (newValue: boolean) => {
    mockContextValue.isObjectInfoPanelOpen = newValue;
};

const mockContextValue = {
    isObjectInfoPanelOpen: true,
    setIsObjectInfoPanelOpen,
};

const server = setupServer(
    rest.get(`/api/v2/custom-nodes`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoPanel', async () => {
    it('should not display a badge when role based filtering is true and show filter banner is false by default', async () => {
        mockUseRoleBasedFiltering.mockReturnValue(true);

        render(
            <ObjectInfoPanelContext.Provider value={mockContextValue}>
                <EntityInfoPanel {...testProps} />
            </ObjectInfoPanelContext.Provider>
        );

        expect(
            screen.queryByTestId('explore_entity-information-panel-role-based-filtering-badge')
        ).not.toBeInTheDocument();
    });

    it('should display a badge when role based filtering is true and show filter banner is true', async () => {
        mockUseRoleBasedFiltering.mockReturnValue(true);

        render(
            <ObjectInfoPanelContext.Provider value={mockContextValue}>
                <EntityInfoPanel {...testProps} showFilteringBanner={true} />
            </ObjectInfoPanelContext.Provider>
        );

        expect(screen.queryByTestId('explore_entity-information-panel-role-based-filtering-badge')).toBeInTheDocument();
    });

    it('should display a message to select an object ', async () => {
        render(<EntityInfoPanel {...testProps} selectedNode={null} showPlaceholderMessage={true} />);

        const selectObjectMessage = screen.getByText(/Select an object to view the associated information/i);
        expect(selectObjectMessage).toBeInTheDocument();
    });
});
