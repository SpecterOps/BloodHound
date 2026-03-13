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
import { DialogClose, DialogPortal } from '@radix-ui/react-dialog';
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from '../Button';
import { VisuallyHidden } from '../VisuallyHidden';
import { Dialog, DialogActions, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from './Dialog';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Dialog',
    component: DialogContent,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        DialogOverlayProps: {
            control: 'select',
            options: ['default', 'blur'],
            mapping: {
                blur: { blurBackground: true },
            },
        },
        maxWidth: {
            control: 'select',
            options: ['xl', 'lg', 'md', 'sm', 'xs'],
        },
    },
    args: { DialogOverlayProps: { blurBackground: false }, maxWidth: 'sm' },
    render: (args) => {
        return (
            <Dialog>
                <DialogTrigger asChild>
                    <Button variant='primary'>Default Dialog</Button>
                </DialogTrigger>
                <DialogPortal>
                    <DialogContent {...args}>
                        <DialogTitle>Are you absolutely sure?</DialogTitle>
                        <VisuallyHidden>
                            something that we want to hide visually but still want in the DOM for accessibility
                        </VisuallyHidden>
                        <DialogDescription>
                            This action cannot be undone. This will permanently delete your account and remove your data
                            from our servers.
                        </DialogDescription>
                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button>Submit</Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>
        );
    },
} satisfies Meta<typeof DialogContent>;

export default meta;
type Story = StoryObj<typeof meta>;

export const DefaultDialog: Story = {};
