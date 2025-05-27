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

import userEvent from '@testing-library/user-event';
import { createMemoryHistory } from 'history';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { Route, Routes } from 'react-router-dom';
import { tierHandlers } from '../../../mocks';
import { render, screen } from '../../../test-utils';
import { apiClient } from '../../../utils';
import { MembersList } from './MembersList';

const handlers = [...tierHandlers];

const server = setupServer(
    rest.get(`/api/v2/customnode`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    ...handlers
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const membersListSpy = vi.spyOn(apiClient, 'getAssetGroupSelectorMembers');

describe('MembersList', () => {
    it('sorting the list updates the list by changing the call made to the API', async () => {
        const user = userEvent.setup();

        const history = createMemoryHistory({
            initialEntries: ['/tier-management/details/tag/1/selector/1'],
        });

        render(
            <Routes>
                <Route path={'/'} element={<MembersList selected='1' onClick={vi.fn()} />} />;
                <Route
                    path={'/tier-management/details/tag/:tagId/selector/:selectorId'}
                    element={<MembersList selected='1' onClick={vi.fn()} itemCount={1} />}
                />
            </Routes>,
            { history }
        );

        expect(membersListSpy).toBeCalledWith('1', '1', 0, 128, 'name');

        await user.click(screen.getByText('Objects', { exact: false }));

        expect(membersListSpy).toBeCalledWith('1', '1', 0, 128, '-name');
    });
});
