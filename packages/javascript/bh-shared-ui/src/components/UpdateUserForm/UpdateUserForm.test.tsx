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
import UpdateUserForm from './UpdateUserForm';

const DEFAULT_PROPS = {
    onCancel: () => null,
    onSubmit: () => vi.fn,
    userId: '2d92f310-68fc-402a-915a-438a57f81342',
    hasSelectedSelf: false,
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

const MOCK_USER = [
    {
        data: {
            sso_provider_id: null,
            AuthSecret: {
                digest_method: 'argon2',
                expires_at: '2025-08-17T22:14:58.489177Z',
                totp_activated: false,
                id: 1,
                created_at: '2025-05-19T22:14:58.490455Z',
                updated_at: '2025-05-19T22:14:58.490455Z',
                deleted_at: {
                    Time: '0001-01-01T00:00:00Z',
                    Valid: false,
                },
            },
            roles: [
                {
                    name: 'Administrator',
                    description: 'Can manage users, clients, and application configuration',
                    permissions: [
                        {
                            authority: 'app',
                            name: 'ReadAppConfig',
                            id: 1,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'app',
                            name: 'WriteAppConfig',
                            id: 2,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'risks',
                            name: 'GenerateReport',
                            id: 3,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'risks',
                            name: 'ManageRisks',
                            id: 4,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'auth',
                            name: 'CreateToken',
                            id: 5,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'auth',
                            name: 'ManageAppConfig',
                            id: 6,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'auth',
                            name: 'ManageProviders',
                            id: 7,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'auth',
                            name: 'ManageSelf',
                            id: 8,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'auth',
                            name: 'ManageUsers',
                            id: 9,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'clients',
                            name: 'Manage',
                            id: 10,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'clients',
                            name: 'Tasking',
                            id: 11,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'collection',
                            name: 'ManageJobs',
                            id: 12,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'graphdb',
                            name: 'Read',
                            id: 13,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'graphdb',
                            name: 'Write',
                            id: 14,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'saved_queries',
                            name: 'Read',
                            id: 15,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'saved_queries',
                            name: 'Write',
                            id: 16,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'clients',
                            name: 'Read',
                            id: 17,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'db',
                            name: 'Wipe',
                            id: 18,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'graphdb',
                            name: 'Mutate',
                            id: 19,
                            created_at: '2025-05-19T22:14:58.188368Z',
                            updated_at: '2025-05-19T22:14:58.188368Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                        {
                            authority: 'graphdb',
                            name: 'Ingest',
                            id: 20,
                            created_at: '2025-05-19T22:14:58.311151Z',
                            updated_at: '2025-05-19T22:14:58.311151Z',
                            deleted_at: {
                                Time: '0001-01-01T00:00:00Z',
                                Valid: false,
                            },
                        },
                    ],
                    id: 1,
                    created_at: '2025-05-19T22:14:58.188368Z',
                    updated_at: '2025-05-19T22:14:58.188368Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
            first_name: 'BloodHound',
            last_name: 'Dev',
            email_address: 'spam@example.com',
            principal_name: 'admin',
            last_login: '2025-05-30T15:00:41.511369Z',
            is_disabled: false,
            eula_accepted: true,
            id: '2d92f310-68fc-402a-915a-438a57f81342',
            created_at: '2025-05-19T22:14:58.489615Z',
            updated_at: '2025-05-19T22:16:51.805983Z',
            deleted_at: {
                Time: '0001-01-01T00:00:00Z',
                Valid: false,
            },
        },
    },
];

describe('UpdateUserForm', () => {
    it('should not allow the input to exceed the allowed length', async () => {
        const mockState = [
            {
                key: ['getUser', DEFAULT_PROPS.userId],
                data: MOCK_USER,
            },
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];
        const queryClient = SetUpQueryClient(mockState);

        render(<UpdateUserForm {...DEFAULT_PROPS} />, { queryClient });

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });

        await user.type(screen.getByLabelText(/email/i), 'a'.repeat(309) + '@domain.com');
        await user.type(screen.getByLabelText(/principal/i), 'a'.repeat(1001));
        await user.type(screen.getByLabelText(/first/i), 'a'.repeat(1001));
        await user.type(screen.getByLabelText(/last/i), 'a'.repeat(1001));
        await user.click(button);

        expect(await screen.findByText(`Email address must be less than ${MAX_EMAIL_LENGTH} characters`))
            .toBeInTheDocument;
        expect(await screen.findByText(`Principal Name must be less than ${MAX_NAME_LENGTH} characters`))
            .toBeInTheDocument;
        expect(await screen.findByText(`First Name must be less than ${MAX_NAME_LENGTH} characters`)).toBeInTheDocument;
        expect(await screen.findByText(`Last Name must be less than ${MAX_NAME_LENGTH} characters`)).toBeInTheDocument;
    });

    it('should not have less characters than the minimum requirement', async () => {
        const mockState = [
            {
                key: ['getUser', DEFAULT_PROPS.userId],
                data: MOCK_USER,
            },
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];
        const queryClient = SetUpQueryClient(mockState);

        render(<UpdateUserForm {...DEFAULT_PROPS} />, { queryClient });

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });

        await user.type(screen.getByLabelText(/principal/i), 'a');
        await user.type(screen.getByLabelText(/first/i), 'a');
        await user.type(screen.getByLabelText(/last/i), 'a');
        await user.click(button);

        expect(await screen.findByText(`Principal Name must be ${MIN_NAME_LENGTH} characters or more`))
            .toBeInTheDocument;
        expect(await screen.findByText(`First Name must be ${MIN_NAME_LENGTH} characters or more`)).toBeInTheDocument;
        expect(await screen.findByText(`Last Name must be ${MIN_NAME_LENGTH} characters or more`)).toBeInTheDocument;
    });

    it('should not allow leading or trailing empty spaces', async () => {
        const mockState = [
            {
                key: ['getUser', DEFAULT_PROPS.userId],
                data: MOCK_USER,
            },
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];
        const queryClient = SetUpQueryClient(mockState);

        render(<UpdateUserForm {...DEFAULT_PROPS} />, { queryClient });

        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });

        await user.type(screen.getByLabelText(/principal/i), ' dd');
        await user.type(screen.getByLabelText(/first/i), ' bsg!');
        await user.type(screen.getByLabelText(/last/i), 'asdfw ');
        await user.click(button);

        expect(await screen.findByText('Principal Name does not allow leading or trailing spaces')).toBeInTheDocument;
        expect(await screen.findByText('First Name does not allow leading or trailing spaces')).toBeInTheDocument;
        expect(await screen.findByText('Last Name does not allow leading or trailing spaces')).toBeInTheDocument;
    });
});
