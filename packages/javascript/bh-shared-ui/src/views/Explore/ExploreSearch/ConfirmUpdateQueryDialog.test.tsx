import userEvent from '@testing-library/user-event';
import { render } from '../../../test-utils';
import ConfirmUpdateQueryDialog from './ConfirmUpdateQueryDialog';

const testDialogContent = 'lorem ipsum';

describe('ConfirmUpdateQueryDialog', () => {
    const setup = async () => {
        const testHandleApply = vi.fn();
        const testHandleCancel = vi.fn();

        const screen = render(
            <ConfirmUpdateQueryDialog
                open={true}
                handleApply={testHandleApply}
                handleCancel={testHandleCancel}
                dialogContent={testDialogContent}
            />
        );
        const user = userEvent.setup();
        return { screen, user, testHandleApply, testHandleCancel };
    };

    it('should render with the correct content', async () => {
        const { screen } = await setup();
        expect(screen.getByText(testDialogContent)).toBeInTheDocument();
    });

    it('should handle click events on apply and cancel', async () => {
        const { screen, user, testHandleApply, testHandleCancel } = await setup();
        expect(screen.getByText(testDialogContent)).toBeInTheDocument();

        const cancel = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancel);
        expect(testHandleCancel).toBeCalled();

        const apply = screen.getByRole('button', { name: /ok/i });
        await user.click(apply);
        expect(testHandleApply).toBeCalled();
    });

    it('should not render if open === false', async () => {
        const screen = render(
            <ConfirmUpdateQueryDialog
                open={false}
                handleApply={vi.fn()}
                handleCancel={vi.fn()}
                dialogContent={testDialogContent}
            />
        );
        const testContent = screen.queryByText(testDialogContent);
        expect(testContent).toBeNull();
    });
});
