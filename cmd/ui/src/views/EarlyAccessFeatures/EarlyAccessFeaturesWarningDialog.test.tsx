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
import { EarlyAccessFeaturesWarningDialog } from './EarlyAccessFeatures';

describe('EarlyAccessFeaturesWarningDialog', () => {
    it('renders', () => {
        const testOnCancel = vi.fn();
        const testOnConfirm = vi.fn();
        render(<EarlyAccessFeaturesWarningDialog open={true} onCancel={testOnCancel} onConfirm={testOnConfirm} />);

        expect(screen.getByText('Heads up!')).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Take me back' })).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'I understand, show me the new stuff!' })).toBeInTheDocument();
    });

    it('calls onCancel when cancel button clicked', async () => {
        const user = userEvent.setup();
        const testOnCancel = vi.fn();
        const testOnConfirm = vi.fn();

        render(<EarlyAccessFeaturesWarningDialog open={true} onCancel={testOnCancel} onConfirm={testOnConfirm} />);

        // Click cancel button
        await user.click(screen.getByRole('button', { name: 'Take me back' }));

        expect(testOnCancel).toHaveBeenCalled();
    });

    it('calls onConfirm when confirm button clicked', async () => {
        const user = userEvent.setup();
        const testOnCancel = vi.fn();
        const testOnConfirm = vi.fn();

        render(<EarlyAccessFeaturesWarningDialog open={true} onCancel={testOnCancel} onConfirm={testOnConfirm} />);

        // Click confirm button
        await user.click(screen.getByRole('button', { name: 'I understand, show me the new stuff!' }));

        expect(testOnConfirm).toHaveBeenCalled();
    });
});
