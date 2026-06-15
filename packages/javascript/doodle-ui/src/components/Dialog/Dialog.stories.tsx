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
        docs: {
            description: {
                component:
                    'A modal dialog that overlays the page to focus the user on a single, self-contained task. Use it to confirm destructive actions, request input that interrupts the current flow, or surface contextual information that the user must acknowledge before continuing.',
            },
        },
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/doc-blocks/doc-block-argtypes
    argTypes: {
        DialogOverlayProps: {
            control: 'select',
            options: ['default', 'blur'],
            mapping: {
                blur: { blurBackground: true },
            },
            description:
                'Props forwarded to the underlying `DialogOverlay`. Use the `blur` option to enable a backdrop blur (`blurBackground: true`) behind the dialog.',
            table: {
                type: { summary: 'DialogOverlayProps' },
                defaultValue: { summary: '{ blurBackground: false }' },
            },
        },
        disableStickyLayout: {
            control: 'boolean',
            description:
                'When `true`, opts out of the default sticky layout. By default, `DialogTitle` and `DialogActions` are pinned to the top and bottom of `DialogContent` and any other children are placed inside an internal scrollable container, so the title and actions remain in view while the middle content scrolls. With this prop enabled, all children render in source order and the entire dialog scrolls as a single block — a difference that only becomes visible when the content overflows.',
            table: {
                type: { summary: 'boolean' },
                defaultValue: { summary: 'false' },
            },
        },
        maxWidth: {
            control: 'select',
            options: ['xl', 'lg', 'md', 'sm', 'xs'],
            description:
                'Constrains the maximum width of the dialog. Values match the MUI breakpoint scale: `xs` (444px), `sm` (600px), `md` (900px), `lg` (1200px), `xl` (1536px).',
            table: {
                type: { summary: "'xl' | 'lg' | 'md' | 'sm' | 'xs'" },
                defaultValue: { summary: 'sm' },
            },
        },
        allowNav: {
            control: 'boolean',
            description:
                'When `true`, lowers the overlay z-index and re-enables pointer events on the document body so the user can still interact with the surrounding application (for example, switching tabs) while the dialog is open.',
            table: {
                type: { summary: 'boolean' },
                defaultValue: { summary: 'false' },
            },
        },
        className: {
            control: 'text',
            description: 'Additional CSS class names to apply to the `DialogContent` container.',
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
                            Something that we want to hide visually but still want in the DOM for accessibility
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

export const DefaultDialog: Story = {
    parameters: {
        docs: {
            description: {
                story: 'A standard confirmation dialog composed of a `DialogTitle`, a `DialogDescription`, and a `DialogActions` footer. Clicking the trigger opens the dialog. Open dialogs may be dismissed either by clicking the `DialogClose` button or by clicking outside the dialog on the overlay. Use `VisuallyHidden` to provide accessible content (such as an alternate description) that should not appear visually.',
            },
            source: {
                code: `
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Default Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent>
            <DialogTitle>Are you absolutely sure?</DialogTitle>
            <VisuallyHidden>
                Something that we want to hide visually but still want in the DOM for accessibility
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
</Dialog>`,
                language: 'tsx',
                type: 'code',
            },
        },
    },
};

const loremParagraph =
    'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.';

export const OverflowingContent: Story = {
    parameters: {
        docs: {
            description: {
                story: 'When the body content exceeds the available height, only the middle section scrolls — the `DialogTitle` stays pinned to the top of the dialog and the `DialogActions` footer stays pinned to the bottom. The dialog is constrained by `max-h-[calc(100vh-32px)]` and any children other than the title and actions are placed inside an internal scrollable container. This works for both `DialogDescription` and arbitrary JSX children.',
            },
            source: {
                code: `
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Overflowing Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent>
            <DialogTitle>Terms and Conditions</DialogTitle>
            <DialogDescription>A bunch of repeating text</DialogDescription>
            {Array.from({ length: 25 }, (_, index) => (
                <p key={index} className='mb-4'>
                    {loremParagraph}
                </p>
            ))}
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Accept</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>`,
                language: 'tsx',
                type: 'code',
            },
        },
    },
    render: (args) => {
        return (
            <Dialog>
                <DialogTrigger asChild>
                    <Button variant='primary'>Overflowing Dialog</Button>
                </DialogTrigger>
                <DialogPortal>
                    <DialogContent {...args}>
                        <DialogTitle>Terms and Conditions</DialogTitle>
                        <DialogDescription>A bunch of repeating text</DialogDescription>
                        {Array.from({ length: 25 }, (_, index) => (
                            <p key={index} className='mb-4'>
                                {loremParagraph}
                            </p>
                        ))}
                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button>Accept</Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>
        );
    },
};

export const OverflowingContentWithDisabledStickyLayout: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Setting `disableStickyLayout` to `true` opts out of the default sticky layout behavior. The `DialogTitle` and `DialogActions` are no longer pinned to the top and bottom of the dialog, and children are rendered in source order without an internal scrollable container. When the body content exceeds the available height the entire `DialogContent` scrolls as a single block, so the title and actions scroll out of view along with the rest of the content. Use this when you need full control over the dialog layout.',
            },
            source: {
                code: `
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Overflowing Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent disableStickyLayout>
            <DialogTitle>Terms and Conditions</DialogTitle>
            <DialogDescription>A bunch of repeating text</DialogDescription>
            {Array.from({ length: 25 }, (_, index) => (
                <p key={index} className='mb-4'>
                    {loremParagraph}
                </p>
            ))}
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Accept</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>`,
                language: 'tsx',
                type: 'code',
            },
        },
    },
    args: { disableStickyLayout: true },
    render: (args) => {
        return (
            <Dialog>
                <DialogTrigger asChild>
                    <Button variant='primary'>Overflowing Dialog without auto </Button>
                </DialogTrigger>
                <DialogPortal>
                    <DialogContent {...args}>
                        <DialogTitle>Terms and Conditions</DialogTitle>
                        <DialogDescription>A bunch of repeating text</DialogDescription>
                        {Array.from({ length: 25 }, (_, index) => (
                            <p key={index} className='mb-4'>
                                {loremParagraph}
                            </p>
                        ))}
                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button>Accept</Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>
        );
    },
};
