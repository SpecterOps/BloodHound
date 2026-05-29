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
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { render, screen, waitForElementToBeRemoved } from '../../test-utils';
import { EntityKinds } from '../../utils';
import { ObjectInfoPanelContextProvider } from '../../views';
import { EntityInfoDataTableGraphed } from '../EntityInfoDataTableGraphed/EntityInfoDataTableGraphed';
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
    }),
    rest.get('/api/v2/asset-group-tags', (req, res, ctx) => {
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
}: {
    testId: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
}) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoContent
            id={testId}
            nodeType={nodeType}
            databaseId={databaseId}
            DataTable={EntityInfoDataTableGraphed}
        />
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

        const objectInfoSectionTitle = await screen.findByText(/object information/i);
        expect(objectInfoSectionTitle).toBeInTheDocument();

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

    it('Calls the cypher search endpoint for a node that is hidden from user with not all environments access', async () => {
        const testId = '1';
        const nodeType = 'HIDDEN';

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);

        expect(
            await screen.findByText(
                'This object’s information is not disclosed. Please contact your admin in order to get access.'
            )
        ).toBeInTheDocument();
    });
});
