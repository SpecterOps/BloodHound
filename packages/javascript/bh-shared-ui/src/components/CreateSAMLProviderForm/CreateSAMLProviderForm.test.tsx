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
import { render, screen, waitFor } from '../../test-utils';
import CreateSAMLProviderForm from './CreateSAMLProviderForm';

describe('CreateSAMLProviderForm', () => {
    it('should render inputs, labels, and action buttons', () => {
        const testOnClose = vi.fn();
        const testOnSubmit = vi.fn();
        render(<CreateSAMLProviderForm onClose={testOnClose} onSubmit={testOnSubmit} />);

        expect(screen.getByLabelText('SAML Provider Name')).toBeInTheDocument();

        expect(screen.getByLabelText('Choose File')).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();

        expect(screen.getByRole('button', { name: 'Submit' })).toBeInTheDocument();
    });

    it('should call onClose when the Close button is clicked', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSubmit = vi.fn();
        render(<CreateSAMLProviderForm onClose={testOnClose} onSubmit={testOnSubmit} />);

        await user.click(screen.getByRole('button', { name: 'Cancel' }));

        expect(testOnClose).toHaveBeenCalled();
    });

    it('should not call onSubmit when input is invalid and Submit button is clicked', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSubmit = vi.fn();
        render(<CreateSAMLProviderForm onClose={testOnClose} onSubmit={testOnSubmit} />);

        await user.click(screen.getByRole('button', { name: 'Submit' }));

        await waitFor(() => expect(screen.getByText('SAML Provider Name is required')).toBeInTheDocument());

        await waitFor(() => expect(screen.getByText('Metadata is required')).toBeInTheDocument());

        expect(testOnSubmit).not.toHaveBeenCalled();
    });

    it('should call onSubmit when input is valid and Submit button is clicked', async () => {
        const user = userEvent.setup();
        const testOnClose = vi.fn();
        const testOnSubmit = vi.fn();
        const validProviderName = 'test-provider-name';
        const validMetadata = new File([], 'test-metadata.xml');
        render(<CreateSAMLProviderForm onClose={testOnClose} onSubmit={testOnSubmit} />);

        await user.type(screen.getByLabelText('SAML Provider Name'), validProviderName);

        await user.upload(screen.getByLabelText('Choose File'), validMetadata);

        await user.click(screen.getByRole('button', { name: 'Submit' }));

        await waitFor(() => expect(testOnSubmit).toHaveBeenCalled());
    });
});
