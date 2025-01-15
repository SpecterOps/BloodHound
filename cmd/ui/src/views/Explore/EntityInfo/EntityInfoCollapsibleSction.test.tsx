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

import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import userEvent from '@testing-library/user-event';
import { screen, render, waitFor } from 'src/test-utils';
import { EntityInfoPanelContextProvider } from './EntityInfoPanelContextProvider';

describe('EntityInfoCollapsibleSection', () => {
    it('renders an error message without throwing a TypeError', async () => {
        const user = userEvent.setup();
        const testLabel = 'Section';
        const testCount = 100;
        const testOnChange = vi.fn();
        const testIsLoading = false;
        const testIsError = true;
        const error = {};

        render(
            <EntityInfoPanelContextProvider>
                <EntityInfoCollapsibleSection
                    label={testLabel}
                    count={testCount}
                    onChange={testOnChange}
                    isLoading={testIsLoading}
                    isError={testIsError}
                    error={error}
                />
            </EntityInfoPanelContextProvider>
        );

        expect(screen.getByText(testLabel)).toBeInTheDocument();
        user.click(screen.getByText(testLabel));
        await waitFor(() =>
            expect(screen.getByText('An unknown error occurred during the request.')).toBeInTheDocument()
        );
    });
});
