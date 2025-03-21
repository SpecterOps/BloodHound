import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test-utils';
import ConfirmationDialog from './ConfirmationDialog';

describe('ConfirmationDialog', () => {
    const user = userEvent.setup();
    const testOnClose = vi.fn();

    beforeEach(async () => {
        render(<ConfirmationDialog open={true} onClose={testOnClose} text='text-test' title='title-test' />);
        await waitFor(() => expect(screen.queryByRole('progressbar')).not.toBeInTheDocument());
    });

    it('should display correctly', () => {
        expect(screen.queryByText('text-test')).toBeInTheDocument();
        expect(screen.queryByText('title-test')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
    });

    it('should fire Cancel once with false', async () => {
        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalledWith(false);
        expect(testOnClose).toHaveBeenCalledTimes(1);
    });

    it('should fire Confirm once with true', async () => {
        await user.click(screen.getByRole('button', { name: /confirm/i }));

        expect(testOnClose).toHaveBeenCalledWith(true);
        expect(testOnClose).toHaveBeenCalledTimes(1);
    });
});
