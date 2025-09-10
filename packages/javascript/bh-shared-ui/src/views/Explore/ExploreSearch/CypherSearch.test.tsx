// Copyright 2023 Specter Ops, Inc.
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
import { render, within } from '../../../test-utils';
import { mockCodemirrorLayoutMethods } from '../../../utils';
import CypherSearch from './CypherSearch';

const CYPHER = 'match (n) return n limit 5';

describe('CypherSearch', () => {
    const setup = async () => {
        const state = {
            cypherQuery: '',
            setCypherQuery: vi.fn(),
            performSearch: vi.fn(),
        };
        const autoRun = true;
        const handleAutoRun = () => {};
        const handleCypherSearch = vi.fn();

        const screen = await render(
            <CypherSearch cypherSearchState={state} autoRun={autoRun} setAutoRun={handleAutoRun} />
        );
        const user = await userEvent.setup();

        return { state, screen, user, handleCypherSearch };
    };

    const server = setupServer(
        rest.get('/api/v2/graphs/kinds', async (_req, res, ctx) => {
            return res(
                ctx.json({
                    data: { kinds: ['Tier Zero', 'Tier One', 'Tier Two'] },
                })
            );
        }),
        rest.get('/api/v2/features', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [{ id: 1, key: 'tier_management_engine', enabled: true }],
                })
            );
        }),
        rest.get('/api/v2/config', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [
                        {
                            key: 'analysis.tiering',
                            name: 'Multi-Tier Analysis Configuration',
                            value: {
                                tier_limit: 3,
                                label_limit: 10,
                                multi_tier_analysis_enabled: true,
                            },
                            id: 8,
                        },
                    ],
                })
            );
        }),
        rest.get('/api/v2/saved-queries', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [],
                })
            );
        }),
        rest.get('/api/v2/self', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {},
                })
            );
        }),
        rest.get('/api/v2/asset-group-tags', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [],
                })
            );
        }),
        rest.get('/api/v2/bloodhound-users', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [],
                })
            );
        })
    );

    beforeAll(() => {
        server.listen();
    });
    beforeEach(mockCodemirrorLayoutMethods);
    afterEach(vi.restoreAllMocks);
    afterAll(() => {
        server.close();
    });

    it('should render', async () => {
        const { screen } = await setup();
        expect(screen.getByText(/cypher query/i)).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /app-icon-info/i })).toBeInTheDocument();
    });

    it('should call the setCypherQuery handler when the value in the editor changes', async () => {
        const { screen, user, state } = await setup();
        const searchbox = screen.getAllByRole('textbox');
        await user.type(searchbox[1], CYPHER);
        expect(state.setCypherQuery).toBeCalled();
        expect(state.setCypherQuery).toHaveBeenCalledTimes(CYPHER.length);
    });
    it('should call performSearch on click Run with cypher populated', async () => {
        const state = {
            cypherQuery: CYPHER,
            setCypherQuery: vi.fn(),
            performSearch: vi.fn(),
        };
        const autoRun = true;
        const handleAutoRun = () => {};

        const screen = await render(
            <CypherSearch cypherSearchState={state} autoRun={autoRun} setAutoRun={handleAutoRun} />
        );
        const user = await userEvent.setup();
        const saveBtn = screen.getByRole('button', { name: /save/i });
        expect(saveBtn).toBeInTheDocument();

        const runBtn = screen.getByRole('button', { name: /run/i });
        expect(runBtn).toBeInTheDocument();

        await user.click(runBtn);
        expect(state.performSearch).toHaveBeenCalled();
    });

    it('should not performSearh if cypher is not populated', async () => {
        const { screen, user, state } = await setup();
        const runBtn = screen.getByRole('button', { name: /run/i });
        expect(runBtn).toBeInTheDocument();
        await user.click(runBtn);
        expect(state.performSearch).not.toHaveBeenCalled();
    });
    it('should fire save query dialog with Save New flow', async () => {
        const state = {
            cypherQuery: CYPHER,
            setCypherQuery: vi.fn(),
            performSearch: vi.fn(),
        };
        const autoRun = true;
        const handleAutoRun = () => {};

        const screen = await render(
            <CypherSearch cypherSearchState={state} autoRun={autoRun} setAutoRun={handleAutoRun} />
        );
        const user = await userEvent.setup();
        const saveBtn = screen.getByRole('button', { name: /save/i });
        expect(saveBtn).toBeInTheDocument();
        await user.click(saveBtn);
        const saveDialog = screen.getByRole('dialog');

        expect(saveDialog).toBeInTheDocument();

        //Save New Flow
        const queryName = screen.getByRole('textbox', { name: /query name/i });
        const queryDescription = screen.getByRole('textbox', { name: /query description/i });

        //Name and Description are empty
        expect(queryName).toHaveValue('');
        expect(queryDescription).toHaveValue('');

        const dialogContainer = within(saveDialog);
        const searchbox = dialogContainer.getAllByRole('textbox');
        //Cypher is populated
        expect(searchbox[2]).toHaveTextContent(CYPHER);
    });

    it('should fire the Save Query modal with the Save As New flow', async () => {
        const state = {
            cypherQuery: CYPHER,
            setCypherQuery: vi.fn(),
            performSearch: vi.fn(),
        };
        const autoRun = true;
        const handleAutoRun = () => {};

        const screen = await render(
            <CypherSearch cypherSearchState={state} autoRun={autoRun} setAutoRun={handleAutoRun} />
        );
        const user = await userEvent.setup();

        const dropdownBtn = screen.getByRole('button', { name: /app-icon-caret-down/i });

        expect(dropdownBtn).toBeInTheDocument();
        await user.click(dropdownBtn);
        const dropDialog = screen.getByRole('dialog');
        expect(dropDialog).toBeInTheDocument();
        const saveAsBtn = screen.getByText(/save as/i);
        expect(saveAsBtn).toBeInTheDocument();
        await user.click(saveAsBtn);

        const saveDialog = screen.getByRole('dialog');
        expect(saveDialog).toBeInTheDocument();
        screen.debug(saveDialog);

        //Save As New Flow

        //correct title
        expect(screen.getByText(/save as new query/i)).toBeInTheDocument();

        const queryName = screen.getByRole('textbox', { name: /query name/i });
        const queryDescription = screen.getByRole('textbox', { name: /query description/i });

        // Name and Description are empty
        expect(queryName).toHaveValue('');
        expect(queryDescription).toHaveValue('');

        const dialogContainer = within(saveDialog);
        const searchbox = dialogContainer.getAllByRole('textbox');
        // Cypher is populated
        expect(searchbox[2]).toHaveTextContent(CYPHER);
    });
});
