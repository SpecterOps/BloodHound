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
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { render, screen } from '../../test-utils';
import EntityInfoDataTableList from './EntityInfoDataTableList';

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
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    rest.get(`/api/v2/asset-group-tags/*`, async (req, res, ctx) => {
        return res(ctx.json({ data: { tag: { id: 777 } } }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoDataTableList', () => {
    it('Renders the entity info data table list', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;

        render(<EntityInfoDataTableList id={testId} nodeType={nodeType} zoneManagement />);
        screen.debug(undefined, Infinity);

        expect(screen.getByTestId('entity-info-data-table-list')).toBeInTheDocument();
    });

    it.skip('Includes the Selectors section if zone management page is true', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;

        render(<EntityInfoDataTableList id={testId} nodeType={nodeType} zoneManagement />);

        expect(screen.getByTestId('entity-info-data-table-list')).toBeInTheDocument();
        expect(await screen.findByText('Selectors')).toBeInTheDocument();
    });

    it.skip('Excludes the Selectors section if zone management page is false', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;

        render(<EntityInfoDataTableList id={testId} nodeType={nodeType} />);

        expect(screen.getByTestId('entity-info-data-table-list')).toBeInTheDocument();
        expect(await screen.findByText('Selectors')).not.toBeInTheDocument();
    });
});
