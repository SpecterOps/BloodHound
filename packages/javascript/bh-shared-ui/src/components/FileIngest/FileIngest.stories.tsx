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
import { Description, Primary, Title } from '@storybook/blocks';
import type { Meta, StoryObj } from '@storybook/react';
import { expect, userEvent, waitFor, within } from '@storybook/test';
import { useSnackbar } from 'notistack';
import { useEffect, useRef, useState } from 'react';
import { FileUploadDialogContext, useSelf } from '../../hooks';
import { useNotifications } from '../../providers';
import NotificationsProvider from '../../providers/NotificationProvider/NotificationsProvider';
import FileUploadDialog from '../FileUploadDialog';
import { NotificationSnackbar } from '../NotificationSnackbar';
import FileIngest from './FileIngest';
import { createFileIngestHandlers, type IngestScenario } from './FileIngest.stories.mocks';

// The application authenticates before entering this page. Keep that order so the
// mount-only permission notification sees the resolved /self response.
const AuthenticatedFileIngest = () => {
    const { isLoading } = useSelf();
    const [showFileIngestDialog, setShowFileIngestDialog] = useState(false);
    if (isLoading) return <div role='status'>Loading current user…</div>;
    return (
        <FileUploadDialogContext.Provider value={{ showFileIngestDialog, setShowFileIngestDialog }}>
            <FileIngest />
            <FileUploadDialog open={showFileIngestDialog} onClose={() => setShowFileIngestDialog(false)} />
        </FileUploadDialogContext.Provider>
    );
};
// AppNotifications tracks displayed keys at module scope. Keep that presentation
// bookkeeping per mount here so navigating between stories cannot hide notifications.
const StoryNotifications = () => {
    const { notifications, removeNotification } = useNotifications();
    const { enqueueSnackbar, closeSnackbar } = useSnackbar();
    const displayed = useRef(new Set<string>());
    useEffect(() => {
        notifications.forEach(({ key, message, options = {}, dismissed }) => {
            if (dismissed) {
                closeSnackbar(key);
            } else if (!displayed.current.has(key)) {
                const { title, ...snackbarOptions } = options;
                displayed.current.add(key);
                enqueueSnackbar(message, {
                    ...snackbarOptions,
                    key,
                    content: (id, snackMessage) => (
                        <NotificationSnackbar id={id} message={snackMessage} variant={options.variant} title={title} />
                    ),
                    onExited: () => {
                        displayed.current.delete(key);
                        removeNotification(key);
                    },
                });
            }
        });
    }, [notifications, enqueueSnackbar, closeSnackbar, removeNotification]);
    return null;
};

const FileIngestExample = () => (
    <NotificationsProvider>
        <div className='min-h-screen py-4'>
            <AuthenticatedFileIngest />
        </div>
        <StoryNotifications />
    </NotificationsProvider>
);
const FileIngestDocs = () => (
    <>
        <Title />
        <Description />
        <Primary />
    </>
);
const meta = {
    title: 'Components/FileIngest',
    component: FileIngest,
    tags: ['autodocs'],
    parameters: {
        layout: 'fullscreen',
        router: { initialEntries: ['/administration/file-ingest'] },
        // Render only the primary example inline in Docs: simultaneous connected
        // scenarios would replace each other's MSW handlers.
        docs: {
            page: FileIngestDocs,
            story: { inline: true, height: '800px' },
            description: {
                component: `A connected page using real API hooks, permissions, feature flags, notifications and upload dialogs.
MSW intercepts the HTTP boundary. Each story supplies a fresh handler factory via parameters.msw; no backend or hook replacements are needed.
Try the filters, pagination, ingest IDs, and Upload File(s) button. Fixtures include every job status and both successful and problematic file results.
Select the named scenarios in the sidebar to explore permissions, the disabled feature flag, loading/error states, file details, and uploads. Docs renders this one live example so connected scenarios do not overwrite each other’s mocks.
The authenticated wrapper loads /self before mounting the page, matching application routing.`,
            },
        },
        msw: { handlers: () => createFileIngestHandlers() },
    },
    render: FileIngestExample,
} satisfies Meta<typeof FileIngest>;
export default meta;
type Story = StoryObj<typeof meta>;
const scenario = (options: IngestScenario, description: string): Story['parameters'] => ({
    msw: { handlers: () => createFileIngestHandlers(options) },
    docs: { description: { story: description } },
});
const permissionCheck =
    (permission: 'manage' | 'read' | 'none', legacy = false): Story['play'] =>
    async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        const button = await canvas.findByRole('button', { name: 'Upload File(s)' });
        if (permission === 'manage') await expect(button).toBeEnabled();
        else await expect(button).toBeDisabled();
        if (permission === 'none') {
            await expect(canvas.queryByText('collector@example.com')).not.toBeInTheDocument();
            const body = within(canvasElement.ownerDocument.body);
            await expect(await body.findByText(/does not grant permission to upload data/)).toBeVisible();
        } else {
            await expect((await canvas.findAllByText('collector@example.com'))[0]).toBeVisible();
        }
        if (legacy) {
            await expect(canvas.getByRole('columnheader', { name: /^End Time$/ })).toBeVisible();
            await expect(canvas.queryByRole('button', { name: 'Open file ingest filters' })).not.toBeInTheDocument();
            await expect(canvas.queryByRole('button', { name: 'View ingest 124 details' })).not.toBeInTheDocument();
        } else {
            await expect(canvas.getByRole('button', { name: 'Open file ingest filters' })).toBeVisible();
        }
    };
