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

import SAMLProviderTableActionsMenu from '.';

import { act, render, screen } from '../../test-utils';
import userEvent from '@testing-library/user-event';

describe('SAMLProviderTableActionsMenu', () => {
    it('should render', () => {
        const testSAMLProviderId = '1';
        const testOnDeleteSAMLProvider = vi.fn();

        render(
            <SAMLProviderTableActionsMenu
                SAMLProviderId={testSAMLProviderId}
                onDeleteSAMLProvider={testOnDeleteSAMLProvider}
            />
        );

        expect(screen.getByText('Delete SAML Provider')).not.toBeVisible();
        expect(screen.getByText('Delete SAML Provider')).toBeInTheDocument();
    });

    it('should display the menu options when the bars icon is clicked', async () => {
        const user = userEvent.setup();
        const testSAMLProviderId = '1';
        const testOnDeleteSAMLProvider = vi.fn();

        render(
            <SAMLProviderTableActionsMenu
                SAMLProviderId={testSAMLProviderId}
                onDeleteSAMLProvider={testOnDeleteSAMLProvider}
            />
        );

        expect(screen.getByText('Delete SAML Provider')).not.toBeVisible();

        await act(async () => {
            await user.click(screen.getByText('bars'));
        });

        expect(screen.getByText('Delete SAML Provider')).toBeVisible();
    });

    it('should call onDeleteSAMLProvider when Delete SAML Provider is clicked', async () => {
        const user = userEvent.setup();
        const testSAMLProviderId = '1';
        const testOnDeleteSAMLProvider = vi.fn();

        render(
            <SAMLProviderTableActionsMenu
                SAMLProviderId={testSAMLProviderId}
                onDeleteSAMLProvider={testOnDeleteSAMLProvider}
            />
        );

        await act(async () => {
            await user.click(screen.getByText('Delete SAML Provider'));
        });

        expect(testOnDeleteSAMLProvider).toHaveBeenCalled();
    });
});
