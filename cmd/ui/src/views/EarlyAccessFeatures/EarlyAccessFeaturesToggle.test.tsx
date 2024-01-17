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
import { render, screen } from 'src/test-utils';
import { EarlyAccessFeatureToggle } from './EarlyAccessFeatures';
import { Flag } from 'src/hooks/useFeatureFlags';

describe('EarlyAccessFeatureToggle', () => {
    it('renders', () => {
        const testFlag: Flag = {
            id: 1,
            name: 'test-flag-1',
            key: 'test-flag-1',
            description: 'description-1',
            enabled: false,
            user_updatable: true,
        };

        const testOnClick = vi.fn();

        render(<EarlyAccessFeatureToggle flag={testFlag} onClick={testOnClick} disabled={false} />);

        expect(screen.getByText(testFlag.name)).toBeInTheDocument();

        expect(screen.getByText(testFlag.description)).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Disabled' })).toBeInTheDocument();
    });

    it('button displays "Enabled" when flag is enabled', () => {
        const testFlag: Flag = {
            id: 1,
            name: 'test-flag-1',
            key: 'test-flag-1',
            description: 'description-1',
            enabled: true,
            user_updatable: true,
        };

        const testOnClick = vi.fn();

        render(<EarlyAccessFeatureToggle flag={testFlag} onClick={testOnClick} disabled={false} />);

        expect(screen.getByText(testFlag.name)).toBeInTheDocument();

        expect(screen.getByText(testFlag.description)).toBeInTheDocument();

        expect(screen.getByRole('button', { name: /enabled/i })).toBeInTheDocument();
    });

    it('calls onClick when toggle button clicked', async () => {
        const user = userEvent.setup();
        const testFlag: Flag = {
            id: 1,
            name: 'test-flag-1',
            key: 'test-flag-1',
            description: 'description-1',
            enabled: false,
            user_updatable: true,
        };

        const testOnClick = vi.fn();

        render(<EarlyAccessFeatureToggle flag={testFlag} onClick={testOnClick} disabled={false} />);

        // Click toggle button
        await user.click(screen.getByRole('button', { name: /disabled/i }));

        expect(testOnClick).toHaveBeenCalledWith(testFlag.id);
    });
});
