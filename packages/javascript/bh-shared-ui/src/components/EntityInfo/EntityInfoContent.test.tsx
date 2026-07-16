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
import { NodeDetails } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { mockSourceKindsHandler } from '../../mocks';
import { render, screen, waitFor } from '../../test-utils';
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
    }),
    mockSourceKindsHandler()
);

const EntityInfoContentWithProvider = ({ selectedNode }: { selectedNode: NodeDetails }) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoContent selectedNode={selectedNode} DataTable={EntityInfoDataTableGraphed} />
    </ObjectInfoPanelContextProvider>
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoContent', () => {
    it('AZRole information panel will not display a section for PIM Assignments', async () => {
        const testId = '1';
        const nodeType = AzureNodeKind.Role;
        const selectedNode = {
            node_id: 1,
            kinds: [{ name: nodeType, node_kind_id: 1 }],
            properties: { objectid: testId },
        };

        render(<EntityInfoContentWithProvider selectedNode={selectedNode} />);
        // Wait for all section loading spinners to finish before asserting absence,
        // so the assertion runs only after EntityInfoList finishes loading its data tables.
        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
        expect(screen.queryByText('PIM Assignments')).not.toBeInTheDocument();
    });
});
