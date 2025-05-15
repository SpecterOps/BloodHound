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
import { createMemoryHistory } from 'history';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render } from '../../../test-utils';
import { ObjectInfoPanelContext } from '../providers';
import EdgeInfoHeader, { HeaderProps } from './EdgeInfoHeader';

const testProps: HeaderProps = {
    expanded: true,
    name: 'testName',
    onToggleExpanded: vi.fn(),
};

const backButtonSupportFF = {
    key: 'back_button_support',
    enabled: true,
};

const setIsObjectInfoPanelOpen = (newValue: boolean) => {
    mockContextValue.isObjectInfoPanelOpen = newValue;
};

const mockContextValue = {
    isObjectInfoPanelOpen: true,
    setIsObjectInfoPanelOpen,
};

const server = setupServer(
    rest.get('/api/v2/features', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [backButtonSupportFF],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const setup = async () => {
    const url = `?expandedPanelSections=['test','test1']`;

    const history = createMemoryHistory({ initialEntries: [url] });

    const screen = await act(async () => {
        return render(
            <ObjectInfoPanelContext.Provider value={mockContextValue}>
                <EdgeInfoHeader {...testProps} />
            </ObjectInfoPanelContext.Provider>,
            { history }
        );
    });

    const user = userEvent.setup();

    return { screen, user, history };
};

describe('EdgeInfoHeader', async () => {
    it('should render', async () => {
        const { screen } = await setup();

        const collapsePanelButton = screen.getByRole('button', { name: /minus/i });
        const edgeTitle = screen.getByRole('heading');
        const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });

        expect(collapsePanelButton).toBeInTheDocument();
        expect(edgeTitle).toBeInTheDocument();
        expect(edgeTitle).toHaveTextContent(testProps.name);
        expect(collapseAllButton).toBeInTheDocument();
    });
    it('should on clicking collapse all remove expandedPanelSections param from url and set isObjectInfoPanelOpen in context to false', async () => {
        const { screen, history, user } = await setup();
        const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });

        await user.click(collapseAllButton);

        expect(history.location.search).not.toContain('expandedPanelSections');
        expect(mockContextValue.isObjectInfoPanelOpen).toBe(false);
    });
});
