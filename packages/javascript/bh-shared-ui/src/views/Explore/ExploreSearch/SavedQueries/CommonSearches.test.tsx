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

import { render, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { QueryClient, QueryClientProvider } from 'react-query';
import { vi } from 'vitest';
import { SavedQueriesContext } from '../../providers';
import CommonSearches from './CommonSearches';

const extensions = [
    { id: 42, name: 'asset_explorer', display_name: 'Asset Explorer' },
    { id: 43, name: 'Other Extension', display_name: '' },
];

const server = setupServer(
    rest.get('/api/v2/saved-queries', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n) return n limit 5',
                        name: 'me save a query 1',
                        id: 1,
                    },
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n) return n limit 5',
                        name: 'me save a query 2',
                        id: 2,
                        category: 'User Reports',
                    },
                    {
                        user_id: '00000000-0000-0000-0000-000000000000',
                        query: 'match (n:CustomAsset) return n',
                        name: 'Find Custom Assets',
                        description: 'Returns assets supplied by an OpenGraph extension',
                        category: 'Asset Management',
                        extension_id: 42,
                        id: 3,
                    },
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n:Other) return n',
                        name: 'Find Other Assets',
                        category: 'Asset Management',
                        extension_id: 43,
                        id: 4,
                    },
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n:Other) return n',
                        name: 'Find Uncategorized Assets',
                        extension_id: 42,
                        id: 5,
                    },
                ],
            })
        );
    }),
    rest.get('/api/v2/extensions', (_req, res, ctx) =>
        res(
            ctx.json({
                data: {
                    extensions,
                },
            })
        )
    ),
    rest.delete('/api/v2/saved-queries/:id', (req, res, ctx) => {
        return res(ctx.status(201));
    }),
    rest.get('/api/v2/features', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 16, key: 'tier_management_engine', enabled: true }],
            })
        );
    }),
    rest.get('/api/v2/self', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: { id: '4e09c965-65bd-4f15-ae71-5075a6fed14b' },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});
afterAll(() => server.close());

let queryClient: QueryClient;

beforeEach(() => {
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
});

