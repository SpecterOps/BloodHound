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
import { expect, fn, userEvent, within } from '@storybook/test';
import { Button } from 'doodle-ui';
import { ComponentProps, useState } from 'react';
import DeleteConfirmationDialog from './DeleteConfirmationDialog';

function DeleteDialogExample(args: ComponentProps<typeof DeleteConfirmationDialog>) {
    const [open, setOpen] = useState(args.open);
    return (
        <>
            <Button onClick={() => setOpen(true)}>Delete environment</Button>
            <DeleteConfirmationDialog
                {...args}
                open={open}
                onCancel={() => {
                    args.onCancel();
                    setOpen(false);
                }}
                onConfirm={() => {
                    args.onConfirm();
                    setOpen(false);
                }}
            />
        </>
    );
}
const meta: Meta<typeof DeleteConfirmationDialog> = {
    title: 'Components/DeleteConfirmationDialog',
    component: DeleteConfirmationDialog,
    tags: ['autodocs'],
    args: { open: false, itemName: 'Production', itemType: 'environment', onCancel: fn(), onConfirm: fn() },
    argTypes: { open: { control: false } },
    parameters: { layout: 'centered' },
    render: (args) => <DeleteDialogExample {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Loading: Story = { args: { isLoading: true } };
export const Error: Story = { args: { error: 'Unable to delete the environment. Try again.' } };
export const ConfirmChallenge: Story = {
    play: async ({ canvasElement, args }) => {
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Delete environment' }));
        const body = within(canvasElement.ownerDocument.body);
        const confirm = await body.findByRole('button', { name: 'Confirm' });
        await expect(confirm).toBeDisabled();
        await userEvent.type(body.getByPlaceholderText('Delete this environment'), 'Delete this environment');
        await expect(confirm).toBeEnabled();
        await userEvent.click(confirm);
        await expect(args.onConfirm).toHaveBeenCalled();
    },
};
