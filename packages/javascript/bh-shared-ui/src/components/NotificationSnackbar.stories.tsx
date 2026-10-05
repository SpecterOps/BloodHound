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
import type { Meta, StoryObj } from '@storybook/react';
import { expect, userEvent, waitFor, within } from '@storybook/test';
import { Button } from 'doodle-ui';
import { SnackbarProvider, useSnackbar } from 'notistack';
import { ComponentProps } from 'react';
import { NotificationSnackbar } from './NotificationSnackbar';

function NotificationTrigger(args: ComponentProps<typeof NotificationSnackbar>) {
    const { enqueueSnackbar } = useSnackbar();
    return (
        <Button
            onClick={() =>
                enqueueSnackbar(args.message, {
                    key: args.id,
                    persist: true,
                    content: (id, message) => <NotificationSnackbar {...args} id={id} message={message} />,
                })
            }>
            Show notification
        </Button>
    );
}
function NotificationExample(args: ComponentProps<typeof NotificationSnackbar>) {
    return (
        <SnackbarProvider maxSnack={1} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
            <NotificationTrigger {...args} />
        </SnackbarProvider>
    );
}
const meta: Meta<typeof NotificationSnackbar> = {
    title: 'Components/NotificationSnackbar',
    component: NotificationSnackbar,
    tags: ['autodocs'],
    args: { id: 'example-notification', message: 'Collection settings saved.', title: 'Success', variant: 'success' },
    parameters: { layout: 'centered' },
    render: (args) => <NotificationExample {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Success: Story = {};
export const Error: Story = {
    args: { title: 'Error', variant: 'error', message: 'Collection failed. Please try again.' },
};
export const Warning: Story = {
    args: { title: 'Warning', variant: 'warning', message: 'Some files could not be processed.' },
};
export const Info: Story = { args: { title: 'Information', variant: 'info', message: 'Analysis is running.' } };
export const Dismiss: Story = {
    play: async ({ canvasElement }) => {
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Show notification' }));
        const body = within(canvasElement.ownerDocument.body);
        await expect(await body.findByText('Collection settings saved.')).toBeVisible();
        await userEvent.click(body.getByRole('button', { name: 'Dismiss alert' }));
        await waitFor(() => expect(body.queryByText('Collection settings saved.')).not.toBeInTheDocument());
    },
};
