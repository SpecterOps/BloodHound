// Copyright 2026 Specter Ops, Inc.
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

import { AssetGroupTagMember, SelectorSeedRequest } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { zoneHandlers } from '../../../../mocks';
import { render, screen, within } from '../../../../test-utils';
import { SeedSelectionPreview } from './SeedSelectionPreview';

const TestSeeds: SelectorSeedRequest[] = [
    {
        type: 2,
        value: 'MATCH (n:Group) WHERE n.name = "$141000-LDV9MS90TKNJ@WRAITH.CORP" RETURN n\n',
    },
];

const BothListsResults = [
    {
        id: 1777,
        object_id: 'S-1-5-21-3702535222-3822678775-2090119576-1170',
        environment_id: 'S-1-5-21-3702535222-3822678775-2090119576',
        primary_kind: 'Group',
        name: 'EXCHANGE-ADMINS@WRAITH.CORP',
        source: 2,
        asset_group_tag_id: 1,
    },
    {
        id: 1069,
        object_id: '2A2FEF21-A610-4A28-AAE0-D5D69A8E5BA8',
        environment_id: 'S-1-5-21-3702535222-3822678775-2090119576',
        primary_kind: 'Container',
        name: 'USERS@WRAITH.CORP',
        source: 3,
        asset_group_tag_id: 1,
    },
    {
        id: 1810,
        object_id: 'S-1-5-21-3702535222-3822678775-2090119576-1153',
        environment_id: 'S-1-5-21-3702535222-3822678775-2090119576',
        primary_kind: 'Group',
        name: '$141000-LDV9MS90TKNJ@WRAITH.CORP',
        source: 1,
        asset_group_tag_id: 1,
    },
];

const DirectObjectsResults = [
    {
        id: 1810,
        object_id: 'S-1-5-21-3702535222-3822678775-2090119576-1153',
        environment_id: 'S-1-5-21-3702535222-3822678775-2090119576',
        primary_kind: 'Group',
        name: '$141000-LDV9MS90TKNJ@WRAITH.CORP',
        source: 1,
        asset_group_tag_id: 1,
    },
];

let previewResults: AssetGroupTagMember[] | undefined;

const setPreviewResultsTestData = (previewResultsList: AssetGroupTagMember[]) => {
    previewResults = previewResultsList;
};

const handlers = [
    ...zoneHandlers,
    rest.post(`/api/v2/asset-group-tags/preview-selectors`, (_, res, ctx) => {
        return res(ctx.json({ data: { members: previewResults } }));
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    previewResults = undefined;
});
afterAll(() => server.close());

describe('Seed Selection Results', () => {
    it('shows the empty form message when Object Rule Form is empty', async () => {
        render(<SeedSelectionPreview seeds={[]} ruleType={1} />);
        const emptyMessage = await screen.findByText(/enter object id to see sample results/i);
        expect(emptyMessage).toBeInTheDocument();
    });
    it('shows the empty form message when Cypher Rule Form is empty', async () => {
        render(<SeedSelectionPreview seeds={[]} ruleType={2} />);
        const emptyMessage = await screen.findByText(/enter cypher to see sample results/i);
        expect(emptyMessage).toBeInTheDocument();
    });
    it('shows the direct object and expanded object list when results are present for both', async () => {
        setPreviewResultsTestData(BothListsResults);

        render(<SeedSelectionPreview seeds={TestSeeds} ruleType={2} />);

        const directObjectsListContainer = await screen.findByTestId('pz-rule-preview__direct-objects-list');
        const directObjectsListTitle = await within(directObjectsListContainer).findByText(/direct objects/i);
        const directObjectsList = await within(directObjectsListContainer).findAllByTestId('entity-row');

        const expandedObjectListContainer = await screen.findByTestId('pz-rule-preview__expanded-objects-list');
        const expandedObjectListTitle = await within(expandedObjectListContainer).findByText(/expanded objects/i);
        const expandedObjectList = await within(expandedObjectListContainer).findAllByTestId('entity-row');

        expect(directObjectsListTitle).toBeInTheDocument();
        expect(directObjectsList[0]).toBeInTheDocument();
        expect(expandedObjectListTitle).toBeInTheDocument();
        expect(expandedObjectList[0]).toBeInTheDocument();
    });
    it('shows the direct object list and expanded object list empty, when only direct objects in results', async () => {
        setPreviewResultsTestData(DirectObjectsResults);
        render(<SeedSelectionPreview seeds={TestSeeds} ruleType={2} />);

        const directObjectsListContainer = await screen.findByTestId('pz-rule-preview__direct-objects-list');
        const directObjectsListTitle = await within(directObjectsListContainer).findByText(/direct objects/i);
        const directObjectsList = await within(directObjectsListContainer).findAllByTestId('entity-row');

        const expandedObjectsListContainer = await screen.findByTestId('pz-rule-preview__expanded-objects-list');
        const expandedObjectsListTitle = await within(expandedObjectsListContainer).findByText(/expanded objects/i);
        const expandedObjectsEmptyMessage = await within(expandedObjectsListContainer).findByText(/no results found/i);

        expect(directObjectsListTitle).toBeInTheDocument();
        expect(directObjectsList[0]).toBeInTheDocument();
        expect(expandedObjectsListTitle).toBeInTheDocument();
        expect(expandedObjectsEmptyMessage).toBeInTheDocument();
    });
    it('shows the direct object empty and expanded object list empty, when no results in both', async () => {
        setPreviewResultsTestData([]);
        render(<SeedSelectionPreview seeds={TestSeeds} ruleType={2} />);

        const directObjectsListContainer = await screen.findByTestId('pz-rule-preview__direct-objects-list');
        const directObjectsListTitle = await within(directObjectsListContainer).findByText(/direct objects/i);
        const directObjectsEmptyMessage = await within(directObjectsListContainer).findByText(/no results found/i);

        const expandedObjectsListContainer = await screen.findByTestId('pz-rule-preview__expanded-objects-list');
        const expandedObjectsListTitle = await within(expandedObjectsListContainer).findByText(/expanded objects/i);
        const expandedObjectsEmptyMessage = await within(expandedObjectsListContainer).findByText(/no results found/i);

        expect(directObjectsListTitle).toBeInTheDocument();
        expect(directObjectsEmptyMessage).toBeInTheDocument();
        expect(expandedObjectsListTitle).toBeInTheDocument();
        expect(expandedObjectsEmptyMessage).toBeInTheDocument();
    });
});
