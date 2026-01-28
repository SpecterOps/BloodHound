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
import { render } from '../../../test-utils';
import { mockCodemirrorLayoutMethods } from '../../../utils';
import CypherSearch from './CypherSearch';

const CYPHER = 'match (n) return n limit 5';
const INCOMPLETE_CYPHER = 'match (n:';

describe('CypherSearch', () => {
    const testPerformSearch = vi.fn();

    const testState = {
        cypherQuery: '',
        setCypherQuery: vi.fn(),
        performSearch: testPerformSearch,
    };
    const setup = async (state = testState) => {
        const autoRun = true;
        const handleAutoRun = () => {};
        const testOnRunSearchClick = vi.fn();

        const screen = render(<CypherSearch cypherSearchState={state} autoRun={autoRun} setAutoRun={handleAutoRun} />);
        const user = userEvent.setup();

        return { state, screen, user, testOnRunSearchClick };
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
        rest.get('/api/v2/bloodhound-users-minimal', (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {
                        users: [],
                    },
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
        expect(screen.getByRole('link', { name: /Learn more about cypher/i })).toBeInTheDocument();
    });

    it('should call the setCypherQuery handler when the value in the editor changes', async () => {
        const { screen, user, state } = await setup();
        const searchbox = screen.getAllByRole('textbox');

        await user.type(searchbox[1], CYPHER);

        expect(state.setCypherQuery).toBeCalled();
        expect(state.setCypherQuery).toHaveBeenCalledTimes(CYPHER.length);
    });

    it('should display a dropdown when a user types a partial query that can be autocompleted', async () => {
        const { screen, user } = await setup();
        const searchbox = screen.getAllByRole('textbox');

        await user.type(searchbox[1], INCOMPLETE_CYPHER);

        const autocomplete = await screen.findByRole('listbox');

        expect(autocomplete).toBeVisible();
    });

    it('should call performSearch on keyboard press alt+R', async () => {
        const { user } = await setup({ ...testState, cypherQuery: 'Anything' });

        expect(testPerformSearch).not.toHaveBeenCalled();

        await user.keyboard('{Alt>}r{/Alt}');

        expect(testPerformSearch).toHaveBeenCalled();
    });

    it('should open save dialog on keyboard press alt+S', async () => {
        const { user, screen } = await setup({ ...testState, cypherQuery: 'Anything' });

        expect(screen.queryByTestId('save-query-dialog')).not.toBeInTheDocument();

        await user.keyboard('{Alt>}s{/Alt}');

        expect(screen.queryByTestId('save-query-dialog')).toBeInTheDocument();
    });
});
