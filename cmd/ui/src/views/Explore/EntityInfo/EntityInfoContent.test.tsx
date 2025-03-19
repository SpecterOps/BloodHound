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

import { ActiveDirectoryNodeKind, AzureNodeKind } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitForElementToBeRemoved } from 'src/test-utils';
import EntityInfoContent from './EntityInfoContent';
import { EntityInfoPanelContextProvider } from './EntityInfoPanelContextProvider';

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

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoContent', () => {
    it('AZRole information panel will not display a section for PIM Assignments', async () => {
        const testId = '1';

        render(
            <EntityInfoPanelContextProvider>
                <EntityInfoContent id={testId} nodeType={AzureNodeKind.Role} />
            </EntityInfoPanelContextProvider>
        );
        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));
        expect(screen.queryByText('PIM Assignments')).not.toBeInTheDocument();
    });
});

describe('EntityObjectInformation', () => {
    it('Calls the `Base` endpoint for a LocalGroup type node', async () => {
        const testId = '1';

        render(
            <EntityInfoPanelContextProvider>
                <EntityInfoContent id={testId} nodeType={ActiveDirectoryNodeKind.LocalGroup} />
            </EntityInfoPanelContextProvider>
        );

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('test')).toBeInTheDocument();
    });

    it('Calls the `Base` endpoint for a LocalUser type node', async () => {
        const testId = '1';

        render(
            <EntityInfoPanelContextProvider>
                <EntityInfoContent id={testId} nodeType={ActiveDirectoryNodeKind.LocalUser} />
            </EntityInfoPanelContextProvider>
        );

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('test')).toBeInTheDocument();
    });

    it('Calls the cypher search endpoint for a node with a type that is not in our schema', async () => {
        const testId = '1';

        render(
            <EntityInfoPanelContextProvider>
                <EntityInfoContent id={testId} nodeType={'Unknown'} databaseId='42' />
            </EntityInfoPanelContextProvider>
        );

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('unknown kind')).toBeInTheDocument();
    });
});
