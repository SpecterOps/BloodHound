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
import { mockSourceKindsHandler } from '../../mocks';
import { render, screen } from '../../test-utils';

import { AzureNodeKind } from '../../graphSchema';
import { useRoleBasedFiltering } from '../../hooks';
import { allSections, NoEntitySelectedHeader } from '../../utils';
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
    }),
    mockSourceKindsHandler()
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoPanel', async () => {
    it('should not display a badge when role based filtering is false', async () => {
        mockUseRoleBasedFiltering.mockReturnValue(false);

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
                <EntityInfoPanel {...testProps} />
            </ObjectInfoPanelContext.Provider>
        );

        expect(screen.queryByTestId('explore_entity-information-panel-role-based-filtering-badge')).toBeInTheDocument();
    });

    it('should display a none selected header and a message to select an object when there is no selected node', async () => {
        render(<EntityInfoPanel {...testProps} showPlaceholderMessage={true} />);

        const entityHeaderTitle = screen.getByText(NoEntitySelectedHeader);
        const selectObjectMessage = screen.getByText(/Select an object to view the associated information/i);
        expect(entityHeaderTitle).toBeInTheDocument();
        expect(selectObjectMessage).toBeInTheDocument();
    });
});
