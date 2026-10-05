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
import { fn } from '@storybook/test';
import { Button } from 'doodle-ui';
import { ComponentProps, useState } from 'react';
import PasswordDialog from './PasswordDialog';

function PasswordDialogExample(args: ComponentProps<typeof PasswordDialog>) {
    const [open, setOpen] = useState(args.open);
    return (
        <>
            <Button onClick={() => setOpen(true)}>Change password</Button>
            <PasswordDialog
                {...args}
                open={open}
                onClose={() => {
                    args.onClose();
                    setOpen(false);
                }}
                onSave={(payload) => {
                    args.onSave(payload);
                    setOpen(false);
                }}
            />
        </>
    );
}
const meta: Meta<typeof PasswordDialog> = {
    title: 'Components/PasswordDialog',
    component: PasswordDialog,
    tags: ['autodocs'],
    args: { open: false, userId: 'example-user', onClose: fn(), onSave: fn() },
    argTypes: { open: { control: false } },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component:
                    'Used for self-service and administrator password changes. Saves are logged to Actions without changing an account.',
            },
        },
    },
    render: (args) => <PasswordDialogExample {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Administrator: Story = { args: { showNeedsPasswordReset: true, initialNeedsPasswordReset: true } };
export const SelfService: Story = { args: { requireCurrentPassword: true } };
