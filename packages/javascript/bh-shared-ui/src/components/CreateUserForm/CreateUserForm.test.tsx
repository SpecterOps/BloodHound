import userEvent from '@testing-library/user-event';
// import { QueryClient } from 'react-query';
import { render, screen } from '../../test-utils';
import { QueryTestWrapper } from '../QueryTestWrapper';
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
    it('should not allow leading empty spaces in the form', async () => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];

        render(
            <QueryTestWrapper stateMap={mockState}>
                <CreateUserForm {...DEFAULT_PROPS} />
            </QueryTestWrapper>
        );

        const userInput = userEvent.type;
        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });
        await userInput(screen.getByLabelText(/email/i), ' ');
        await userInput(screen.getByLabelText(/principal/i), ' ');
        await userInput(screen.getByLabelText(/first/i), ' ');
        await userInput(screen.getByLabelText(/last/i), ' ');
        await userInput(screen.getByLabelText(/Initial password/i), ' ');
        await user.click(button);

        expect(await screen.findByText('Email Address is required'));
        expect(await screen.findByText('Principal Name must be 2 letters or more'));
        expect(await screen.findByText('First Name must be 2 letters or more'));
        expect(await screen.findByText('Last Name must be 2 letters or more'));
        expect(await screen.findByText('Password must be at least 12 characters long'));
    });

    it('should not allow the input to exceed the allowed length', async () => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];

        render(
            <QueryTestWrapper stateMap={mockState}>
                <CreateUserForm {...DEFAULT_PROPS} />
            </QueryTestWrapper>
        );

        const userInput = userEvent.type;
        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });
        await userInput(screen.getByLabelText(/email/i), 'a'.repeat(320) + '@domain.com');
        await userInput(screen.getByLabelText(/principal/i), 'a'.repeat(1001));
        await userInput(screen.getByLabelText(/first/i), 'a'.repeat(1001));
        await userInput(screen.getByLabelText(/last/i), 'a'.repeat(1001));
        await userInput(screen.getByLabelText(/Initial password/i), 'a');
        await user.click(button);

        expect(await screen.findByText('Email address must be less than 319 characters'));
        expect(await screen.findByText('Principal Name must be less than 1000 characters'));
        expect(await screen.findByText('First Name must be less than 1000 characters'));
        expect(await screen.findByText('Last Name must be less than 1000 characters'));
        expect(await screen.findByText('Password must be at least 12 characters long'));
    });

    it('should not allow empty spaces or special characters', async () => {
        const mockState = [
            {
                key: ['getRoles'],
                data: MOCK_ROLES,
            },
            { key: ['listSSOProviders'], data: null },
        ];

        render(
            <QueryTestWrapper stateMap={mockState}>
                <CreateUserForm {...DEFAULT_PROPS} />
            </QueryTestWrapper>
        );

        const userInput = userEvent.type;
        const user = userEvent.setup();
        const button = screen.getByRole('button', { name: 'Save' });
        await userInput(screen.getByLabelText(/email/i), ' c$c@gmail.com');
        await userInput(screen.getByLabelText(/principal/i), ' $$');
        await userInput(screen.getByLabelText(/first/i), ' $$ad$!');
        await userInput(screen.getByLabelText(/last/i), '@asd# ');
        await userInput(screen.getByLabelText(/Initial password/i), 'a');
        await user.click(button);

        expect(await screen.findByText('Please follow the example@domain.com format'));
        expect(await screen.findByText('Principal Name does not allow spaces or special characters'));
        expect(await screen.findByText('First Name does not allow spaces or special characters'));
        expect(await screen.findByText('Last Name does not allow spaces or special characters'));
        expect(await screen.findByText('Password must be at least 12 characters long'));
    });
});
