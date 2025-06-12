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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import FileIngest from '.';
import { createAuthStateWithPermissions } from '../../mocks';
import { fireEvent, render, screen, waitFor } from '../../test-utils';
import { Permission } from '../../utils';

const server = setupServer(
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions([Permission.GRAPH_DB_WRITE]).user,
            })
        );
    }),
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
    }),
    rest.get('/api/v2/file-upload/accepted-types', (req, res, ctx) => {
        return res(
            ctx.json({
                data: ['application/json'],
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
        render(<FileIngest />);

        const openButton = screen.getByText('Upload File(s)');
        await waitFor(() => expect(openButton).toBeEnabled());

        fireEvent.click(openButton);

        const fileInput = screen.getByTestId('ingest-file-upload');
        await waitFor(() => expect(fileInput).toBeEnabled());

        await waitFor(() => fireEvent.change(fileInput, { target: { files: [testFile] } }));

        const submitButton = screen.getByTestId('confirmation-dialog_button-yes');
        await expect(submitButton).toBeEnabled();

        fireEvent.click(submitButton);
        expect(screen.getByText('Press "Upload" to continue.')).toBeInTheDocument();

        fireEvent.click(submitButton);
        await waitFor(() => screen.getByText('All files have successfully been uploaded for ingest.'));
        expect(screen.getByText('All files have successfully been uploaded for ingest.')).toBeInTheDocument();
    });

    it('prevents a user from proceeding if the file is not valid', async () => {
        render(<FileIngest />);

        const openButton = screen.getByText('Upload File(s)');
        await waitFor(() => expect(openButton).toBeEnabled());

        fireEvent.click(openButton);

        const fileInput = screen.getByTestId('ingest-file-upload');
        await waitFor(() => expect(fileInput).toBeEnabled());

        await waitFor(() => fireEvent.change(fileInput, { target: { files: [errorFile] } }));

        const submitButton = screen.getByTestId('confirmation-dialog_button-yes');
        expect(submitButton).toBeDisabled();
    });

    it('displays a table of completed ingest logs', async () => {
        render(<FileIngest />);
        await waitFor(() => screen.getByText('test_email@specterops.io'));

        expect(screen.getByText('test_email@specterops.io')).toBeInTheDocument();
        expect(screen.getByText('1 minute')).toBeInTheDocument();
    });

    it('disables the upload button and does not populate a table if the user lacks the permission', async () => {
        render(<FileIngest />);

        expect(screen.queryByText('test_email@specterops.io')).toBeNull();
        expect(screen.queryByText('1 minute')).toBeNull();

        expect(screen.getByTestId('file-ingest_button-upload-files')).toBeDisabled();
    });
});