describe('CommonSearches', () => {
    it('renders headers', async () => {
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );

        const header = screen.getByText(/Saved Queries/i);
        expect(header).toBeInTheDocument();
    });

    it('should display filter dropwdowns', async () => {
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );

        const platformLabel = await screen.findByLabelText(/platforms/i);
        const categoriesLabel = await screen.findByLabelText(/categories/i);
        const sourceLabel = await screen.findByLabelText(/source/i);
        expect(platformLabel).toBeInTheDocument();
        expect(categoriesLabel).toBeInTheDocument();
        expect(sourceLabel).toBeInTheDocument();
    });

    it('renders a filter search and platform dropdown menu', async () => {
        server.use(
            rest.get('/api/v2/extensions', (_req, res, ctx) => res(ctx.delay(200), ctx.json({ data: { extensions } })))
        );
        const user = userEvent.setup();

        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={false}
                />
            </QueryClientProvider>
        );

        const testSearch = await screen.findByPlaceholderText('Search');
        expect(testSearch).toBeInTheDocument();
        expect(testSearch).toHaveValue('');
        const testPlatforms = await screen.findByLabelText(/platform/i);
        expect(testPlatforms).toBeInTheDocument();
        await user.click(testPlatforms);
        const testListBox = await screen.findByRole('listbox');
        expect(testListBox).toBeInTheDocument();
        expect(testListBox).toBeVisible();

        expect(await within(testListBox).findByRole('option', { name: 'Asset Explorer' })).toBeInTheDocument();
        expect(within(testListBox).getByRole('option', { name: 'Other Extension' })).toBeInTheDocument();

        await user.click(within(testListBox).getByRole('option', { name: 'All' }));

        expect(screen.getByText(/all domain admins/i)).toBeInTheDocument();
    });

    it('displays correct content based on platform filter Azure', async () => {
        const user = userEvent.setup();

        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={false}
                />
            </QueryClientProvider>
        );

        const testPlatforms = await screen.findByLabelText(/platform/i);
        expect(testPlatforms).toBeInTheDocument();
        await user.click(testPlatforms);
        const testListBox = await screen.findByRole('listbox');
        expect(testListBox).toBeInTheDocument();
        expect(testListBox).toBeVisible();

        //select Azure
        await user.click(within(testListBox).getByRole('option', { name: 'Azure' }));

        //Azure query present
        expect(screen.getByText(/All members of high privileged roles/i)).toBeInTheDocument();

        //AD query not present
        const adText = screen.queryByText(/all domain admins/i);
        expect(adText).toBeNull();
    });

    it('displays correct content based on platform filter AD', async () => {
        const user = userEvent.setup();

        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={false}
                />
            </QueryClientProvider>
        );

        const testPlatforms = await screen.findByLabelText(/platform/i);
        expect(testPlatforms).toBeInTheDocument();
        await user.click(testPlatforms);
        const testListBox = await screen.findByRole('listbox');

        //select AD
        await user.click(within(testListBox).getByRole('option', { name: 'Active Directory' }));

        //AD query present
        expect(screen.getByText(/all domain admins/i)).toBeInTheDocument();

        //Axure query not present
        const adText = screen.queryByText(/All members of high privileged roles/i);
        expect(adText).toBeNull();
    });

    it.each(['', undefined])(
        'groups extension queries using display names and falls back when display_name is %s',
        async (displayName) => {
            server.use(
                rest.get('/api/v2/extensions', (_req, res, ctx) =>
                    res(
                        ctx.json({
                            data: {
                                extensions: [extensions[0], { ...extensions[1], display_name: displayName }],
                            },
                        })
                    )
                )
            );
            const user = userEvent.setup();

            const screen = render(
                <QueryClientProvider client={queryClient}>
                    <CommonSearches
                        onSetCypherQuery={vi.fn()}
                        onPerformCypherSearch={vi.fn()}
                        onToggleCommonQueries={vi.fn()}
                        showCommonQueries={true}
                    />
                </QueryClientProvider>
            );

            expect(await screen.findByText('Asset Explorer')).toBeInTheDocument();
            expect(screen.getByText('Find Custom Assets')).toBeInTheDocument();
            expect(screen.getAllByText('Asset Management')).toHaveLength(2);
            expect(screen.getAllByText('Uncategorized')).toHaveLength(2);
            expect(screen.getByText('Other Extension')).toBeInTheDocument();

            await user.click(screen.getByLabelText('Source'));
            await user.click(await screen.findByRole('option', { name: 'Extension' }));

            expect(screen.getByText('Find Custom Assets')).toBeInTheDocument();
            expect(screen.queryByText(/all domain admins/i)).not.toBeInTheDocument();
            expect(screen.queryByTestId('saved-query-action-menu-trigger')).not.toBeInTheDocument();
        }
    );

    it('filters extension platforms, limits categories to that platform, and clears incompatible selections', async () => {
        const user = userEvent.setup();
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={vi.fn()}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );

        expect(await screen.findByText('Find Custom Assets')).toBeInTheDocument();
        await user.click(screen.getByLabelText('Platforms'));
        await user.click(await screen.findByRole('option', { name: 'Asset Explorer' }));
        expect(screen.getByText('Find Custom Assets')).toBeInTheDocument();
        expect(screen.getByText('Find Uncategorized Assets')).toBeInTheDocument();
        expect(screen.queryByText('Find Other Assets')).not.toBeInTheDocument();
        expect(screen.queryByText('me save a query 1')).not.toBeInTheDocument();

        await user.click(screen.getByLabelText('Categories'));
        expect(screen.getByRole('option', { name: 'Asset Management' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: 'Uncategorized' })).toBeInTheDocument();
        expect(screen.queryByRole('option', { name: 'User Reports' })).not.toBeInTheDocument();
        await user.click(screen.getByRole('option', { name: 'Uncategorized' }));
        expect(screen.getByText('Find Uncategorized Assets')).toBeInTheDocument();
        expect(screen.queryByText('Find Custom Assets')).not.toBeInTheDocument();

        await user.click(screen.getByLabelText('Platforms'));
        await user.click(screen.getByRole('option', { name: 'Saved Queries' }));
        expect(screen.getByText('me save a query 1')).toBeInTheDocument();
        expect(screen.getByText('me save a query 2')).toBeInTheDocument();
    });

    it('prevents deleting selected extension queries', async () => {
        const user = userEvent.setup();
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <SavedQueriesContext.Provider
                    value={{
                        selected: { query: 'match (n:CustomAsset) return n', id: 3 },
                        selectedQuery: {
                            id: 3,
                            name: 'Find Custom Assets',
                            description: '',
                            query: 'match (n:CustomAsset) return n',
                            schema_extension_id: 42,
                            canEdit: true,
                        },
                        showSaveQueryDialog: false,
                        saveAction: undefined,
                        setSelected: vi.fn(),
                        setShowSaveQueryDialog: vi.fn(),
                        setSaveAction: vi.fn(),
                        runQuery: vi.fn(),
                        editQuery: vi.fn(),
                    }}>
                    <CommonSearches
                        onSetCypherQuery={vi.fn()}
                        onPerformCypherSearch={vi.fn()}
                        onToggleCommonQueries={vi.fn()}
                        showCommonQueries={true}
                    />
                </SavedQueriesContext.Provider>
            </QueryClientProvider>
        );

        await user.click(await screen.findByText('Find Custom Assets'));
        expect(screen.queryByRole('button', { name: 'delete' })).not.toBeInTheDocument();
        expect(screen.queryByTestId('saved-query-action-menu-trigger')).not.toBeInTheDocument();
        expect(
            within(screen.getByTestId('list-sections')).getAllByRole('button', { name: /Extension-managed query/ })
        ).toHaveLength(3);
        await user.hover(screen.getByRole('button', { name: 'Extension-managed query: Find Custom Assets' }));
        expect(await screen.findByRole('tooltip')).toHaveTextContent('managed by an extension');
    });

    //Toggle switch - test visibility
    it('handles chevron click event', async () => {
        const user = userEvent.setup();
        const handleToggle = vi.fn();
        const screen = render(
            <QueryClientProvider client={queryClient}>
                <CommonSearches
                    onSetCypherQuery={vi.fn()}
                    onPerformCypherSearch={vi.fn()}
                    onToggleCommonQueries={handleToggle}
                    showCommonQueries={true}
                />
            </QueryClientProvider>
        );
        const queriesToggle = screen.getByTestId('common-queries-toggle');
        await user.click(queriesToggle);
        expect(handleToggle).toBeCalled();
        expect(handleToggle).toBeCalledTimes(1);
        expect(screen.getByText(/chevron-up/i)).toBeInTheDocument();
    });
});
