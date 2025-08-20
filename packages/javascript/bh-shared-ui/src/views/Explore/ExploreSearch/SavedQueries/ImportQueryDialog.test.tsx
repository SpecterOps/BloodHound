import { fireEvent, render, screen, waitFor } from '../../../../test-utils';
import ImportQueryDialog from './ImportQueryDialog';

describe('ImportQueryDialog', () => {
    const testHandleClose = vi.fn();

    const errorFile = new File(['test text'], 'test.txt', { type: 'text/plain' });

    it('renders the Import Query Dialog', () => {
        render(<ImportQueryDialog open={true} onClose={testHandleClose} />);
        expect(screen.getByRole('dialog')).toBeInTheDocument();
        expect(screen.getByText('Upload Files')).toBeInTheDocument();
    });

    it('hanldes close event', () => {
        render(<ImportQueryDialog open={true} onClose={testHandleClose} />);
        const cancelButton = screen.getByText('Cancel');
        expect(cancelButton).toBeInTheDocument();
        fireEvent.click(cancelButton);
        expect(testHandleClose).toBeCalledTimes(1);
    });

    it('prevents a user from proceeding if the file is not valid', async () => {
        render(<ImportQueryDialog open={true} onClose={testHandleClose} />);

        const testUploadBtn = screen.getByRole('button', { name: 'Upload' });
        expect(testUploadBtn).toBeDisabled();

        const fileInput = screen.getByTestId('ingest-file-upload');

        await waitFor(() => expect(fileInput).toBeEnabled());
        expect(fileInput).toBeInTheDocument();

        await waitFor(() => fireEvent.change(fileInput, { target: { files: [errorFile] } }));
        expect(testUploadBtn).toBeDisabled();
    });
});
