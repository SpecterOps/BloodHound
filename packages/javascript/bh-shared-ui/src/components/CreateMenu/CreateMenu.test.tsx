import userEvent from '@testing-library/user-event';
import { render, screen } from '../../test-utils';
import CreateClientMenu from './CreateMenu';

const mockUseFeatureFlag = vi.fn();

vi.mock('src/hooks/useFeatureFlags', () => {
    return {
        useFeatureFlag: (flagKey: string) => mockUseFeatureFlag(flagKey),
    };
});

describe('CreateClientMenu', () => {
    describe('azure_support feature flag enabled', () => {
        beforeEach(() => {
            mockUseFeatureFlag.mockImplementation((flagKey: string) => {
                return {
                    data: {
                        key: flagKey,
                        enabled: true, // flag enabled
                    },
                    isLoading: false,
                    isError: false,
                    error: null,
                };
            });
        });

        it('renders a button and a menu', async () => {
            const user = userEvent.setup();
            const testCreateSharpHoundClient = vi.fn();
            const testCreateAzureHoundClient = vi.fn();
            render(
                <CreateClientMenu
                    onClickCreateSharpHoundClient={testCreateSharpHoundClient}
                    onClickCreateAzureHoundClient={testCreateAzureHoundClient}
                />
            );

            expect(screen.getByRole('button', { name: /create client/i })).toBeInTheDocument();

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create client/i }));

            expect(await screen.findByRole('menu')).toBeInTheDocument();
            expect(screen.getByRole('menuitem', { name: /create sharphound client/i })).toBeInTheDocument();
            expect(screen.getByRole('menuitem', { name: /create azurehound client/i })).toBeInTheDocument();
        });

        it('calls onClickCreateSharpHoundClient when Create SharpHound Client menu item is clicked', async () => {
            const user = userEvent.setup();
            const testCreateSharpHoundClient = vi.fn();
            const testCreateAzureHoundClient = vi.fn();
            render(
                <CreateClientMenu
                    onClickCreateSharpHoundClient={testCreateSharpHoundClient}
                    onClickCreateAzureHoundClient={testCreateAzureHoundClient}
                />
            );

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create client/i }));

            expect(await screen.findByRole('menu')).toBeInTheDocument();

            // click menu item
            await user.click(screen.getByRole('menuitem', { name: /create sharphound client/i }));

            expect(testCreateSharpHoundClient).toHaveBeenCalled();

            // menu has been closed
            expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });

        it('calls onClickCreateAzureHoundClient when Create AzureHound Client menu item is clicked', async () => {
            const user = userEvent.setup();
            const testCreateSharpHoundClient = vi.fn();
            const testCreateAzureHoundClient = vi.fn();
            render(
                <CreateClientMenu
                    onClickCreateSharpHoundClient={testCreateSharpHoundClient}
                    onClickCreateAzureHoundClient={testCreateAzureHoundClient}
                />
            );

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create client/i }));

            expect(await screen.findByRole('menu')).toBeInTheDocument();

            // click menu item
            await user.click(screen.getByRole('menuitem', { name: /create azurehound client/i }));

            expect(testCreateAzureHoundClient).toHaveBeenCalled();

            // menu has been closed
            expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });
    });

    describe('azure_support feature flag disabled', () => {
        beforeEach(() => {
            mockUseFeatureFlag.mockImplementation((flagKey: string) => {
                return {
                    data: {
                        key: flagKey,
                        enabled: false, // flag disabled
                    },
                    isLoading: false,
                    isError: false,
                    error: null,
                };
            });
        });

        it('renders a button', () => {
            const testCreateSharpHoundClient = vi.fn();
            const testCreateAzureHoundClient = vi.fn();
            render(
                <CreateClientMenu
                    onClickCreateSharpHoundClient={testCreateSharpHoundClient}
                    onClickCreateAzureHoundClient={testCreateAzureHoundClient}
                />
            );

            expect(screen.getByRole('button', { name: /create client/i })).toBeInTheDocument();
        });

        it('calls onClickCreateSharpHoundClient when Create Client button is clicked', async () => {
            const user = userEvent.setup();
            const testCreateSharpHoundClient = vi.fn();
            const testCreateAzureHoundClient = vi.fn();
            render(
                <CreateClientMenu
                    onClickCreateSharpHoundClient={testCreateSharpHoundClient}
                    onClickCreateAzureHoundClient={testCreateAzureHoundClient}
                />
            );

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create client/i }));

            expect(testCreateSharpHoundClient).toHaveBeenCalled();
        });
    });
});
