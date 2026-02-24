import {ConfigurationKey} from "js-client-library";
import {setupServer} from "msw/node";
import {rest} from "msw";
import {render, screen} from "../../test-utils";
import {QueryClient} from "react-query";
import {configurationKeys} from "../../hooks";
import UserActionsMenu from "./UserActionsMenu";


const CONFIG_ENABLED_RESPONSE = {
    data: [
        {
            key: ConfigurationKey.APITokens,
            value: {
                enabled: true,
            },
        },
    ],
}

const CONFIG_DISABLED_RESPONSE = {
    data: [
        {
            key: ConfigurationKey.APITokens,
            value: {
                enabled: false,
            },
        },
    ],
}

const noop = () => undefined;

interface testProps {
    userId: string;
    onOpen: (e: any, userId: string) => any;
    showPasswordOptions: boolean;
    showAuthMgmtButtons: boolean;
    showDisableMfaButton: boolean;
    userDisabled: boolean;
    onUpdateUser: (e: any) => any;
    onDisableUser: (e: any) => any;
    onEnableUser: (e: any) => any;
    onDeleteUser: (e: any) => any;
    onUpdateUserPassword: (e: any) => any;
    onExpireUserPassword: (e: any) => any;
    onManageUserTokens: (e: any) => any;
    onDisableUserMfa: (e: any) => any;
    index: number;
}


const server = setupServer(
    rest.get(`/api/v2/config`, async (_req, res, ctx) => {
        return res(
            ctx.json(CONFIG_ENABLED_RESPONSE)
        );
    }),
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Api Keys', () => {

    const setSelectedUserId = (id: string) => {
        return noop();
    }

    const user = {
        id: "testID",
        sso_provider_id: true,
    }


    const setup = async ({
        userId,
        onOpen,
        showPasswordOptions,
        showAuthMgmtButtons,
        showDisableMfaButton,
        userDisabled,
        onUpdateUser,
        onDisableUser,
        onEnableUser,
        onDeleteUser,
        onUpdateUserPassword,
        onExpireUserPassword,
        onManageUserTokens,
        onDisableUserMfa,
        index,
    }: testProps): Promise<void> => {
        ( () => {
            render(
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
        });
    };


    /*
    const setup: React.FC<testProps> = ({
        userId,
        onOpen,
        showPasswordOptions,
        showAuthMgmtButtons,
        showDisableMfaButton,
        userDisabled,
        onUpdateUser,
        onDisableUser,
        onEnableUser,
        onDeleteUser,
        onUpdateUserPassword,
        onExpireUserPassword,
        onManageUserTokens,
        onDisableUserMfa,
        index,
    }) => {{
            await act(async () => {
                render(
                    <UserActionsMenu
                        userId={user.id}
                        onOpen={(_, userId) => {
                            setSelectedUserId(userId);
                        }}
                        showPasswordOptions={!user.sso_provider_id}
                        showAuthMgmtButtons={user.id !== "notTestId"}
                        showDisableMfaButton={false}
                        userDisabled={false}
                        onUpdateUser={() => noop()}
                        onDisableUser={() => noop()}
                        onEnableUser={() => noop()}
                        onDeleteUser={() => noop()}
                        onUpdateUserPassword={() => noop()}
                        onExpireUserPassword={() => noop()}
                        onManageUserTokens={() => noop()}
                        onDisableUserMfa={() => noop()}
                        index={98}
                    />
                );
        });
    };
    */

    it('should display generate/revoke api tokens button', async () => {

        const queryClient = new QueryClient()

        render(
            <UserActionsMenu
                userId={user.id}
                onOpen={(_, userId) => {
                    setSelectedUserId(userId);
                }}
                showPasswordOptions={!user.sso_provider_id}
                showAuthMgmtButtons={user.id !== "notTestId"}
                showDisableMfaButton={false}
                userDisabled={false}
                onUpdateUser={() => noop()}
                onDisableUser={() => noop()}
                onEnableUser={() => noop()}
                onDeleteUser={() => noop()}
                onUpdateUserPassword={() => noop()}
                onExpireUserPassword={() => noop()}
                onManageUserTokens={() => noop()}
                onDisableUserMfa={() => noop()}
                index={98}
            />, {queryClient}
        )
        const apiKeyManagementButton = screen.findByRole('menuitem', { name: /generate \/ revoke api tokens/i });
        expect(apiKeyManagementButton).toBeInTheDocument();

    },);

    it('should not display generate/revoke api tokens button', async () => {

        server.use(
            rest.get(`/api/v2/config`, async (_req, res, ctx) => {
                return res(
                    ctx.json(CONFIG_DISABLED_RESPONSE)
                );
            }),
        );

        const queryClient = new QueryClient()
        render(
            <UserActionsMenu
                userId={user.id}
                onOpen={(_, userId) => {
                    setSelectedUserId(userId);
                }}
                showPasswordOptions={!user.sso_provider_id}
                showAuthMgmtButtons={user.id !== "notTestId"}
                showDisableMfaButton={false}
                userDisabled={false}
                onUpdateUser={() => noop()}
                onDisableUser={() => noop()}
                onEnableUser={() => noop()}
                onDeleteUser={() => noop()}
                onUpdateUserPassword={() => noop()}
                onExpireUserPassword={() => noop()}
                onManageUserTokens={() => noop()}
                onDisableUserMfa={() => noop()}
                index={98}
            />, {queryClient}
        )

        await queryClient.invalidateQueries(configurationKeys.all)
        const apiKeyManagementButton = screen.queryByRole('menuitem', { name: /generate \/ revoke api tokens/i });
        expect(apiKeyManagementButton).not.toBeInTheDocument();
    });
});
