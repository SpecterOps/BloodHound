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

import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import UserTokenManagementDialog from './UserTokenManagementDialog';
import { render, screen, waitFor, waitForElementToBeRemoved } from '../../test-utils';

const testUserId = '1';
const testTokens = [
    {
        name: 'test token 1',
        hmac_method: 'hmac-sha2-256',
        last_access: '2021-08-20T16:00:10.781058Z',
        id: '3cefd8d2-fd82-4820-ad42-dcf51f6cedcd',
        created_at: '2021-08-20T16:00:10.781412Z',
        updated_at: '2021-08-20T16:00:10.781412Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
    {
        name: 'test token 2',
        hmac_method: 'hmac-sha2-256',
        last_access: '2021-08-20T16:00:15.626439Z',
        id: '05263cbb-52e5-4268-988c-ba29fb61eada',
        created_at: '2021-08-20T16:00:15.626756Z',
        updated_at: '2021-08-20T16:00:15.626756Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
];

const server = setupServer(
    rest.get(`/api/v2/tokens`, (req, res, ctx) => {
        req.params['user_id'] = `eq:${testUserId}`;
        return res(
            ctx.json({
                data: {
                    tokens: testTokens,
                },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('UserTokenManagementDialog', () => {
    const user = userEvent.setup();
    const testOnClose = vi.fn();

    beforeEach(async () => {
        render(<UserTokenManagementDialog open={true} onClose={testOnClose} userId={testUserId} />);
        await waitForElementToBeRemoved(screen.queryByRole('progressbar'));
    });

    it('should display "Generate/Revoke API Tokens" title', () => {
        expect(screen.getByText('Generate/Revoke API Tokens')).toBeInTheDocument();
    });

    it('should eventually display list of tokens', async () => {
        await waitFor(() => {
            // number of tokens plus one additional row for the table header
            expect(screen.getAllByRole('row')).toHaveLength(testTokens.length + 1);
            expect(screen.getByText(testTokens[0].name)).toBeInTheDocument();
            expect(screen.getByText(testTokens[1].name)).toBeInTheDocument();
        });
    });

    it('should display "Close" button', () => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument();
    });

    describe('when user clicks "Close" button', () => {
        beforeEach(async () => {
            await user.click(screen.getByRole('button', { name: 'Close' }));
        });

        it('should fire "onClose" once', () => {
            expect(testOnClose).toHaveBeenCalledTimes(1);
        });
    });

    it('should display "Create Token" button', () => {
        expect(screen.getByRole('button', { name: 'Create Token' })).toBeInTheDocument();
    });

    describe('when user clicks "Create Token" button', () => {
        beforeEach(async () => {
            await user.click(screen.getByRole('button', { name: 'Create Token' }));
        });

        it('should display CreateUserTokenDialog', () => {
            expect(
                screen.getByRole('dialog', {
                    name: 'Create User Token',
                })
            ).toBeInTheDocument();
        });
    });
});
