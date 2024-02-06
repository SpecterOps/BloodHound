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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import AssetGroupMemberList from './AssetGroupMemberList';
import { render, waitFor } from '../../test-utils';
import { createMockAssetGroup, createMockAssetGroupMembers } from '../../mocks/factories';
import userEvent from '@testing-library/user-event';

const assetGroup = createMockAssetGroup();
const assetGroupMembers = createMockAssetGroupMembers();

const server = setupServer(
    rest.get('/api/v2/asset-groups/1/members', (req, res, ctx) => {
        return res(
            ctx.json({
                count: assetGroupMembers.members.length,
                limit: 100,
                skip: 0,
                data: assetGroupMembers,
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('AssetGroupMemberList', () => {
    const setup = () => {
        const handleSelectMember = vi.fn();
        const user = userEvent.setup();
        const screen = render(
            <AssetGroupMemberList
                assetGroup={assetGroup}
                filter={{}}
                onSelectMember={handleSelectMember}
                canFilterToEmpty={false}
            />
        );
        return { screen, user, handleSelectMember };
    };

    it('Should display headers for member name and count', () => {
        const { screen } = setup();
        expect(screen.getByText('Name')).toBeInTheDocument();
        expect(screen.getByText('Custom Member')).toBeInTheDocument();
    });

    it('Should display a list of the asset group members', () => {
        const { screen } = setup();
        waitFor(() => {
            for (const member of assetGroupMembers.members) {
                expect(screen.getByText(member.name)).toBeInTheDocument();
            }
        });
    });

    it('Should call handler when a member is clicked', async () => {
        const { screen, user, handleSelectMember } = setup();
        const testMember = assetGroupMembers.members[0];
        const entry = await waitFor(() => screen.getByText(testMember.name));
        await user.click(entry);
        expect(handleSelectMember).toHaveBeenCalledWith(testMember);
    });
});
