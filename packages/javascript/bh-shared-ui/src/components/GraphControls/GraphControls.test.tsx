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
import { act, render, screen, waitFor } from '../../test-utils';
import * as exportUtils from '../../utils/exportGraphData';
import GraphControls, { type GraphExportAction } from './GraphControls';

const exportToJsonSpy = vi.spyOn(exportUtils, 'exportToJson');

const server = setupServer(
    rest.get('/api/v2/features', (_req, res, ctx) => {
        return res(ctx.json({}));
    }),
    rest.get('/api/v2/custom-nodes', (_, res, ctx) => {
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
            kinds: [],
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
    const onAdditionalExportActionOne = vi.fn();
    const onAdditionalExportActionTwo = vi.fn();
    const additionalExportActions: readonly GraphExportAction[] = [
        { id: 'additional-one', label: 'Additional action one', onSelect: onAdditionalExportActionOne },
        { id: 'additional-two', label: 'Additional action two', onSelect: onAdditionalExportActionTwo },
    ];

    afterEach(() => {
        onResetFn.mockClear();
        onLayoutChangeFn.mockClear();
        onToggleNodeLabelsFn.mockClear();
        onToggleEdgeLabelsFn.mockClear();
        onSearchedNodeClickFn.mockClear();
        onAdditionalExportActionOne.mockClear();
        onAdditionalExportActionTwo.mockClear();
    });

    type SetupOptions = {
        showNodeLabels?: boolean;
        showEdgeLabels?: boolean;
        json?: Record<string, any>;
        selectedLayout?: string;
        layoutOptionsOverride?: readonly string[];
        isExploreLayoutSelected?: boolean;
        isExploreTableSelected?: boolean;
        additionalExportActions?: readonly GraphExportAction[];
        route?: string;
    };

    const setup = ({
        showNodeLabels = true,
        showEdgeLabels = true,
        json = mockJsonData,
        selectedLayout = preselectedLayout,
        layoutOptionsOverride,
        isExploreLayoutSelected,
        isExploreTableSelected,
        additionalExportActions,
        route = '/',
    }: SetupOptions = {}) => {
        const options = layoutOptionsOverride ?? layoutOptions;
        render(
            <GraphControls
                onReset={onResetFn}
                onLayoutChange={onLayoutChangeFn}
                onToggleNodeLabels={onToggleNodeLabelsFn}
                onToggleEdgeLabels={onToggleEdgeLabelsFn}
                onSearchedNodeClick={onSearchedNodeClickFn}
                showNodeLabels={showNodeLabels}
                showEdgeLabels={showEdgeLabels}
                jsonData={json}
                layoutOptions={options}
                selectedLayout={selectedLayout}
                isExploreLayoutSelected={isExploreLayoutSelected}
                isExploreTableSelected={isExploreTableSelected}
                additionalExportActions={additionalExportActions}
                currentNodes={currentNodes}
            />,
            { route }
        );

        const user = userEvent.setup();

        return { user };
    };

    describe('Accessible icon controls', () => {
        it('provides accessible names without visible label text', () => {
            setup();

            for (const name of ['Reset Graph', 'Hide Labels', 'Layout', 'Export', 'Search']) {
                const control = screen.getByRole('button', { name });

                expect(control).toBeInTheDocument();
                expect(control).not.toHaveTextContent(name);
            }
        });

        it('shows matching tooltips on hover and keyboard focus', async () => {
            const { user } = setup();
            const reset = screen.getByRole('button', { name: 'Reset Graph' });
            const layout = screen.getByRole('button', { name: 'Layout' });

            await user.hover(reset);
            expect(await screen.findByRole('tooltip', { name: 'Reset Graph' })).toBeVisible();

            await user.unhover(reset);
            act(() => layout.focus());
            expect(await screen.findByRole('tooltip', { name: 'Layout' })).toBeVisible();
        });

        it('exposes menu state and supports keyboard focus restoration', async () => {
            const { user } = setup();
            const layout = screen.getByRole('button', { name: 'Layout' });

            expect(layout).toHaveAttribute('aria-haspopup', 'menu');
            expect(layout).toHaveAttribute('aria-expanded', 'false');

            act(() => layout.focus());
            await user.keyboard('{Enter}');

            const menu = await screen.findByRole('menu');
            expect(menu).toBeVisible();
            expect(layout).toHaveAttribute('aria-expanded', 'true');

            // Radix generates the trigger/menu relationships while the menu is open.
            const menuId = layout.getAttribute('aria-controls');
            expect(menuId).toBeTruthy();
            expect(menu).toHaveAttribute('id', menuId);

            await user.keyboard('{Escape}');

            expect(layout).toHaveAttribute('aria-expanded', 'false');
            expect(layout).toHaveFocus();
        });
    });

    describe('Resetting graph', () => {
        it('calls the onReset prop when the crop button is clicked', async () => {
            const { user } = setup();

            const crop = screen.getByTestId('explore_graph-controls_reset-button');
            await user.click(crop);

            expect(onResetFn).toBeCalled();
        });
    });

    describe('Toggling labels', () => {
        it('calls onToggleNodeLabels when click show all node labels', async () => {
            const { user } = setup();
            const labelMenu = screen.getByTestId('explore_graph-controls_hide-labels-menu');
            await user.click(labelMenu);

            const hideNodeLabels = await screen.findByText('Hide Node Labels');
            await user.click(hideNodeLabels);

            expect(onToggleNodeLabelsFn).toBeCalled();
        });
        it('calls onToggleEdgeLabels when click show all edge labels', async () => {
            const { user } = setup();
            const labelMenu = screen.getByTestId('explore_graph-controls_hide-labels-menu');
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
                const { user } = setup({ showEdgeLabels, showNodeLabels });

                const menuTestId =
                    !showNodeLabels || !showEdgeLabels
                        ? 'explore_graph-controls_show-labels-menu'
                        : 'explore_graph-controls_hide-labels-menu';
                await user.click(screen.getByTestId(menuTestId));

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
            const { user } = setup();

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            const selectedLayout = layoutOptions[0];
            const firstLayout = await screen.findByText(layoutOptions[0]);
            await user.click(firstLayout);

            expect(onLayoutChangeFn).toBeCalledWith(selectedLayout);
        });

        it('does not highlight any layout option on first load when no layout has been manually selected', async () => {
            const { user } = setup({ isExploreLayoutSelected: false });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            for (const option of layoutOptions) {
                const menuItem = await screen.findByTestId(`explore_graph-controls_${option}-buttonLabel`);
                expect(menuItem).not.toHaveClass('!bg-primary');
            }
        });

        it('highlights the selectedLayout when isExploreLayoutSelected is true', async () => {
            const { user } = setup({ isExploreLayoutSelected: true });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            const selected = await screen.findByTestId(`explore_graph-controls_${preselectedLayout}-buttonLabel`);
            expect(selected).toHaveClass('!bg-primary');

            const otherOptions = layoutOptions.filter((option) => option !== preselectedLayout);
            for (const option of otherOptions) {
                const menuItem = await screen.findByTestId(`explore_graph-controls_${option}-buttonLabel`);
                expect(menuItem).not.toHaveClass('!bg-primary');
            }
        });

        it('calls onLayoutChange with the same layout when the currently selected option is clicked, enabling de-selection', async () => {
            const { user } = setup({ isExploreLayoutSelected: true });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            const selected = await screen.findByTestId(`explore_graph-controls_${preselectedLayout}-buttonLabel`);
            await user.click(selected);

            expect(onLayoutChangeFn).toBeCalledWith(preselectedLayout);
        });

        it('reverts all menu options to an unselected state when isExploreLayoutSelected becomes false', async () => {
            // After a user de-selects a previously selected layout, isExploreLayoutSelected is false even though
            // selectedLayout may still be present. No option should appear selected.
            const { user } = setup({ isExploreLayoutSelected: false, selectedLayout: preselectedLayout });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            for (const option of layoutOptions) {
                const menuItem = await screen.findByTestId(`explore_graph-controls_${option}-buttonLabel`);
                expect(menuItem).not.toHaveClass('!bg-primary');
            }
        });
    });

    describe('Table view selection state', () => {
        const layoutOptionsWithTable = ['sequential', 'standard', 'table'] as const;

        it('highlights the table option instead of selectedLayout when table is selected on a cypher query', async () => {
            const { user } = setup({
                layoutOptionsOverride: layoutOptionsWithTable,
                selectedLayout: 'sequential',
                isExploreLayoutSelected: true,
                isExploreTableSelected: true,
                route: '/?searchType=cypher',
            });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            const tableItem = await screen.findByTestId('explore_graph-controls_table-buttonLabel');
            expect(tableItem).toHaveClass('!bg-primary');

            const sequentialItem = await screen.findByTestId('explore_graph-controls_sequential-buttonLabel');
            expect(sequentialItem).not.toHaveClass('!bg-primary');
        });

        it('does not highlight the table option when isExploreTableSelected is true but isExploreLayoutSelected is false', async () => {
            const { user } = setup({
                layoutOptionsOverride: layoutOptionsWithTable,
                selectedLayout: 'sequential',
                isExploreLayoutSelected: false,
                isExploreTableSelected: true,
                route: '/?searchType=cypher',
            });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            for (const option of layoutOptionsWithTable) {
                const menuItem = await screen.findByTestId(`explore_graph-controls_${option}-buttonLabel`);
                expect(menuItem).not.toHaveClass('!bg-primary');
            }
        });

        it('highlights selectedLayout rather than the table option when searchType is not cypher', async () => {
            const { user } = setup({
                layoutOptionsOverride: layoutOptionsWithTable,
                selectedLayout: 'sequential',
                isExploreLayoutSelected: true,
                isExploreTableSelected: true,
                route: '/?searchType=node',
            });

            const layoutMenu = screen.getByTestId('explore_graph-controls_layout-menu');
            await user.click(layoutMenu);

            const sequentialItem = await screen.findByTestId('explore_graph-controls_sequential-buttonLabel');
            expect(sequentialItem).toHaveClass('!bg-primary');

            const tableItem = await screen.findByTestId('explore_graph-controls_table-buttonLabel');
            expect(tableItem).not.toHaveClass('!bg-primary');
        });
    });
    describe('Exporting json', () => {
        it('renders only the JSON action when no additional actions are supplied', async () => {
            const { user } = setup();

            await user.click(screen.getByRole('button', { name: 'Export' }));

            const menuItems = await screen.findAllByRole('menuitem');
            expect(menuItems).toHaveLength(1);
            expect(menuItems[0]).toHaveAccessibleName('JSON');
        });

        it('disables the JSON button if the JSON is empty', async () => {
            const { user } = setup({ json: {} });

            const exportMenu = screen.getByTestId('explore_graph-controls_export-menu');
            await user.click(exportMenu);

            const jsonButton = await screen.findByRole('menuitem', { name: 'JSON' });

            expect(jsonButton).toHaveAttribute('aria-disabled', 'true');
        });

        it('calls exportToJson util when valid non a empty object is passed as the jsonData prop', async () => {
            exportToJsonSpy.mockImplementationOnce(() => undefined);

            const json = { test: 'data' };
            const { user } = setup({ json });

            const exportMenu = screen.getByTestId('explore_graph-controls_export-menu');
            await user.click(exportMenu);

            const jsonButton = await screen.findByText('JSON');
            await user.click(jsonButton);

            expect(exportToJsonSpy).toBeCalledWith(json);
        });

        it('renders additional actions before JSON in the supplied order', async () => {
            const { user } = setup({ additionalExportActions });

            await user.click(screen.getByRole('button', { name: 'Export' }));

            const menuItems = await screen.findAllByRole('menuitem');
            expect(menuItems.map((item) => item.textContent)).toEqual([
                'Additional action one',
                'Additional action two',
                'JSON',
            ]);
        });

        it('invokes the selected action, closes the menu, and restores focus', async () => {
            const { user } = setup({ additionalExportActions });
            const exportMenu = screen.getByRole('button', { name: 'Export' });

            await user.click(exportMenu);
            await user.click(await screen.findByRole('menuitem', { name: 'Additional action one' }));

            expect(onAdditionalExportActionOne).toHaveBeenCalledOnce();
            expect(onAdditionalExportActionTwo).not.toHaveBeenCalled();
            expect(exportMenu).toHaveAttribute('aria-expanded', 'false');
            expect(exportMenu).toHaveFocus();
        });

        it('exposes disabled additional actions and prevents activation', async () => {
            const disabledExportActions = [{ ...additionalExportActions[0], disabled: true }];
            const { user } = setup({ additionalExportActions: disabledExportActions });
            const exportMenu = screen.getByRole('button', { name: 'Export' });

            await user.click(exportMenu);
            const disabledAction = await screen.findByRole('menuitem', { name: 'Additional action one' });

            expect(disabledAction).toHaveAttribute('aria-disabled', 'true');
            act(() => disabledAction.focus());
            await user.keyboard('{Enter}');
            expect(onAdditionalExportActionOne).not.toHaveBeenCalled();
            expect(exportMenu).toHaveAttribute('aria-expanded', 'true');
        });

        it('supports keyboard activation of additional actions', async () => {
            const { user } = setup({ additionalExportActions: additionalExportActions.slice(0, 1) });
            const exportMenu = screen.getByRole('button', { name: 'Export' });

            act(() => exportMenu.focus());
            await user.keyboard('{Enter}');
            const additionalAction = await screen.findByRole('menuitem', { name: 'Additional action one' });
            expect(additionalAction).toHaveFocus();

            await user.keyboard('{Enter}');

            expect(onAdditionalExportActionOne).toHaveBeenCalledOnce();
            expect(exportMenu).toHaveAttribute('aria-expanded', 'false');
            expect(exportMenu).toHaveFocus();
        });
    });
    describe('Searching current results', () => {
        it('renders the search button', async () => {
            setup();
            const searchResultsMenu = await screen.findByTestId('explore_graph-controls_search-current-results');

            expect(searchResultsMenu).toBeInTheDocument();
        });

        it('disables the search button when isCurrentSearchOpen is true', async () => {
            const { user } = setup();

            const searchResultsMenu = screen.getByTestId('explore_graph-controls_search-current-results');
            await user.click(searchResultsMenu);

            expect(searchResultsMenu).toBeDisabled();
        });

        it('shows the search panel when isCurrentSearchOpen is true', async () => {
            const { user } = setup();

            expect(screen.queryByTestId('explore_graph-controls_search-current-nodes-panel')).not.toBeInTheDocument();

            const searchResultsMenu = screen.getByTestId('explore_graph-controls_search-current-results');
            await user.click(searchResultsMenu);

            expect(screen.getByTestId('explore_graph-controls_search-current-nodes-panel')).toBeInTheDocument();
        });

        it('opens when keyboard shortcut is pressed', async () => {
            const { user } = setup();

            expect(screen.queryByTestId('explore_graph-controls_search-current-nodes-panel')).not.toBeInTheDocument();

            await user.keyboard('{Alt>}{Shift>}[Slash]{/Shift}{/Alt}');

            expect(screen.getByTestId('explore_graph-controls_search-current-nodes-panel')).toBeInTheDocument();
        });

        it('sets the selectedItem param and closes the search panel when a node is selected', async () => {
            const { user } = setup();

            const searchResultsMenu = screen.getByTestId('explore_graph-controls_search-current-results');

            await user.click(searchResultsMenu);

            const searchInput = await screen.findByPlaceholderText('Search node in results');

            await user.type(searchInput, currentNodes[1].label);

            const searchedNode = await screen.findByTestId('explore_search_result-list-item');
            await act(async () => {
                await user.click(searchedNode);
            });

            expect(onSearchedNodeClickFn).toBeCalled();

            expect(screen.queryByTestId('explore_graph-controls_search-current-nodes-panel')).not.toBeInTheDocument();
        });

        it('restores focus to the search button after the panel closes', async () => {
            const { user } = setup();

            const searchResultsMenu = screen.getByTestId('explore_graph-controls_search-current-results');
            await user.click(searchResultsMenu);

            expect(screen.getByTestId('explore_graph-controls_search-current-nodes-panel')).toBeInTheDocument();

            await user.keyboard('{Escape}');

            expect(screen.queryByTestId('explore_graph-controls_search-current-nodes-panel')).not.toBeInTheDocument();
            await waitFor(() => expect(searchResultsMenu).toHaveFocus());
        });
    });
});
