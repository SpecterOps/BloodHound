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

import { setupServer } from 'msw/node';
import { act, render, waitFor } from '../../test-utils';
import GroupManagement from './GroupManagement';
import { rest } from 'msw';
import { createMockDomain } from 'src/mocks/factories';
import { createMockAssetGroup, createMockAssetGroupMembers } from 'bh-shared-ui';
import userEvent from '@testing-library/user-event';

const domain = createMockDomain();
const assetGroup = createMockAssetGroup();
const assetGroupMembers = createMockAssetGroupMembers();

const server = setupServer(
    rest.get('/api/v2/available-domains', (req, res, ctx) => {
        return res(ctx.json({ data: [domain] }));
    }),
    rest.get('/api/v2/asset-groups', (req, res, ctx) => {
        return res(ctx.json({ data: { asset_groups: [assetGroup] } }));
    }),
    rest.get('/api/v2/asset-groups/1/members', (req, res, ctx) => {
        return res(
            ctx.json({
                count: assetGroupMembers.members.length,
                limit: 100,
                skip: 0,
                data: assetGroupMembers,
            })
        );
    }),
    rest.get('*', (req, res, ctx) => res(ctx.json({ data: [] })))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('GroupManagement', () => {
    const setup = async () =>
        await act(async () => {
            const user = userEvent.setup();
            const screen = render(<GroupManagement />, {
                initialState: {
                    global: {
                        options: { domain: {} },
                    },
                },
            });
            return { user, screen };
        });

    it('renders group and tenant dropdown selectors', async () => {
        const { screen } = await setup();
        const groupSelector = screen.getByTestId('dropdown_context-selector');
        const tenantSelector = await waitFor(() => screen.getByTestId('data-quality_context-selector'));

        expect(screen.getByText('Group:')).toBeInTheDocument();
        expect(screen.getByText('Environment:')).toBeInTheDocument();
        expect(groupSelector).toBeInTheDocument();
        expect(tenantSelector).toBeInTheDocument();
    });

    it('renders an edit form for the selected asset group', async () => {
        const { screen } = await setup();
        const input = screen.getByRole('combobox');
        expect(input).toBeInTheDocument();
    });

    it('renders a list of asset group members', async () => {
        const { screen } = await setup();
        const member = assetGroupMembers.members[0];

        expect(screen.getByRole('table')).toBeInTheDocument();
        expect(screen.getByText(member.name)).toBeInTheDocument();
    });

    it('renders an empty message for the entity panel before a node is selected', async () => {
        const { screen } = await setup();

        expect(screen.getByText('None Selected')).toBeInTheDocument();
        expect(screen.getByText('No information to display.')).toBeInTheDocument();
    });

    it('renders the node in the entity panel when member is clicked', async () => {
        const { screen, user } = await setup();
        const member = assetGroupMembers.members[0];
        const listItem = screen.getByText(member.name);
        const entityPanel = screen.getByTestId('explore_entity-information-panel');

        await user.click(listItem);
        const header = await waitFor(() => screen.getByText('Object Information'));

        expect(header).toBeInTheDocument();
        expect(entityPanel).toHaveTextContent(member.name);
    });

    it('renders a link to the explore page when member is clicked', async () => {
        const { screen, user } = await setup();
        const member = assetGroupMembers.members[0];
        const listItem = screen.getByText(member.name);

        await user.click(listItem);
        const link = screen.getByTestId('group-management_explore-link');
        expect(link).toBeInTheDocument();
    });
});
