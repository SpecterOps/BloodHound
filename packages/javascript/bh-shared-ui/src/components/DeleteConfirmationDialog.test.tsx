import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test-utils';
import DeleteConfirmationDialog from './DeleteConfirmationDialog';

describe('DeleteConfirmationDialog', () => {
    const user = userEvent.setup();
    const testOnClose = vi.fn();

    beforeEach(async () => {
        render(
            <DeleteConfirmationDialog open={true} onClose={testOnClose} itemName='test-item' itemType='test-type' />
        );
        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
    });

    it('should display correctly', () => {
        expect(screen.queryByText(/delete test-item\?/i)).toBeInTheDocument();
        expect(
            screen.queryByText(
                /continuing onwards will delete test-item and all associated configurations and findings\./i
            )
        ).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /confirm/i })).toBeDisabled();
        expect(screen.getByTestId('confirmation-dialog_challenge-text')).toBeInTheDocument();
    });

    it('should fire Cancel once with false', async () => {
        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalledWith(false);
        expect(testOnClose).toHaveBeenCalledTimes(1);
    });

    it('should fire Confirm once with true after typing challenge text', async () => {
        await user.type(screen.getByTestId('confirmation-dialog_challenge-text'), 'delete this test-type');
        expect(screen.getByRole('button', { name: /confirm/i })).not.toBeDisabled();
        await user.click(screen.getByRole('button', { name: /confirm/i }));

        expect(testOnClose).toHaveBeenCalledWith(true);
        expect(testOnClose).toHaveBeenCalledTimes(1);
    });
});
