// Copyright 2026 Specter Ops, Inc.
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
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { QueryClient } from 'react-query';
import { extensionsKeys } from '../../hooks';
import { withoutErrorLogging } from '../../mocks';
import { render, waitFor } from '../../test-utils';
import { SchemaUploadDialog } from './SchemaUploadDialog';

const testFile = new File([JSON.stringify({ value: 'test' })], 'test.json', { type: 'application/json' });

const addNotificationMock = vi.fn();
const checkPermissionMock = vi.fn().mockReturnValue(true);
const showFileIngestDialogMock = { value: false };

vi.mock('../../providers', async () => {
    const actual = await vi.importActual('../../providers');
    return {
        ...actual,
        useNotifications: () => {
            return { addNotification: addNotificationMock };
        },
    };
});

vi.mock('../../hooks/usePermissions', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        usePermissions: () => ({
            checkPermission: checkPermissionMock,
            isSuccess: true,
        }),
    };
});

vi.mock('../../hooks/useFileUploadDialogContext', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        useFileUploadDialogContext: () => ({
            showFileIngestDialog: showFileIngestDialogMock.value,
        }),
    };
});

const server = setupServer(
    rest.put('/api/v2/extensions', (req, res, ctx) => {
        return res(
            ctx.json({
                data: '',
                status: 201,
            })
        );
    })
);

const OriginalXMLHttpRequest = XMLHttpRequest;

beforeAll(() => {
    server.listen();

    class MockXMLHttpRequest extends OriginalXMLHttpRequest {
        private __upload = {
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
            onabort: vi.fn(),
            onerror: vi.fn(),
            onload: vi.fn(),
            onloadend: vi.fn(),
            onloadstart: vi.fn(),
            onprogress: vi.fn(),
            ontimeout: vi.fn(),
            dispatchEvent: vi.fn(),
        };
        get upload() {
            return this.__upload as any;
        }
    }
    vi.stubGlobal('XMLHttpRequest', MockXMLHttpRequest);
});
afterEach(() => {
    server.resetHandlers();
    addNotificationMock.mockClear();
    checkPermissionMock.mockReturnValue(true);
    showFileIngestDialogMock.value = false;
});
afterAll(() => {
    server.close();
    vi.stubGlobal('XMLHttpRequest', OriginalXMLHttpRequest);
});

describe('SchemaUploadDialog', () => {
    it('renders a button for opening the dialog', () => {
        const screen = render(<SchemaUploadDialog />);
        const button = screen.getByRole('button', { name: 'Upload File' });
        expect(button).toBeInTheDocument();
    });

    it('disables the upload file button when the user does not have the opengraph write permission', () => {
        checkPermissionMock.mockReturnValue(false);

        const screen = render(<SchemaUploadDialog />);

        expect(screen.getByRole('button', { name: 'Upload File' })).toBeDisabled();
    });

    it('does not open the dialog when the user does not have the opengraph write permission', async () => {
        checkPermissionMock.mockReturnValue(false);

        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        expect(screen.getByRole('button', { name: 'Upload File' })).toBeDisabled();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        expect(screen.queryByRole('dialog', { name: 'Upload Schema Files' })).not.toBeInTheDocument();
    });

    it('does not open the Schema Upload dialog via drag when the Quick Upload dialog is already open', async () => {
        showFileIngestDialogMock.value = true;

        const screen = render(<SchemaUploadDialog />);

        const dragEvent = new Event('dragenter', { bubbles: true }) as any;
        dragEvent.dataTransfer = {
            types: ['Files'],
            items: [{ kind: 'file', type: 'application/json' }],
        };
        document.dispatchEvent(dragEvent);

        expect(screen.queryByRole('dialog', { name: 'Upload Schema Files' })).not.toBeInTheDocument();
    });

    it('opens the dialog when the button is clicked', async () => {
        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        expect(screen.queryByRole('dialog', { name: 'Upload Schema Files' })).not.toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        expect(screen.queryByRole('dialog', { name: 'Upload Schema Files' })).toBeInTheDocument();
    });

    it('closes the dialog when the "Cancel" button is clicked', async () => {
        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        await user.click(screen.getByRole('button', { name: 'Cancel' }));
        expect(screen.queryByRole('dialog', { name: 'Upload Schema Files' })).not.toBeInTheDocument();
    });

    it('allows a user to upload a single file and displays its name in the dialog', async () => {
        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        const fileInput = screen.getByTestId('ingest-file-upload');
        await user.upload(fileInput, testFile);

        expect(screen.getByText('test.json')).toBeInTheDocument();
    });

    it('On successful upload, displays a completion of 100% and adds a "Complete" button that closes the dialog', async () => {
        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        const fileInput = screen.getByTestId('ingest-file-upload');
        await user.upload(fileInput, testFile);
        await user.click(screen.getByRole('button', { name: 'Upload' }));

        expect(await screen.findByText('100%')).toBeInTheDocument();
        expect(await screen.findByRole('button', { name: 'Complete' })).toBeInTheDocument();
    });

    it('displays a completion of 100% and adds a "Close" button that closes the dialog after a successful upload', async () => {
        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        const fileInput = screen.getByTestId('ingest-file-upload');
        await user.upload(fileInput, testFile);
        await user.click(screen.getByRole('button', { name: 'Upload' }));

        expect(await screen.findByText('100%')).toBeInTheDocument();
        expect(await screen.findByRole('button', { name: 'Close' })).toBeInTheDocument();
    });

    it('adds a "Close" button that closes the dialog after a failed upload', async () => {
        server.use(rest.put('/api/v2/extensions', (req, res, ctx) => res(ctx.status(400))));

        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        const fileInput = screen.getByTestId('ingest-file-upload');
        await user.upload(fileInput, testFile);
        await user.click(screen.getByRole('button', { name: 'Upload' }));

        expect(await screen.findByText('Failed to Upload')).toBeInTheDocument();
        expect(await screen.findByRole('button', { name: 'Close' })).toBeInTheDocument();
    });

    it('invalidates the extensions query after a successful upload', async () => {
        const queryClient = new QueryClient({
            defaultOptions: {
                queries: {
                    retry: false,
                },
            },
        });
        const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');
        const screen = render(<SchemaUploadDialog />, { queryClient });
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        await user.upload(screen.getByTestId('ingest-file-upload'), testFile);
        await user.click(screen.getByRole('button', { name: 'Upload' }));

        await screen.findByRole('button', { name: 'Complete' });

        await waitFor(() => {
            expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: extensionsKeys.all });
        });
    });

    it('On unsuccessful upload, notifies with an error and displays a retry button', async () => {
        server.use(rest.put('/api/v2/extensions', (req, res, ctx) => res(ctx.status(400))));

        const screen = render(<SchemaUploadDialog />);
        const user = userEvent.setup();

        await user.click(screen.getByRole('button', { name: 'Upload File' }));
        const fileInput = screen.getByTestId('ingest-file-upload');
        await user.upload(fileInput, testFile);

        await withoutErrorLogging(async () => {
            await user.click(screen.getByRole('button', { name: 'Upload' }));

            expect(await screen.findByText('Failed to Upload')).toBeInTheDocument();
            expect(await screen.findByRole('button', { name: 'Retry upload' })).toBeInTheDocument();

            expect(addNotificationMock).toBeCalled();
        });
    });
});
