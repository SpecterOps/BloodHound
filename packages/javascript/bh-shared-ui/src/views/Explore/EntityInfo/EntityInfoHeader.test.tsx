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
import { act, render } from '../../../test-utils';

import userEvent from '@testing-library/user-event';
import { AzureNodeKind } from '../../../graphSchema';
import { ObjectInfoPanelContext } from '../providers/ObjectInfoPanelProvider';
import EntityInfoHeader, { HeaderProps } from './EntityInfoHeader';

const testProps: HeaderProps = {
    expanded: true,
    name: 'testName',
    onToggleExpanded: vi.fn(),
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
    rest.get(`/api/v2/custom-node`, async (req, res, ctx) => {
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
        const { screen } = await setup();

        const collapsePanelButton = screen.getByRole('button', { name: /minus/i });
        const nodeIcon = screen.getByTitle(testProps.nodeType!);
        const nodeTitle = screen.getByRole('heading');
        const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });

        expect(collapsePanelButton).toBeInTheDocument();
        expect(nodeIcon).toBeInTheDocument();
        expect(nodeTitle).toBeInTheDocument();
        expect(nodeTitle).toHaveTextContent(testProps.name);
        expect(collapseAllButton).toBeInTheDocument();
    });
    it('should on clicking collapse all remove expandedPanelSections param from url and set isObjectInfoPanelOpen in context to false', async () => {
        const { screen, user } = await setup();
        const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });

        await user.click(collapseAllButton);

        expect(window.location.search).not.toContain('expandedPanelSections');
        expect(mockContextValue.isObjectInfoPanelOpen).toBe(false);
    });
});
