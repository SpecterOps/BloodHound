import { render, screen } from 'src/test-utils';
import SaveQueryInput from './SaveQueryInput';
import userEvent from '@testing-library/user-event';
import { apiClient } from 'bh-shared-ui';

describe('SaveQueryInput', () => {
    beforeEach(() => {
        render(<SaveQueryInput cypherQuery='test query' />);
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

    it('should handle user input', async () => {
        const apiSpy = jest.spyOn(apiClient, 'createUserQuery');
        // @ts-ignore mock the apiClient instead of msw.setupServer(), we dont use the apiClient's response within the component
        apiSpy.mockReturnValue({ dont: 'care' });

        const user = userEvent.setup();

        const saveButton = screen.getByRole('button', { name: /floppy-disk/i });
        expect(saveButton).toBeEnabled();

        // open text input
        await user.click(saveButton);

        const textInput = screen.getByRole('textbox', { name: /search name/i });
        await user.type(textInput, 'my favorite cypher query');

        await user.click(saveButton);

        expect(apiSpy).toHaveBeenCalledTimes(1);
        expect(apiSpy).toHaveBeenCalledWith({ name: 'my favorite cypher query', query: 'test query' });
    });
});
