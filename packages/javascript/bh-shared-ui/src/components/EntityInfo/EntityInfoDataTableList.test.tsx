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
import { NodeDetails, SeedTypeCypher } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import * as hooks from '../../hooks';
import { mockSourceKindsHandler, zoneHandlers } from '../../mocks';
import { zonesPath } from '../../routes';
import { render, screen } from '../../test-utils';
import { ObjectInfoPanelContextProvider } from '../../views';
import EntitysRulesInformation from '../../views/PrivilegeZones/Details/EntityRulesInformation';
import { EntityInfoDataTable } from '../EntityInfoDataTable';
import EntityInfoContent from './EntityInfoContent';

const testSelector = {
    id: 444,
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

const server = setupServer(
    ...zoneHandlers,
    rest.get('/api/v2/asset-group-tags/:tagId/members/:memberId', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: testSelector,
            })
        );
    }),
    rest.get('/api/v2/nodes/:id', (_, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    node_id: 7,
                    kinds: [{ name: 'User', node_kind_id: 1 }],
                    properties: { objectid: 'test-user' },
                },
            })
        );
    }),
    mockSourceKindsHandler()
);

const EntityInfoContentWithProvider = ({
    selectedNode,
    additionalTables,
}: {
    selectedNode: NodeDetails;
    additionalTables?: {
        sectionProps: any;
        TableComponent: React.FC<any>;
    }[];
}) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoContent
            DataTable={EntityInfoDataTable}
            additionalTables={additionalTables}
            selectedNode={selectedNode}
        />
    </ObjectInfoPanelContextProvider>
);

vi.mock('../../hooks', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        useExploreParams: vi.fn(),
        usePZQueryParams: vi.fn(),
    };
});

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoDataTableList', () => {
    it('Displays the rules list if passed in through additional sections', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.User;
        const selectedNode = {
            node_id: 1,
            kinds: [{ name: nodeType, node_kind_id: 1 }],
            properties: { objectid: testId },
        };

        vi.mocked(hooks.useExploreParams).mockReturnValue({ selectedItem: '7' } as unknown as ReturnType<
            typeof hooks.useExploreParams
        >);
        vi.mocked(hooks.usePZQueryParams).mockReturnValue({ assetGroupTagId: 1 } as unknown as ReturnType<
            typeof hooks.usePZQueryParams
        >);

        render(
            <EntityInfoContentWithProvider
                selectedNode={selectedNode}
                additionalTables={[
                    {
                        sectionProps: { tagType: zonesPath },
                        TableComponent: EntitysRulesInformation,
                    },
                ]}
            />
        );

        expect(await screen.findByText('Rules')).toBeInTheDocument();
    });

    it('Hides selector information if additionalSections is false', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.User;
        const selectedNode = {
            node_id: 1,
            kinds: [{ name: nodeType, node_kind_id: 1 }],
            properties: { objectid: testId },
        };

        render(<EntityInfoContentWithProvider selectedNode={selectedNode} />);

        const selectorsInfoSectionTitle = await screen.queryByText(/rules/i);
        expect(selectorsInfoSectionTitle).not.toBeInTheDocument();
    });
});
