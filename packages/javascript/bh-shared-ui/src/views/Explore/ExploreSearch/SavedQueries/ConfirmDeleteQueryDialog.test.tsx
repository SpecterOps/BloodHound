import { vi } from 'vitest';

import userEvent from '@testing-library/user-event';
import { render } from '../../../../test-utils';
import ConfirmDeleteQueryDialog from './ConfirmDeleteQueryDialog';

describe('ConfirmDeleteQueryDialog', () => {
    const testDeleteHandler = vi.fn();
    const testClose = vi.fn();
    it('renders a confirm dialg', async () => {
        const screen = render(
            <ConfirmDeleteQueryDialog
                open={true}
                queryId={123}
                deleteHandler={testDeleteHandler}
                handleClose={testClose}
            />
        );
        expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
    });
    it('fires click handlers', async () => {
        const user = userEvent.setup();
        const screen = render(
            <ConfirmDeleteQueryDialog
                open={true}
                queryId={123}
                deleteHandler={testDeleteHandler}
                handleClose={testClose}
            />
        );
        const testCancelBtn = screen.getByRole('button', { name: /cancel/i });
        const testConfirmBtn = screen.getByRole('button', { name: /confirm/i });

        expect(testCancelBtn).toBeInTheDocument();
        expect(testConfirmBtn).toBeInTheDocument();

        await user.click(testCancelBtn);
        expect(testClose).toBeCalled();

        await user.click(testConfirmBtn);
        expect(testDeleteHandler).toBeCalled();
    });
});
