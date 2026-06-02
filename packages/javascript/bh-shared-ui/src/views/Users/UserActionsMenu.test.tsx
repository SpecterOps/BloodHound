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
import userEvent from '@testing-library/user-event';
import { ConfigurationKey } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor } from '../../test-utils';
import { noop } from '../../utils';
import UserActionsMenu from './UserActionsMenu';

const CONFIG_ENABLED_RESPONSE = {
    data: [
        {
            key: ConfigurationKey.APITokens,
            value: {
                enabled: true,
            },
        },
    ],
};

const CONFIG_DISABLED_RESPONSE = {
    data: [
        {
            key: ConfigurationKey.APITokens,
            value: {
                enabled: false,
            },
        },
    ],
};

const createSelfResponse = (permissions: Array<{ authority: string; name: string }>) => ({
    data: {
        roles: [
            {
                permissions,
            },
        ],
    },
});

const MANAGE_USERS_RESPONSE = createSelfResponse([{ authority: 'auth', name: 'ManageUsers' }]);
const READ_USERS_RESPONSE = createSelfResponse([{ authority: 'auth', name: 'ReadUsers' }]);

type ComponentProps = React.ComponentProps<typeof UserActionsMenu>;

const server = setupServer(
    rest.get(`/api/v2/config`, async (_req, res, ctx) => {
        return res(ctx.json(CONFIG_ENABLED_RESPONSE));
    }),
    rest.get(`/api/v2/self`, async (_req, res, ctx) => {
        return res(ctx.json(MANAGE_USERS_RESPONSE));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('User Actions Menu', () => {
    const setup = (
        {
            userId = '',
            onOpen = noop,
            showPasswordOptions = false,
            showAuthMgmtButtons = false,
            showDisableMfaButton = false,
            userDisabled = false,
            onUpdateUser = noop,
            onDisableUser = noop,
            onEnableUser = noop,
            onDeleteUser = noop,
            onUpdateUserPassword = noop,
            onExpireUserPassword = noop,
            onManageUserTokens = noop,
            onDisableUserMfa = noop,
            index = 0,
        } = {} as ComponentProps
    ) => {
        const user = userEvent.setup();

        const screen = render(
            <UserActionsMenu
                userId={userId}
                onOpen={onOpen}
                showPasswordOptions={showPasswordOptions}
                showAuthMgmtButtons={showAuthMgmtButtons}
                showDisableMfaButton={showDisableMfaButton}
                userDisabled={userDisabled}
                onUpdateUser={onUpdateUser}
                onDisableUser={onDisableUser}
                onEnableUser={onEnableUser}
                onDeleteUser={onDeleteUser}
                onUpdateUserPassword={onUpdateUserPassword}
                onExpireUserPassword={onExpireUserPassword}
                onManageUserTokens={onManageUserTokens}
                onDisableUserMfa={onDisableUserMfa}
                index={index}
            />
        );

        return { screen, user };
    };

    describe('Api Keys', () => {
        it('should display generate/revoke api tokens button', async () => {
            const { user } = setup();
            const button = screen.getByRole('button', { name: /show user actions/i });

            await user.click(button);
            await screen.findByRole('menuitem', { name: /generate \/ revoke api tokens/i });
        });

        it('should not display generate/revoke api tokens button', async () => {
            server.use(
                rest.get(`/api/v2/config`, async (_req, res, ctx) => {
                    return res(ctx.json(CONFIG_DISABLED_RESPONSE));
                })
            );
            const { user } = setup();
            const button = screen.getByRole('button', { name: /show user actions/i });
            await user.click(button);

            await waitFor(() =>
                expect(
                    screen.queryByRole('menuitem', { name: /generate \/ revoke api tokens/i })
                ).not.toBeInTheDocument()
            );
        });
    });

    describe('Button Menu State', () => {
        it('enables the user actions menu for a user with the administrator role', async () => {
            server.use(
                rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                    return res(ctx.json(MANAGE_USERS_RESPONSE));
                })
            );

            setup();
            const button = screen.getByRole('button', { name: /show user actions/i });

            await waitFor(() => expect(button).not.toBeDisabled());
        });

        it('disables the user actions menu for a user with the auditor role', async () => {
            server.use(
                rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                    return res(ctx.json(READ_USERS_RESPONSE));
                })
            );

            setup();
            const button = screen.getByRole('button', { name: /show user actions/i });

            await waitFor(() => expect(button).toBeDisabled());
        });
    });
});
