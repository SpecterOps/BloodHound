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
import { faRefresh, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { useState } from 'react';
import { Button } from '../Button';
import { ConfirmationDialog } from './ConfirmationDialog';

const meta = {
    title: 'Components/ConfirmationDialog',
    component: ConfirmationDialog,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        open: {
            control: false,
        },
        text: {
            description: 'Accepts string or JSX element.  Wrap sting in quotes.',
        },
        cancelIcon: {
            description: 'Ex: `<FontAwesomeIcon icon={faRefresh} />`',
            control: false,
        },
        confirmIcon: {
            description: 'Ex: `<FontAwesomeIcon icon={faTrash} />`',
            control: false,
        },
    },
} satisfies Meta<typeof ConfirmationDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const FullExample: Story = {
    args: {
        open: false,
        title: 'Title',
        text: 'Prompt text goes here',
        onCancel: fn(),
        onConfirm: fn(),
    },
    render: (args) => {
        const [showDialog, setShowDialog] = useState(args.open);
        return (
            <>
                <Button onClick={() => setShowDialog(true)}>Full Example</Button>
                <ConfirmationDialog
                    open={showDialog}
                    title={args.title}
                    text={args.text.toString()}
                    onCancel={() => setShowDialog(false)}
                    onConfirm={() => setShowDialog(false)}
                    cancelIcon={<FontAwesomeIcon icon={faRefresh} />}
                    confirmIcon={<FontAwesomeIcon icon={faTrash} />}
                    iconPosition={args.iconPosition}
                    cancelText={args.cancelText}
                    confirmText={args.confirmText}
                    challengeTxt={args.challengeTxt}
                />
            </>
        );
    },
};

export const Basic: Story = {
    args: {
        open: false,
        title: 'Title',
        text: 'Prompt text goes here',
        onCancel: fn(),
        onConfirm: fn(),
    },
    render: (args) => {
        const [showDialog, setShowDialog] = useState(args.open);
        return (
            <>
                <Button onClick={() => setShowDialog(true)}>Basic</Button>
                <ConfirmationDialog
                    open={showDialog}
                    title={args.title}
                    text={args.text.toString()}
                    onCancel={() => setShowDialog(false)}
                    onConfirm={() => setShowDialog(false)}
                />
            </>
        );
    },
};

export const OptionalIcons: Story = {
    args: {
        open: false,
        title: 'Title',
        text: 'Prompt text goes here',
        onCancel: fn(),
        onConfirm: fn(),
        cancelIcon: <FontAwesomeIcon icon={faRefresh} />,
        confirmIcon: <FontAwesomeIcon icon={faTrash} />,
    },
    render: (args) => {
        const [showDialog, setShowDialog] = useState(args.open);
        return (
            <>
                <Button onClick={() => setShowDialog(true)}>Optional Button Icons</Button>
                <ConfirmationDialog
                    open={showDialog}
                    title='Title'
                    text='Are you sure you want to add Icons this component?'
                    onCancel={() => setShowDialog(false)}
                    onConfirm={() => setShowDialog(false)}
                    cancelIcon={args.cancelIcon}
                    confirmIcon={args.confirmIcon}
                />
            </>
        );
    },
};

export const IconPosition: Story = {
    args: {
        open: false,
        title: 'Title',
        text: 'Prompt text goes here',
        onCancel: fn(),
        onConfirm: fn(),
        cancelIcon: <FontAwesomeIcon icon={faRefresh} />,
        confirmIcon: <FontAwesomeIcon icon={faTrash} />,
    },
    render: (args) => {
        const [showDialog, setShowDialog] = useState(args.open);
        return (
            <>
                <Button onClick={() => setShowDialog(true)}>Optional Button Icons</Button>
                <ConfirmationDialog
                    open={showDialog}
                    title='Title'
                    text='Icons can be positioned left or right.'
                    onCancel={() => setShowDialog(false)}
                    onConfirm={() => setShowDialog(false)}
                    cancelIcon={args.cancelIcon}
                    confirmIcon={args.confirmIcon}
                    iconPosition='right'
                    confirmText='Delete'
                />
            </>
        );
    },
};

export const CustomButtonText: Story = {
    args: {
        open: false,
        title: 'Title',
        text: 'Prompt text goes here',
        onCancel: fn(),
        onConfirm: fn(),
        cancelIcon: <FontAwesomeIcon icon={faRefresh} />,
        confirmIcon: <FontAwesomeIcon icon={faTrash} />,
    },
    render: (args) => {
        const [showDialog, setShowDialog] = useState(args.open);
        return (
            <>
                <Button onClick={() => setShowDialog(true)}>Custom Button Text</Button>
                <ConfirmationDialog
                    open={showDialog}
                    title='Title'
                    text='Button text can be customized.  '
                    onCancel={() => setShowDialog(false)}
                    onConfirm={() => setShowDialog(false)}
                    cancelText='No'
                    confirmText='Yes'
                />
            </>
        );
    },
};

export const ChallengeText: Story = {
    args: {
        open: false,
        title: 'Title',
        text: 'Prompt text goes here',
        onCancel: fn(),
        onConfirm: fn(),
        cancelIcon: <FontAwesomeIcon icon={faRefresh} />,
        confirmIcon: <FontAwesomeIcon icon={faTrash} />,
    },
    render: (args) => {
        const [showDialog, setShowDialog] = useState(args.open);
        return (
            <>
                <Button onClick={() => setShowDialog(true)}>Challenge Text</Button>
                <ConfirmationDialog
                    open={showDialog}
                    title='Title'
                    text='Are you really sure you want to do this?'
                    challengeTxt='confirm'
                    onCancel={() => setShowDialog(false)}
                    onConfirm={() => setShowDialog(false)}
                />
            </>
        );
    },
};
