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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, fireEvent, render } from '../../test-utils';
import * as exportUtils from '../../utils/exportGraphData';
import GraphControls from './GraphControls';

const exportToJsonSpy = vi.spyOn(exportUtils, 'exportToJson');

const server = setupServer(
    rest.get('/api/v2/features', (_req, res, ctx) => {
        return res(ctx.json({}));
    }),
    rest.get('/api/v2/customnode', (_, res, ctx) => {
        return res(ctx.json({}));
    }),
    rest.post('/api/v2/graphs/cypher', (_, res, ctx) => {
        return res(ctx.json({ data: { nodes: { test: {} } } }));
    })
);
beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
});
afterAll(() => server.close());

describe('GraphControls', () => {
    const mockJsonData = {};
    const layoutOptions = ['1', '2', '3'] as const;
    const preselectedLayout = layoutOptions[1];
    const currentNodes = {
        '1': {
            label: 'computer_node',
            kind: 'Computer',
            objectId: '001',
            isTierZero: false,
            isOwnedObject: false,
            lastSeen: '',
        },
    };

    const onResetFn = vi.fn();
    const onLayoutChangeFn = vi.fn();
    const onToggleNodeLabelsFn = vi.fn();
    const onToggleEdgeLabelsFn = vi.fn();
    const onSearchedNodeClickFn = vi.fn();

    afterEach(() => {
        onResetFn.mockClear();
        onLayoutChangeFn.mockClear();
        onToggleNodeLabelsFn.mockClear();
        onToggleEdgeLabelsFn.mockClear();
        onSearchedNodeClickFn.mockClear();
    });

    const setup = ({ showNodeLabels = true, showEdgeLabels = true, json = mockJsonData } = {}) => {
        const screen = render(
            <GraphControls
                onReset={onResetFn}
                onLayoutChange={onLayoutChangeFn}
                onToggleNodeLabels={onToggleNodeLabelsFn}
                onToggleEdgeLabels={onToggleEdgeLabelsFn}
                onSearchedNodeClick={onSearchedNodeClickFn}
                showNodeLabels={showNodeLabels}
                showEdgeLabels={showEdgeLabels}
                jsonData={json}
                layoutOptions={layoutOptions}
                selectedLayout={preselectedLayout}
                currentNodes={currentNodes}
            />
        );

        const user = userEvent.setup();

        return { screen, user };
    };

    describe('Resetting graph', () => {
        it('calls the onReset prop when the crop button is clicked', async () => {
            const { screen, user } = setup();

            const crop = screen.getByText('crop-simple');
            await user.click(crop);

            expect(onResetFn).toBeCalled();
        });
    });

    describe('Toggling labels', () => {
        it('calls onToggleNodeLabels when click show all node labels', async () => {
            const { screen, user } = setup();
            const labelMenu = screen.getByText('Hide Labels');
            await user.click(labelMenu);

            const hideNodeLabels = await screen.findByText('Hide Node Labels');
            await user.click(hideNodeLabels);

            expect(onToggleNodeLabelsFn).toBeCalled();
        });
        it('calls onToggleEdgeLabels when click show all edge labels', async () => {
            const { screen, user } = setup();
            const labelMenu = screen.getByText('Hide Labels');
            await user.click(labelMenu);

            const hideEdgeLabels = await screen.findByText('Hide Edge Labels');
            await user.click(hideEdgeLabels);

            expect(onToggleEdgeLabelsFn).toBeCalled();
        });
        it.each([
            { showNodeLabels: true, showEdgeLabels: true },
            { showNodeLabels: false, showEdgeLabels: false },
            { showNodeLabels: false, showEdgeLabels: true },
            { showNodeLabels: true, showEdgeLabels: false },
        ])(
            'Toggles node and edge labels on/off depending on their existing state',
            async ({ showEdgeLabels, showNodeLabels }) => {
                const { screen, user } = setup({ showEdgeLabels, showNodeLabels });
                const labelMenu = screen.getByText('Hide Labels');
                await user.click(labelMenu);

                const allLabelsController = await screen.findByRole('menuitem', { name: /All Labels/i });
                await user.click(allLabelsController);

                if (!showEdgeLabels) expect(onToggleEdgeLabelsFn).toBeCalled();
                if (!showNodeLabels) expect(onToggleNodeLabelsFn).toBeCalled();

                if (showEdgeLabels && showNodeLabels) {
                    expect(onToggleEdgeLabelsFn).toBeCalled();
                    expect(onToggleNodeLabelsFn).toBeCalled();
                }
            }
        );
    });

    describe('Selecting a layout', () => {
        it('calls onLayoutChange with the selected layout', async () => {
            const { screen, user } = setup();

            const layoutMenu = screen.getByText('Layout');
            await user.click(layoutMenu);

            const selectedLayout = layoutOptions[0];
            const firstLayout = await screen.findByText(layoutOptions[0]);
            await user.click(firstLayout);

            expect(onLayoutChangeFn).toBeCalledWith(selectedLayout);
        });
        it('displays active styles for the selected layout when explore table is enabled', async () => {
            server.use(
                rest.get('/api/v2/features', async (_req, res, ctx) => {
                    return res(ctx.json({ data: [{ key: 'explore_table_view', enabled: true }] }));
                })
            );

            const { screen, user } = setup();

            const layoutMenu = screen.getByText('Layout');
            await user.click(layoutMenu);

            const selectedLayout = await screen.findByText(preselectedLayout);
            expect(selectedLayout).toHaveClass('Mui-selected');
        });
    });
    describe('Exporting json', () => {
        it('disables the JSON button if the JSON is empty', async () => {
            const { screen, user } = setup();

            const exportMenu = screen.getByText('Export');
            await user.click(exportMenu);

            const jsonButton = await screen.findByText('JSON');

            expect(jsonButton).toHaveClass('Mui-disabled');
        });

        it('calls exportToJson util when valid non a empty object is passed as the jsonData prop', async () => {
            exportToJsonSpy.mockImplementationOnce(() => undefined);

            const json = { test: 'data' };
            const { screen, user } = setup({ json });

            const exportMenu = screen.getByText('Export');
            await user.click(exportMenu);

            const jsonButton = await screen.findByText('JSON');
            await user.click(jsonButton);

            expect(exportToJsonSpy).toBeCalledWith(json);
        });
    });
    describe('Searching current results', () => {
        it('renders GraphButton with correct text', () => {
            const { screen } = setup();

            const searchResultsMenu = screen.getByText('Search Current Results');

            expect(searchResultsMenu).toBeInTheDocument();
        });

        it('disables GraphButton when isCurrentSearchOpen is true', async () => {
            const { screen, user } = setup();

            const searchResultsMenu = screen.getByText('Search Current Results');
            await user.click(searchResultsMenu);

            expect(searchResultsMenu).toBeDisabled();
        });

        it('shows Popper when isCurrentSearchOpen is true', async () => {
            const { screen, user } = setup();

            expect(screen.queryByTestId('explore_graph-controls')).not.toBeInTheDocument();

            const searchResultsMenu = screen.getByText('Search Current Results');
            await user.click(searchResultsMenu);

            expect(screen.getByTestId('explore_graph-controls')).toBeInTheDocument();
        });

        it('sets the selectedItem param and closes popper when a node is selected', async () => {
            const { screen, user } = setup();

            const searchResultsMenu = screen.getByText('Search Current Results');
            await user.click(searchResultsMenu);

            const searchInput = await screen.findByPlaceholderText('Search Current Results');
            await user.type(searchInput, currentNodes[1].label);

            const searchedNode = await screen.findByTestId('explore_search_result-list-item');
            // Something about fireEvent.click triggers the correct event within SearchCurrentNodes.
            // This is a known issue in downshift and seems to be resolved in newer versions
            act(() => fireEvent.click(searchedNode));

            expect(onSearchedNodeClickFn).toBeCalled();
            expect(screen.queryByTestId('explore_graph-controls')).not.toBeInTheDocument();
        });
    });
});
