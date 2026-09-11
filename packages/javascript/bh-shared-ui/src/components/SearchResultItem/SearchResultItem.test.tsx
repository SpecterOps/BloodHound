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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render } from '../../test-utils';
import SearchResultItem, { NodeSearchResult } from './SearchResultItem';

const baseItem: NodeSearchResult = {
    label: 'ADMIN@TESTLAB.LOCAL',
    objectId: 'S-1-5-21-1',
    kind: 'User',
    distinguishedName: 'CN=Admin,OU=Users,DC=testlab,DC=local',
};

const server = setupServer(
    rest.get(`/api/v2/custom-nodes`, async (_req, res, ctx) => {
        return res(ctx.json({ data: [] }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const setup = async (item: NodeSearchResult, keyword?: string) => {
    return await act(async () => {
        return render(<SearchResultItem item={item} index={0} keyword={keyword} getItemProps={() => ({})} />);
    });
};

describe('SearchResultItem', () => {
    it('renders the list item', async () => {
        const screen = await setup(baseItem);
        expect(screen.getByTestId('explore_search_result-list-item')).toBeInTheDocument();
    });

    it('displays the label', async () => {
        const screen = await setup(baseItem);
        expect(screen.getByText(baseItem.label)).toBeInTheDocument();
    });

    it('falls back to objectId when the label is empty', async () => {
        const screen = await setup({ ...baseItem, label: '' });
        expect(screen.getByText(baseItem.objectId)).toBeInTheDocument();
    });

    it('displays the distinguished name when present', async () => {
        const screen = await setup(baseItem);
        expect(screen.getByText(baseItem.distinguishedName as string)).toBeInTheDocument();
    });

    it('does not render a distinguishe name line when it is missing', async () => {
        const { distinguishedName: _omitted, ...withoutDN } = baseItem;
        const screen = await setup(withoutDN);
        expect(screen.queryByText(/CN=Admin/)).not.toBeInTheDocument();
        expect(screen.getByText(baseItem.label)).toBeInTheDocument();
    });

    it('does not render a distinguished name line when it is an empty string', async () => {
        const screen = await setup({ ...baseItem, distinguishedName: '' });
        expect(screen.queryByText(/CN=Admin/)).not.toBeInTheDocument();
    });

    it('highlights the matched keyword in the label', async () => {
        const screen = await setup(baseItem, 'ADMIN');
        expect(screen.getByText('ADMIN')).toHaveAttribute('style', 'font-weight: bold;');
    });

    it('does not highlight the keyword within the distinguished name', async () => {
        const screen = await setup(baseItem, 'Admin');
        const dn = screen.getByText(baseItem.distinguishedName as string);
        expect(dn).not.toHaveAttribute('style', 'font-weight: bold;');
    });
});
