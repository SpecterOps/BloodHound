import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import * as hooks from '../../hooks';
import { render, screen, within } from '../../test-utils';
import { FileIngestFilterDialog } from './FileIngestFilterDialog';

const originalTZ = process.env.TZ;

const mockObjectHook = (initialState = {}) => {
    const applyState = vi.fn();
    const deleteKeys = vi.fn();
    const state = initialState;
    vi.spyOn(hooks, 'useObjectState').mockReturnValue({ applyState, deleteKeys, state } as any);

    return { applyState, deleteKeys };
};

const renderFilterDialog = async (open = true) => {
    const onConfirmMock = vi.fn();
    render(<FileIngestFilterDialog onConfirm={onConfirmMock} />);

    const user = userEvent.setup();
    const filterButton = await screen.findByTestId('file_ingest_log-open_filter_dialog');

    if (open) {
        await user.click(filterButton);
    }

    return { filterButton, onConfirmMock, user };
};

// Find the select trigger and click it
const clickSelect = async (user: ReturnType<typeof userEvent.setup>, name: string) => {
    const select = screen.getByRole('combobox', { name: `${name} Select` });

    await user.click(select);
};

// Click on a date input and type in a date
const inputDate = async (user: ReturnType<typeof userEvent.setup>, placeholder: string, value: string) => {
    const input = screen.getByPlaceholderText(placeholder);
    await user.click(input);

    if (value) {
        await user.type(input, value);
    }

    // Dates are not updated until input loses focus
    await user.tab();
};

const server = setupServer(
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'open_graph_phase_2',
                        enabled: true,
                    },
                ],
            })
        );
    }),
    rest.get('/api/v2/bloodhound-users-minimal', (_, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    users: [
                        {
                            id: '1',
                            first_name: 'Client 1 First',
                            last_name: 'Client 1 Last',
                            email_address: 'client1@example.com',
                        },
                        {
                            id: '2',
                            first_name: 'Client 2 First',
                            last_name: 'Client 2 Last',
                            email_address: 'client2@example.com',
                        },
                        {
                            id: '3',
                            first_name: 'Client 3 First',
                            last_name: 'Client 3 Last',
                            email_address: 'client3@example.com',
                        },
                    ],
                },

                status: 200,
                statusText: 'OK',
            })
        );
    })
);

beforeAll(() => {
    process.env.TZ = 'UTC';
    server.listen();
});

afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});

afterAll(() => {
    process.env.TZ = originalTZ;
    server.close();
});

describe('FileIngestFilterDialog', () => {
    it('renders a filter button', async () => {
        const { filterButton } = await renderFilterDialog(false);
        expect(filterButton).toBeInTheDocument();
    });

    it('opens and closes the filter', async () => {
        const { user } = await renderFilterDialog();

        // No assertions needed as getBy* throws if elements are not found
        ['Filter', 'Status', 'Date Range', 'Users'].forEach((text) => screen.getByText(text));

        await user.click(screen.getByRole('button', { name: 'Cancel' }));
        expect(screen.queryByRole('button', { name: 'Cancel' })).not.toBeInTheDocument();
    });

    describe('Status Select', () => {
        it('has status filters', async () => {
            const { user } = await renderFilterDialog();
            await clickSelect(user, 'Status');

            // Grab the listbox that just opened (menu items)
            const listbox = await screen.findByRole('listbox');

            // Get all the menu items
            const options = within(listbox).getAllByRole('option');
            expect(options.map((o) => o.textContent)).toEqual(['None', 'Running', 'Complete', 'Failed']);
        });

        it('filters by the selected status', async () => {
            const { applyState } = mockObjectHook({ status: '' });
            const { user } = await renderFilterDialog(true);
            await clickSelect(user, 'Status');

            // Select a status from the listbox
            const completeStatus = await screen.findByRole('option', { name: 'Complete' });
            await user.click(completeStatus);

            expect(applyState).toBeCalledWith({ status: 2 });
        });

        it('clears the applied filter', async () => {
            const { deleteKeys } = mockObjectHook({ status: '2' });
            const { user } = await renderFilterDialog(true);
            await clickSelect(user, 'Status');

            // Select a status from the listbox
            const completeStatus = await screen.findByRole('option', { name: 'None' });
            await user.click(completeStatus);

            expect(deleteKeys).toBeCalledWith('status');
        });
    });

    describe('Date Range Inputs', () => {
        it('has date range input filters', async () => {
            const { user } = await renderFilterDialog();
            await inputDate(user, 'Start Date', '');
            await inputDate(user, 'End Date', '');
        });

        it('filters by the selected date range', async () => {
            const { onConfirmMock, user } = await renderFilterDialog();
            await inputDate(user, 'Start Date', '2025-01-01');
            await user.click(screen.getByRole('button', { name: 'Confirm' }));

            expect(onConfirmMock).toBeCalledWith({ start_time: '2025-01-01T00:00:00.000Z' });
        });

        it('does not allow confirmation while dates are invalid', async () => {
            const { onConfirmMock, user } = await renderFilterDialog();
            await inputDate(user, 'Start Date', '2025-01-03');
            await inputDate(user, 'End Date', '2025-01-01');
            await user.click(screen.getByRole('button', { name: 'Confirm' }));

            expect(onConfirmMock).not.toBeCalledWith();
        });
    });

    describe('User Select', () => {
        it('has user filters', async () => {
            const { user } = await renderFilterDialog();
            await clickSelect(user, 'User');

            // Grab the listbox that just opened (menu items)
            const listbox = await screen.findByRole('listbox');
            // Get all the menu items
            const options = within(listbox).getAllByRole('option');

            expect(options.map((o) => o.textContent)).toEqual([
                'None',
                'client1@example.com',
                'client2@example.com',
                'client3@example.com',
            ]);
        });

        it('filters by the selected user', async () => {
            const { applyState } = mockObjectHook({ status: '' });
            const { user } = await renderFilterDialog(true);
            await clickSelect(user, 'User');

            // Select a status from the listbox
            const selecteUser = await screen.findByRole('option', { name: 'client2@example.com' });
            await user.click(selecteUser);

            expect(applyState).toBeCalledWith({ user_id: '2' });
        });

        it('clears the applied filter', async () => {
            const { deleteKeys } = mockObjectHook({ user_id: '2' });
            const { user } = await renderFilterDialog(true);
            await clickSelect(user, 'User');

            // Select a status from the listbox
            const completeStatus = await screen.findByRole('option', { name: 'None' });
            await user.click(completeStatus);

            expect(deleteKeys).toBeCalledWith('user_id');
        });
    });
});
