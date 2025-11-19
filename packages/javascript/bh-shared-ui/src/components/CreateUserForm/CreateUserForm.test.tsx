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
import { Dialog } from '@bloodhoundenterprise/doodleui';
import userEvent from '@testing-library/user-event';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { render, screen, waitFor } from '../../test-utils';
import { setUpQueryClient } from '../../utils';
import { Roles } from '../../utils/roles';
import CreateUserForm from './CreateUserForm';

const DEFAULT_PROPS = {
    onSubmit: vi.fn(),
    isLoading: false,
    error: false,
    showEnvironmentAccessControls: false,
};

const MOCK_ROLES = [
    {
        name: Roles.ADMINISTRATOR,
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
        name: Roles.USER,
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
        name: Roles.READ_ONLY,
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
        name: Roles.UPLOAD_ONLY,
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
        name: Roles.POWER_USER,
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
    type SetupOptions = {
        renderShowEnvironmentAccessControls?: boolean;
    };

    const createFormInitSetup = (options?: SetupOptions) => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];

        const queryClient = setUpQueryClient(mockState);

        render(
            <Dialog open={true}>
                <CreateUserForm
                    {...DEFAULT_PROPS}
                    showEnvironmentAccessControls={options?.renderShowEnvironmentAccessControls || false}
                />
            </Dialog>,
            { queryClient }
        );
    };

    it('should not have less characters than the minimum requirement', async () => {
        createFormInitSetup();
        const user = userEvent.setup();

        const button = await waitFor(() => screen.getByRole('button', { name: 'Save' }));

        await user.type(screen.getByLabelText(/principal/i), ' ');
        await user.type(screen.getByLabelText(/first/i), ' ');
        await user.type(screen.getByLabelText(/last/i), ' ');
        await user.type(screen.getByLabelText(/Initial password/i), ' ');
        await user.click(button);

        expect(
            await screen.findByText(`Principal Name must be ${MIN_NAME_LENGTH} characters or more`)
        ).toBeInTheDocument();
        expect(await screen.findByText(`First Name must be ${MIN_NAME_LENGTH} characters or more`)).toBeInTheDocument();
        expect(await screen.findByText(`Last Name must be ${MIN_NAME_LENGTH} characters or more`)).toBeInTheDocument();
        expect(await screen.findByText('Password must be at least 12 characters long')).toBeInTheDocument();
    });

    it('should not allow the input to exceed the allowed length', async () => {
        createFormInitSetup();

        const user = userEvent.setup();

        const button = await waitFor(() => screen.getByRole('button', { name: 'Save' }));

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

        expect(
            await screen.findByText(`Email address must be less than ${MAX_EMAIL_LENGTH} characters`)
        ).toBeInTheDocument();
        expect(
            await screen.findByText(`Principal Name must be less than ${MAX_NAME_LENGTH} characters`)
        ).toBeInTheDocument();
        expect(
            await screen.findByText(`First Name must be less than ${MAX_NAME_LENGTH} characters`)
        ).toBeInTheDocument();
        expect(
            await screen.findByText(`Last Name must be less than ${MAX_NAME_LENGTH} characters`)
        ).toBeInTheDocument();
        expect(await screen.findByText('Password must be less than 1000 characters')).toBeInTheDocument();
    });

    it('should not allow leading or trailing empty spaces', async () => {
        createFormInitSetup();

        const user = userEvent.setup();
        const button = await waitFor(() => screen.getByRole('button', { name: 'Save' }), {
            timeout: 30000,
        });
        await user.type(screen.getByLabelText(/principal/i), ' dd');
        await user.type(screen.getByLabelText(/first/i), ' bsg!');
        await user.type(screen.getByLabelText(/last/i), 'asdfw ');
        await user.click(button);

        expect(await screen.findByText('Principal Name does not allow leading or trailing spaces')).toBeInTheDocument();
        expect(await screen.findByText('First Name does not allow leading or trailing spaces')).toBeInTheDocument();
        expect(await screen.findByText('Last Name does not allow leading or trailing spaces')).toBeInTheDocument();
    });

    it('should display Environmental Targeted Access Control panel when showEnvironmentAccessControls prop is true and read-only role is selected', async () => {
        createFormInitSetup({ renderShowEnvironmentAccessControls: true });

        const user = userEvent.setup();

        const input = screen.getByRole('combobox', { name: /Role/i });

        await user.click(input);

        const option = screen.getByRole('option', { name: Roles.READ_ONLY });
        await user.click(option);

        expect(option).not.toBeInTheDocument();
        expect(await screen.findByText('Environmental Targeted Access Control')).toBeInTheDocument();
    });

    it('should display Environmental Targeted Access Control panel when showEnvironmentAccessControls prop is true and user role is selected', async () => {
        createFormInitSetup({ renderShowEnvironmentAccessControls: true });

        const user = userEvent.setup();

        const input = screen.getByRole('combobox', { name: /Role/i });

        await user.click(input);

        const option = screen.getByRole('option', { name: Roles.USER });
        await user.click(option);

        expect(option).not.toBeInTheDocument();
        expect(await screen.findByText('Environmental Targeted Access Control')).toBeInTheDocument();
    });

    it('should hide Environmental Targeted Access Control panel when showEnvironmentAccessControls prop is true and power user role is selected', async () => {
        createFormInitSetup({ renderShowEnvironmentAccessControls: true });

        const user = userEvent.setup();

        const input = screen.getByRole('combobox', { name: /Role/i });

        await user.click(input);

        const option = screen.getByRole('option', { name: Roles.READ_ONLY });
        await user.click(option);

        const panelHeader = await screen.findByText(/Environmental Targeted Access Control/i);
        expect(panelHeader).toBeInTheDocument();

        await user.click(input);

        const optionPowerUser = screen.getByRole('option', { name: Roles.POWER_USER });
        await user.click(optionPowerUser);

        expect(option).not.toBeInTheDocument();
        expect(panelHeader).not.toBeInTheDocument();
    });

    it('should hide Environmental Targeted Access Control panel when showEnvironmentAccessControls prop is false', async () => {
        createFormInitSetup({ renderShowEnvironmentAccessControls: false });

        expect(screen.queryByText('Environmental Targeted Access Control')).not.toBeInTheDocument();
        expect(await screen.findByText('Create User')).toBeInTheDocument();
    });
});
