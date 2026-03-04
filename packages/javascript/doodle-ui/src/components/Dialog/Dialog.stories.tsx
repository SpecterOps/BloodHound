import { DialogClose, DialogPortal } from '@radix-ui/react-dialog';
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from 'components/Button';
import { VisuallyHidden } from 'components/VisuallyHidden';
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
