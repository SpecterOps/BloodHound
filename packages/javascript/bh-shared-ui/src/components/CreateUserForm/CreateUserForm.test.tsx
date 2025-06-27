// Copyright 2025 Specter Ops, Inc.
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
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { SetUpQueryClient, render, screen } from '../../test-utils';
import CreateUserForm from './CreateUserForm';
const DEFAULT_PROPS = {
    onCancel: () => null,
    onSubmit: () => vi.fn,
    isLoading: false,
    error: false,
};

const MOCK_ROLES = [
    {
        name: 'Administrator',
        description: 'Can manage users, clients, and application configuration',
        permissions: [],
        id: 1,
        created_at: '2025-04-24T20:28:45.676055Z',
        updated_at: '2025-04-24T20:28:45.676055Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
    {
        name: 'User',
        description: 'Can read data, modify asset group memberships',
        permissions: [],
        id: 2,
        created_at: '2025-04-24T20:28:45.676055Z',
        updated_at: '2025-04-24T20:28:45.676055Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
    {
        name: 'Read-Only',
        description: 'Used for integrations',
        permissions: [],
        id: 3,
        created_at: '2025-04-24T20:28:45.676055Z',
        updated_at: '2025-04-24T20:28:45.676055Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
    {
        name: 'Upload-Only',
        description: 'Used for data collection clients, can post data but cannot read data',
        permissions: [],
        id: 4,
        created_at: '2025-04-24T20:28:45.676055Z',
        updated_at: '2025-04-24T20:28:45.676055Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
    {
        name: 'Power User',
        description: 'Can upload data, manage clients, and perform any action a User can',
        permissions: [],
        id: 5,
        created_at: '2025-04-24T20:28:45.676055Z',
        updated_at: '2025-04-24T20:28:45.676055Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    },
];

describe('CreateUserForm', () => {
    it('should not have less characters than the minimum requirement', async () => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];

        const queryClient = SetUpQueryClient(mockState);

        render(<CreateUserForm {...DEFAULT_PROPS} />, { queryClient });

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });
        await user.type(screen.getByLabelText(/principal/i), ' ');
        await user.type(screen.getByLabelText(/first/i), ' ');
        await user.type(screen.getByLabelText(/last/i), ' ');
        await user.type(screen.getByLabelText(/Initial password/i), ' ');
        await user.click(button);

        expect(await screen.findByText(`Principal Name must be ${MIN_NAME_LENGTH} characters or more`))
            .toBeInTheDocument();
        expect(await screen.findByText(`First Name must be ${MIN_NAME_LENGTH} characters or more`)).toBeInTheDocument();
        expect(await screen.findByText(`Last Name must be ${MIN_NAME_LENGTH} characters or more`)).toBeInTheDocument();
        expect(await screen.findByText('Password must be at least 12 characters long')).toBeInTheDocument();
    });

    it('should not allow the input to exceed the allowed length', async () => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];
        const queryClient = SetUpQueryClient(mockState);

        render(<CreateUserForm {...DEFAULT_PROPS} />, { queryClient });

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });

        await user.click(screen.getByLabelText(/email/i));
        await user.paste('a'.repeat(309) + '@domain.com');

        await user.click(screen.getByLabelText(/principal/i));
        await user.paste('a'.repeat(1001));

        await user.click(screen.getByLabelText(/first/i));
        await user.paste('a'.repeat(1001));

        await user.click(screen.getByLabelText(/last/i));
        await user.paste('a'.repeat(1001));

        await user.click(screen.getByLabelText(/Initial password/i));
        await user.paste('a'.repeat(1001));

        await user.click(button);

        expect(await screen.findByText(`Email address must be less than ${MAX_EMAIL_LENGTH} characters`))
            .toBeInTheDocument();
        expect(await screen.findByText(`Principal Name must be less than ${MAX_NAME_LENGTH} characters`))
            .toBeInTheDocument();
        expect(await screen.findByText(`First Name must be less than ${MAX_NAME_LENGTH} characters`)).toBeInTheDocument();
        expect(await screen.findByText(`Last Name must be less than ${MAX_NAME_LENGTH} characters`)).toBeInTheDocument();
        expect(await screen.findByText('Password must be less than 1000 characters')).toBeInTheDocument();
    });

    it('should not allow leading or trailing empty spaces', async () => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];
        const queryClient = SetUpQueryClient(mockState);

        render(<CreateUserForm {...DEFAULT_PROPS} />, { queryClient });

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });
        await user.type(screen.getByLabelText(/principal/i), ' dd');
        await user.type(screen.getByLabelText(/first/i), ' bsg!');
        await user.type(screen.getByLabelText(/last/i), 'asdfw ');
        await user.click(button);

        expect(await screen.findByText('Principal Name does not allow leading or trailing spaces')).toBeInTheDocument();
        expect(await screen.findByText('First Name does not allow leading or trailing spaces')).toBeInTheDocument();
        expect(await screen.findByText('Last Name does not allow leading or trailing spaces')).toBeInTheDocument();
    });
});