export const ManageIngest: Story = { play: permissionCheck('manage') };
export const ReadOnly: Story = {
    parameters: scenario({ permission: 'read' }, 'Read permission allows history and details, but disables uploads.'),
    play: permissionCheck('read'),
};
export const NoPermission: Story = {
    parameters: scenario(
        { permission: 'none' },
        'Neither ingest permission: history is not requested, uploads are disabled, and real permission notifications appear.'
    ),
    play: permissionCheck('none'),
};
export const FeatureFlagDisabled: Story = {
    parameters: scenario(
        { feature: 'disabled' },
        'open_graph_phase_2 is disabled: the legacy table shows an End Time column and has no filters, clickable ingest IDs, or file-details panel. Pagination and uploads remain available.'
    ),
    play: permissionCheck('manage', true),
};
export const EmptyHistory: Story = {
    parameters: scenario({ history: 'empty' }, 'No jobs yet. Upload a local JSON or ZIP file to create a mocked job.'),
};
export const FeatureFlagLoading: Story = {
    parameters: scenario({ feature: 'loading' }, 'A pending feature request displays the page loading overlay.'),
};
export const Error: Story = {
    parameters: scenario(
        { feature: 'error' },
        'The feature-flag error fallback appears when the feature request fails or the response is missing the required flag. This example mocks a failed request.'
    ),
    play: async ({ canvasElement }) => {
        await expect(await within(canvasElement).findByText('Error')).toBeVisible();
    },
};
export const HistoryLoading: Story = {
    parameters: scenario({ history: 'loading' }, 'A pending job-list request displays the modern table loading state.'),
};
export const HistoryError: Story = {
    parameters: scenario(
        { history: 'error' },
        'A failed job-list request displays the actual fetch-error notification.'
    ),
    play: async ({ canvasElement }) => {
        await expect(
            await within(canvasElement.ownerDocument.body).findByText(
                'Unable to fetch file upload jobs. Please try again.'
            )
        ).toBeVisible();
    },
};
export const FileDetails: Story = {
    parameters: scenario({}, 'Select ingest IDs and file names to inspect warning and error details.'),
    play: async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        await userEvent.click(await canvas.findByRole('button', { name: 'View ingest 124 details' }));
        await userEvent.click(await canvas.findByText('groups.json'));
        await expect(await canvas.findByText('Some objects could not be resolved.')).toBeVisible();
        await userEvent.click(canvas.getByRole('button', { name: 'View ingest 122 details' }));
        await userEvent.click(await canvas.findByText('computers.json'));
        await expect(await canvas.findByText('Invalid collector schema: missing metadata.')).toBeVisible();
    },
};
const chooseFile = async (canvasElement: HTMLElement) => {
    const canvas = within(canvasElement);
    const body = within(canvasElement.ownerDocument.body);
    await userEvent.click(await canvas.findByRole('button', { name: 'Upload File(s)' }));
    const input = await body.findByTestId('ingest-file-upload');
    await waitFor(() => expect(input).toBeEnabled());
    const file = new File([JSON.stringify({ data: [], meta: { type: 'users', count: 0, version: 5 } })], 'users.json', {
        type: 'application/json',
    });
    await userEvent.upload(input, file, { applyAccept: false });
    return body;
};
export const UploadSuccess: Story = {
    parameters: scenario(
        {},
        'Uploads run through start → upload → end. The in-memory job appears in history on its next poll.'
    ),
    play: async ({ canvasElement }) => {
        const body = await chooseFile(canvasElement);
        await userEvent.click(body.getByTestId('confirmation-dialog_button-yes'));
        await expect(await body.findByText('All files have successfully been uploaded for ingest.')).toBeVisible();
    },
};
export const UploadError: Story = {
    parameters: scenario(
        { upload: 'file-error' },
        'The first upload of each file fails with an API error. Use the file retry button to recover successfully.'
    ),
    play: async ({ canvasElement }) => {
        const body = await chooseFile(canvasElement);
        await userEvent.click(body.getByTestId('confirmation-dialog_button-yes'));
        await expect(
            await body.findByText('Some files have failed to upload and have not been included for ingest.')
        ).toBeVisible();
    },
};
