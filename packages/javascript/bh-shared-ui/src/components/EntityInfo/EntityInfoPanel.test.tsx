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
import { act, render } from '../../test-utils';

import userEvent from '@testing-library/user-event';
import { AzureNodeKind } from '../../graphSchema';
import { allSections } from '../../utils';
import { ObjectInfoPanelContext } from '../../views';
//import EntityInfoHeader, { HeaderProps } from './EntityInfoHeader';
import EntityInfoPanel, { EntityInfoPanelProps } from './EntityInfoPanel';

const mockClearSelectedItem = vi.fn();

const objectId = 'fake-object-id';
const azKeyVaultSections: any = allSections[AzureNodeKind.KeyVault]!(objectId);

/*
    id: string;
    label: string;
    countLabel?: string;
    sections?: EntityInfoDataTableProps[];
    parentLabels?: string[];
    queryType?: EntityRelationshipQueryTypes;
*/

vi.mock('../../hooks', async () => {
    const actual = await vi.importActual('../../hooks');

    return {
        ...actual,
        useExploreSelectedItem: () => ({
            clearSelectedItem: mockClearSelectedItem,
            selectedItem: '123',
        }),
    };
});

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

const setup = async () => {
    const url = `?expandedPanelSections=['test','test1']`;

    const screen = await act(async () => {
        return render(
            <ObjectInfoPanelContext.Provider value={mockContextValue}>
                <EntityInfoPanel {...testProps} />
            </ObjectInfoPanelContext.Provider>,
            { route: url }
        );
    });

    const user = userEvent.setup();

    return { screen, user };
};

describe('EntityInfoPanel', async () => {
    it('should render', async () => {
        const { screen, user } = await setup();

        screen.debug(undefined, Infinity);

        const clearItemButton = screen.getByRole('button', { name: 'Clear selected item' });
        await user.hover(clearItemButton);
        expect(await screen.findByRole('tooltip', { name: /Clear selected item/ })).toBeInTheDocument();

        const collapseAllButton = screen.getByRole('button', { name: 'Collapse All' });
        await user.hover(collapseAllButton);
        expect(await screen.findByRole('tooltip', { name: /collapse all/i })).toBeInTheDocument();

        const edgeTitle = screen.getByRole('heading');
        expect(edgeTitle).toBeInTheDocument();
        expect(edgeTitle).toHaveTextContent(testProps.name);
    });
});
