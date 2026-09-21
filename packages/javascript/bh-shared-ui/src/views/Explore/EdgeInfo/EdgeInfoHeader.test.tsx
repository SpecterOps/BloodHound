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
import { act, render } from '../../../test-utils';
import { ObjectInfoPanelContext } from '../providers';
import EdgeInfoHeader, { HeaderProps } from './EdgeInfoHeader';

const mockClearSelectedItem = vi.fn();

vi.mock('../../../hooks', async () => {
    const actual = await vi.importActual('../../../hooks');

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
};

const setIsObjectInfoPanelOpen = (newValue: boolean) => {
    mockContextValue.isObjectInfoPanelOpen = newValue;
};

const mockContextValue = {
    isObjectInfoPanelOpen: true,
    setIsObjectInfoPanelOpen,
};

const setup = async () => {
    const url = `?expandedPanelSections=['test','test1']`;

    const screen = await act(async () => {
        return render(
            <ObjectInfoPanelContext.Provider value={mockContextValue}>
                <EdgeInfoHeader {...testProps} />
            </ObjectInfoPanelContext.Provider>,
            { route: url }
        );
    });

    const user = userEvent.setup();

    return { screen, user };
};

describe('EdgeInfoHeader', async () => {
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
    it('should display hidden label for hidden edge', async () => {
        const url = `?expandedPanelSections=['test','test1']`;

        const hiddenEdgeTestProps: HeaderProps = {
            name: '** Hidden Edge **',
        };

        const screen = render(
            <ObjectInfoPanelContext.Provider value={mockContextValue}>
                <EdgeInfoHeader {...hiddenEdgeTestProps} />
            </ObjectInfoPanelContext.Provider>,
            { route: url }
        );

        expect(await screen.findByText('** Hidden Edge **')).toBeInTheDocument();
    });
});
