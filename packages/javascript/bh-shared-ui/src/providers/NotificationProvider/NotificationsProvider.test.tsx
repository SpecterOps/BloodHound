// Copyright 2026 Specter Ops, Inc.
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
import { NotificationSnackbar } from '../../components/NotificationSnackbar';
import { render, screen } from '../../test-utils';

const closeSnackbarMock = vi.fn();

vi.mock('notistack', async () => {
    const actual = await vi.importActual<typeof import('notistack')>('notistack');
    return {
        ...actual,
        useSnackbar: () => ({ closeSnackbar: closeSnackbarMock, enqueueSnackbar: vi.fn() }),
    };
});

describe('NotificationSnackbar', () => {
    beforeEach(() => {
        closeSnackbarMock.mockClear();
    });

    it('renders the message', () => {
        render(<NotificationSnackbar id='snack-1' message='Something happened' variant='info' />);

        expect(screen.getByRole('alert')).toHaveTextContent('Something happened');
    });

    it('renders the title as a heading when a title is provided', () => {
        render(<NotificationSnackbar id='snack-1' message='Details of the event' variant='error' title='With Title' />);

        expect(screen.getByRole('heading', { name: 'With Title' })).toBeInTheDocument();
        expect(screen.getByText('Details of the event')).toBeInTheDocument();
    });

    it('does not render a heading when no title is provided', () => {
        render(<NotificationSnackbar id='snack-1' message='No title here' variant='error' />);

        expect(screen.queryByRole('heading')).not.toBeInTheDocument();
        expect(screen.getByText('No title here')).toBeInTheDocument();
    });

    it('forwards the variant to the underlying Alert styling', () => {
        render(<NotificationSnackbar id='snack-1' message='Error message' variant='error' />);

        expect(screen.getByRole('alert')).toHaveClass('bg-status-error-fill');
    });

    it('renders gracefully when variant is nullish', () => {
        render(<NotificationSnackbar id='snack-1' message='Default message' variant={null} />);

        expect(screen.getByRole('alert')).toHaveTextContent('Default message');
    });

    it('renders a dismiss button that closes the snackbar by id', async () => {
        const user = userEvent.setup();
        render(<NotificationSnackbar id='snack-42' message='Dismiss me' variant='success' />);

        const dismissButton = screen.getByRole('button', { name: 'Dismiss alert' });
        await user.click(dismissButton);

        expect(closeSnackbarMock).toHaveBeenCalledWith('snack-42');
    });
});
