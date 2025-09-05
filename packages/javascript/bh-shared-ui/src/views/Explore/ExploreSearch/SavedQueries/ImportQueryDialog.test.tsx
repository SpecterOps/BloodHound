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
