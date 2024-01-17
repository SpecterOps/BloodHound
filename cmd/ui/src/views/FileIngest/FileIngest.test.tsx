// Copyright 2023 Specter Ops, Inc.
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

import { fireEvent, render, waitFor } from 'src/test-utils';
import FileIngest from '.';
import { setupServer } from 'msw/node';
import { rest } from 'msw';

const server = setupServer(
    rest.post('/api/v2/file-upload/start', (req, res, ctx) => {
        return res(
            ctx.json({
                data: { id: 1 },
                status: 201,
                statusText: 'Created',
            })
        );
    }),
    rest.post('/api/v2/file-upload/:ingestId', (req, res, ctx) => {
        return res(
            ctx.json({
                data: '',
                status: 202,
                statusText: 'Accepted',
            })
        );
    }),
    rest.post('/api/v2/file-upload/:ingestId/end', (req, res, ctx) => {
        return res(
            ctx.json({
                data: '',
                status: 200,
                statusText: 'OK',
            })
        );
    }),
    rest.get('/api/v2/file-upload', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        status: 2,
                        status_message: 'Complete',
                        id: 1,
                        start_time: '2023-08-01T22:03:20.245299Z',
                        end_time: '2023-08-01T22:04:23.097927Z',
                        user_email_address: 'test_email@specterops.io',
                    },
                ],

                status: 200,
                statusText: 'OK',
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('FileIngest', () => {
    const testFile = new File([JSON.stringify({ value: 'test' })], 'test.json', { type: 'application/json' });
    const errorFile = new File(['test text'], 'test.txt', { type: 'text/plain' });

    it('accepts a valid file and allows the user to continue through the upload process', async () => {
        const { getByTestId, getByText } = render(<FileIngest />);
        const openButton = getByText('Upload File(s)');

        fireEvent.click(openButton);
        const fileInput = getByTestId('ingest-file-upload');

        await waitFor(() => fireEvent.change(fileInput, { target: { files: [testFile] } }));

        const submitButton = getByTestId('confirmation-dialog_button-yes');
        expect(submitButton).toBeEnabled();

        fireEvent.click(submitButton);
        expect(getByText('Press "Upload" to continue.')).toBeInTheDocument();

        fireEvent.click(submitButton);
        await waitFor(() => getByText('All files have successfully been uploaded for ingest.'));
        expect(getByText('All files have successfully been uploaded for ingest.')).toBeInTheDocument();
    });

    it('prevents a user from proceeding if the file is not valid', async () => {
        const { getByTestId, getByText } = render(<FileIngest />);
        const openButton = getByText('Upload File(s)');

        fireEvent.click(openButton);
        const fileInput = getByTestId('ingest-file-upload');

        await waitFor(() => fireEvent.change(fileInput, { target: { files: [errorFile] } }));

        const submitButton = getByTestId('confirmation-dialog_button-yes');
        expect(submitButton).toBeDisabled();
    });

    it('displays a table of completed ingest logs', async () => {
        const { getByText } = render(<FileIngest />);
        await waitFor(() => getByText('test_email@specterops.io'));

        expect(getByText('test_email@specterops.io')).toBeInTheDocument();
        expect(getByText('1 minute')).toBeInTheDocument();
    });
});
