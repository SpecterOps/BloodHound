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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { bloodHoundUsersHandlers } from '../../../../mocks';
import { act, render, waitFor } from '../../../../test-utils';
import { HistoryTableContext } from '../HistoryTableContext';
import FilterDialog from './FilterDialog';

const server = setupServer(
    ...bloodHoundUsersHandlers,
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tags: [{ name: 'foo', id: 77 }] } }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Privilege Zones History Filter Dialog', () => {
    const setup = async (props?: Partial<React.ComponentProps<typeof FilterDialog>>) => {
        return await act(() => {
            const contextValue = {
                currentNote: '',
                setCurrentNote: () => {},
                showNoteDetails: '',
                setShowNoteDetails: () => {},
            };

            const defaultSetFilter = () => {};

            const screen = render(
                <HistoryTableContext.Provider value={contextValue}>
                    <FilterDialog {...props} setFilters={props?.setFilters || defaultSetFilter} />;
                </HistoryTableContext.Provider>
            );
            const user = userEvent.setup();
            const openDialog = async () => await user.click(screen.getByTestId('History_log_filter_dialog'));
            return { screen, user, openDialog };
        });
    };

    it('renders', async () => {
        const { screen, openDialog } = await setup();
        await openDialog();

        expect(screen.getByText('Filter')).toBeInTheDocument();

        expect(screen.getByLabelText('Action')).toBeInTheDocument();
        expect(screen.getByLabelText('Zone/Label')).toBeInTheDocument();
        expect(screen.getByLabelText('Made By')).toBeInTheDocument();
        expect(screen.getByLabelText('Start Date')).toBeInTheDocument();
        expect(screen.getByLabelText('End Date')).toBeInTheDocument();

        expect(screen.getByRole('button', { name: /Clear All/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Confirm/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
    });

    it('renders with applied filters', async () => {
        const { screen, openDialog } = await setup({
            filters: {
                action: 'CertifyNodeManual',
                tagId: 'foo',
                madeBy: 'test_admin@specterops.io',
                'start-date': '2025-07-12',
                'end-date': '2025-08-12',
            },
        });
        await openDialog();

        expect(screen.getByText('Filter')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.getByText('User Certification', { ignore: 'option' })).toBeInTheDocument();
        });

        await waitFor(() => {
            expect(screen.getByText('foo', { ignore: 'option' })).toBeInTheDocument();
        });

        await waitFor(() => {
            expect(screen.getByText('test_admin@specterops.io', { ignore: 'option' })).toBeInTheDocument();
        });

        expect(screen.getByLabelText('Start Date')).toHaveValue('2025-07-12');
        expect(screen.getByLabelText('End Date')).toHaveValue('2025-08-12');
    });

    it('clears applied filters when clicking the Clear button', async () => {
        const user = userEvent.setup();
        const { screen, openDialog } = await setup({
            filters: {
                action: 'CertifyNodeManual',
                tagId: 'foo',
                madeBy: 'test_admin@specterops.io',
                'start-date': '2025-07-12',
                'end-date': '2025-08-12',
            },
        });
        await openDialog();

        expect(screen.getByText('Filter')).toBeInTheDocument();

        expect(screen.getByText('User Certification', { ignore: 'option' })).toBeInTheDocument();
        expect(screen.getByLabelText('Start Date')).toHaveValue('2025-07-12');
        expect(screen.getByLabelText('End Date')).toHaveValue('2025-08-12');

        await user.click(screen.getByRole('button', { name: /Clear/ }));

        expect(screen.queryByText('Certified', { ignore: 'option' })).not.toBeInTheDocument();
        expect(screen.getByLabelText('Start Date')).toHaveValue('');
        expect(screen.getByLabelText('End Date')).toHaveValue('');
    });

    it('should close the dialog when the Cancel button is clicked', async () => {
        const user = userEvent.setup();

        const { screen, openDialog } = await setup({});
        await openDialog();
        expect(screen.getByText('Filter')).toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: /Cancel/ }));

        expect(screen.queryByText('Filter')).not.toBeInTheDocument();
    });

    it('calls setFilters when the Confirm button is clicked', async () => {
        const user = userEvent.setup();
        const setFilters = vi.fn();
        const { screen, openDialog } = await setup({
            setFilters: setFilters,
        });
        await openDialog();

        await user.click(screen.getByRole('button', { name: /Confirm/ }));

        expect(setFilters).toHaveBeenCalled();
    });
});
