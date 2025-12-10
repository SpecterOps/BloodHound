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
import { ObjectInfoPanelContext } from '../../views';
import EntityInfoHeader, { HeaderProps } from './EntityInfoHeader';

const mockClearSelectedItem = vi.fn();

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

const testProps: HeaderProps = {
    name: 'testName',
    nodeType: AzureNodeKind.Group,
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
                <EntityInfoHeader {...testProps} />
            </ObjectInfoPanelContext.Provider>,
            { route: url }
        );
    });

    const user = userEvent.setup();

    return { screen, user };
};

describe('EntityInfoHeader', async () => {
    it('should render', async () => {
        const { screen, user } = await setup();

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
    it('should on clicking collapse all remove expandedPanelSections param from url and set isObjectInfoPanelOpen in context to false', async () => {
        const { screen, user } = await setup();
        const collapseAllButton = screen.getByRole('button', { name: 'Collapse All' });

        await user.click(collapseAllButton);

        expect(window.location.search).not.toContain('expandedPanelSections');
        expect(mockContextValue.isObjectInfoPanelOpen).toBe(false);
    });
    it('should on clicking remove call clearSelectedItem', async () => {
        const { screen, user } = await setup();
        const clearItemButton = screen.getByRole('button', { name: 'Clear selected item' });

        await user.click(clearItemButton);

        expect(mockClearSelectedItem).toBeCalled();
    });
});
