// Copyright 2024 Specter Ops, Inc.
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
import { render, screen } from '../../test-utils';
import CreateMenu from './CreateMenu';

const mockUseFeatureFlag = vi.fn();

vi.mock('../../hooks/useFeatureFlags', () => {
    return {
        useFeatureFlag: (flagKey: string) => mockUseFeatureFlag(flagKey),
    };
});

describe('CreateMenu', () => {
    describe('oidc_support feature flag enabled', () => {
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
            const openSAMLProviderDialog = vi.fn();
            const openOIDCProviderDialog = vi.fn();
            render(
                <CreateMenu
                    createMenuTitle={`Create Provider`}
                    featureFlag='oidc_support'
                    featureFlagEnabledMenuItems={[
                        { title: 'SAML Provider', onClick: openSAMLProviderDialog },
                        { title: 'OIDC Provider', onClick: openOIDCProviderDialog },
                    ]}
                    menuItems={[{ title: 'SAML Provider', onClick: openSAMLProviderDialog }]}
                />
            );

            expect(screen.getByRole('button', { name: /create provider/i })).toBeInTheDocument();

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create provider/i }));

            expect(await screen.findByRole('menu')).toBeInTheDocument();
            expect(screen.getByRole('menuitem', { name: /saml provider/i })).toBeInTheDocument();
            expect(screen.getByRole('menuitem', { name: /oidc provider/i })).toBeInTheDocument();
        });

        it('calls openSAMLProviderDialog when SAML Provider menu item is clicked', async () => {
            const user = userEvent.setup();
            const openSAMLProviderDialog = vi.fn();
            const openOIDCProviderDialog = vi.fn();
            render(
                <CreateMenu
                    createMenuTitle={`Create Provider`}
                    featureFlag='oidc_support'
                    featureFlagEnabledMenuItems={[
                        { title: 'SAML Provider', onClick: openSAMLProviderDialog },
                        { title: 'OIDC Provider', onClick: openOIDCProviderDialog },
                    ]}
                    menuItems={[{ title: 'SAML Provider', onClick: openSAMLProviderDialog }]}
                />
            );

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create provider/i }));

            expect(await screen.findByRole('menu')).toBeInTheDocument();

            // click menu item
            await user.click(screen.getByRole('menuitem', { name: /saml provider/i }));

            expect(openSAMLProviderDialog).toHaveBeenCalled();

            // menu has been closed
            expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });

        it('calls openOIDCProviderDialog when OIDC Provider menu item is clicked', async () => {
            const user = userEvent.setup();
            const openSAMLProviderDialog = vi.fn();
            const openOIDCProviderDialog = vi.fn();
            render(
                <CreateMenu
                    createMenuTitle={`Create Provider`}
                    featureFlag='oidc_support'
                    featureFlagEnabledMenuItems={[
                        { title: 'SAML Provider', onClick: openSAMLProviderDialog },
                        { title: 'OIDC Provider', onClick: openOIDCProviderDialog },
                    ]}
                    menuItems={[{ title: 'SAML Provider', onClick: openSAMLProviderDialog }]}
                />
            );

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create provider/i }));

            expect(await screen.findByRole('menu')).toBeInTheDocument();

            // click menu item
            await user.click(screen.getByRole('menuitem', { name: /oidc provider/i }));

            expect(openOIDCProviderDialog).toHaveBeenCalled();

            // menu has been closed
            expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });
    });

    describe('oidc_support feature flag disabled', () => {
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
            const openSAMLProviderDialog = vi.fn();
            const openOIDCProviderDialog = vi.fn();
            render(
                <CreateMenu
                    createMenuTitle={`Create Provider`}
                    featureFlag='oidc_support'
                    featureFlagEnabledMenuItems={[
                        { title: 'SAML Provider', onClick: openSAMLProviderDialog },
                        { title: 'OIDC Provider', onClick: openOIDCProviderDialog },
                    ]}
                    menuItems={[{ title: 'Create SAML Provider', onClick: openSAMLProviderDialog }]}
                />
            );

            expect(screen.getByRole('button', { name: /create saml provider/i })).toBeInTheDocument();
        });

        it('calls openSAMLProviderDialog when create provider button is clicked', async () => {
            const user = userEvent.setup();
            const openOIDCProviderDialog = vi.fn();
            const openSAMLProviderDialog = vi.fn();
            render(
                <CreateMenu
                    createMenuTitle={`Create Provider`}
                    featureFlag='oidc_support'
                    featureFlagEnabledMenuItems={[
                        { title: 'SAML Provider', onClick: openSAMLProviderDialog },
                        { title: 'OIDC Provider', onClick: openOIDCProviderDialog },
                    ]}
                    menuItems={[{ title: 'Create SAML Provider', onClick: openSAMLProviderDialog }]}
                />
            );

            // click button to open menu
            await user.click(screen.getByRole('button', { name: /create saml provider/i }));

            expect(openSAMLProviderDialog).toHaveBeenCalled();
        });
    });
});
