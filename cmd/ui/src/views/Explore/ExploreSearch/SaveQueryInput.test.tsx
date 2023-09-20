import { render, screen } from 'src/test-utils';
import SaveQueryInput from './SaveQueryInput';
import userEvent from '@testing-library/user-event';

describe('SaveQueryInput', () => {
    beforeEach(() => {
        render(<SaveQueryInput />);
    });
    it('should open the text input when the save button is clicked', async () => {
        const user = userEvent.setup();

        const saveButton = screen.getByRole('button', { name: /floppy-disk/i });
        expect(saveButton).toBeInTheDocument();

        expect(screen.queryByRole('textbox', { name: /search name/i })).not.toBeInTheDocument();

        await user.click(saveButton);

        const textInput = screen.getByRole('textbox', { name: /search name/i });
        expect(textInput).toBeInTheDocument();
    });

    it('should disable the save button when no name has been provided to the text input', async () => {
        const user = userEvent.setup();

        const saveButton = screen.getByRole('button', { name: /floppy-disk/i });
        expect(saveButton).toBeEnabled();

        await user.click(saveButton);

        expect(saveButton).toBeDisabled();
    });

    // todo:
    it('should handle user input', async () => {});
});
