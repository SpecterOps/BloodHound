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
import { SeedTypeCypher } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { zoneHandlers } from '../../mocks';
import { render, screen, waitForElementToBeRemoved } from '../../test-utils';
import { EntityKinds } from '../../utils';
import { ObjectInfoPanelContextProvider } from '../../views';
import EntityInfoContent from './EntityInfoContent';

const server = setupServer(
    rest.get('/api/v2/azure/roles', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    kind: 'AZRole',
                    props: {},
                    active_assignments: 0,
                    pim_assignments: 0,
                },
            })
        );
    }),
    rest.get('/api/v2/base/*', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    kind: ActiveDirectoryNodeKind.Entity,
                    props: { objectid: 'test' },
                },
            })
        );
    }),
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    nodes: {
                        '42': {
                            kind: 'Unknown',
                            properties: { objectid: 'unknown kind' },
                        },
                    },
                },
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

const EntityInfoContentWithProvider = ({
    testId,
    nodeType,
    databaseId,
    zoneManagement,
}: {
    testId: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
    zoneManagement?: boolean;
}) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoContent id={testId} nodeType={nodeType} databaseId={databaseId} zoneManagement={zoneManagement} />
    </ObjectInfoPanelContextProvider>
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoContent', () => {
    it('AZRole information panel will not display a section for PIM Assignments', async () => {
        const testId = '1';
        const nodeType = AzureNodeKind.Role;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);
        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));
        expect(screen.queryByText('PIM Assignments')).not.toBeInTheDocument();
    });
});

describe('EntityObjectInformation', () => {
    it('Calls the `Base` endpoint for a LocalGroup type node', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalGroup;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('test')).toBeInTheDocument();
    });

    it('Calls the `Base` endpoint for a LocalUser type node', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('test')).toBeInTheDocument();
    });

    it('Calls the cypher search endpoint for a node with a type that is not in our schema', async () => {
        const testId = '1';
        const nodeType = 'Unknown';
        const databaseId = '42';

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} databaseId={databaseId} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('unknown kind')).toBeInTheDocument();
    });
});

describe('EntityInfoDataTableList', () => {
    it('Renders Selectors collapsible section if zone management is true', async () => {
        const testId = '2';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;
        const databaseId = '42';

        const testSelector = {
            id: 777,
            asset_group_tag_id: 1,
            name: 'foo',
            allow_disable: true,
            description: 'bar',
            is_default: false,
            auto_certify: true,
            created_at: '2024-10-05T17:54:32.245Z',
            created_by: 'Stephen64@gmail.com',
            updated_at: '2024-07-20T11:22:18.219Z',
            updated_by: 'Donna13@yahoo.com',
            disabled_at: '2024-09-15T09:55:04.177Z',
            disabled_by: 'Roberta_Morar72@hotmail.com',
            count: 3821,
            seeds: [{ selector_id: 777, type: SeedTypeCypher, value: 'match(n) return n limit 5' }],
        };

        const testNodes = [
            {
                name: 'bar',
                objectid: '777',
                type: 'Bat',
            },
        ];
        const testSearchResults = {
            data: testNodes,
        };

        const handlers = [
            ...zoneHandlers,
            rest.get('/api/v2/asset-group-tags/:tagId/selectors/777', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: testSelector,
                    })
                );
            }),
            rest.post(`/api/v2/asset-group-tags/preview-selectors`, (_, res, ctx) => {
                return res(ctx.json({ data: { members: [] } }));
            }),
            rest.post(`/api/v2/graphs/cypher`, (_, res, ctx) => {
                return res(ctx.json({ data: { nodes: {}, edges: [] } }));
            }),
            rest.get(`/api/v2/search`, (_, res, ctx) => {
                return res(ctx.json(testSearchResults));
            }),
        ];

        server.use(...handlers);
        render(
            <EntityInfoContentWithProvider testId={testId} nodeType={nodeType} databaseId={databaseId} zoneManagement />
        );

        //await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('Selectors')).toBeInTheDocument();
    });

    it('Renders Selectors collapsible section if zone management is true', async () => {
        const testId = '2';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;
        const databaseId = '42';

        const testSelector = {
            id: 777,
            asset_group_tag_id: 1,
            name: 'foo',
            allow_disable: true,
            description: 'bar',
            is_default: false,
            auto_certify: true,
            created_at: '2024-10-05T17:54:32.245Z',
            created_by: 'Stephen64@gmail.com',
            updated_at: '2024-07-20T11:22:18.219Z',
            updated_by: 'Donna13@yahoo.com',
            disabled_at: '2024-09-15T09:55:04.177Z',
            disabled_by: 'Roberta_Morar72@hotmail.com',
            count: 3821,
            seeds: [{ selector_id: 777, type: SeedTypeCypher, value: 'match(n) return n limit 5' }],
        };

        const testNodes = [
            {
                name: 'bar',
                objectid: '777',
                type: 'Bat',
            },
        ];
        const testSearchResults = {
            data: testNodes,
        };

        const handlers = [
            ...zoneHandlers,
            rest.get('/api/v2/asset-group-tags/:tagId/selectors/777', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: testSelector,
                    })
                );
            }),
            rest.post(`/api/v2/asset-group-tags/preview-selectors`, (_, res, ctx) => {
                return res(ctx.json({ data: { members: [] } }));
            }),
            rest.post(`/api/v2/graphs/cypher`, (_, res, ctx) => {
                return res(ctx.json({ data: { nodes: {}, edges: [] } }));
            }),
            rest.get(`/api/v2/search`, (_, res, ctx) => {
                return res(ctx.json(testSearchResults));
            }),
        ];

        server.use(...handlers);
        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} databaseId={databaseId} />);

        //await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('Selectors')).not.toBeInTheDocument();
    });
});
