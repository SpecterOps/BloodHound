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
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { zoneHandlers } from '../../mocks';
import { render, screen, waitForElementToBeRemoved } from '../../test-utils';
import { EntityInfoDataTableProps, EntityKinds } from '../../utils';
import { ObjectInfoPanelContextProvider } from '../../views';
import EntitySelectorsInformation from '../../views/ZoneManagement/Details/EntitySelectorsInformation';
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
    })
);

const EntityInfoContentWithProvider = ({
    testId,
    nodeType,
    databaseId,
    additionalTables,
}: {
    testId: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
    additionalTables?: {
        sectionProps: EntityInfoDataTableProps;
        TableComponent: React.FC<EntityInfoDataTableProps>;
    }[];
}) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoContent
            DataTable={EntityInfoDataTable}
            id={testId}
            nodeType={nodeType}
            databaseId={databaseId}
            additionalTables={additionalTables}
        />
    </ObjectInfoPanelContextProvider>
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoDataTableList', () => {
    it('Displays selector information if additionalSections is true', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.User;

        render(
            <EntityInfoContentWithProvider
                testId={testId}
                nodeType={nodeType}
                additionalTables={[
                    {
                        sectionProps: { label: 'Selectors', id: '1' },
                        TableComponent: EntitySelectorsInformation,
                    },
                ]}
            />
        );

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        screen.debug(undefined, Infinity);

        const selectorsInfoSectionTitle = await screen.findByText(/selectors/i);
        expect(selectorsInfoSectionTitle).toBeInTheDocument();
    });

    it('Hides selector information if additionalSections is false', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.User;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        const selectorsInfoSectionTitle = await screen.queryByText(/selectors/i);
        expect(selectorsInfoSectionTitle).not.toBeInTheDocument();
    });
});
