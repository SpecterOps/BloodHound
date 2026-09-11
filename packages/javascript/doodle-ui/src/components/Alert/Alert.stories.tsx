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
import { useState } from 'react';
import { Button } from '../Button';
import { Alert } from './Alert';

const meta = {
    title: 'Components/Alert',
    component: Alert,
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component:
                    'A contextual feedback message for user actions. Alerts communicate status or important information and support an optional title alongside body content. Use the `variant` prop to convey different severities — `default`, `error`, `info`, `success`, or `warning`.',
            },
        },
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/doc-blocks/doc-block-argtypes
    argTypes: {
        children: {
            control: 'text',
            description: 'The body content of the alert.',
        },
        action: {
            control: 'object',
            description:
                'Optional action button rendered inside the alert. Accepts an object with a `label` string and an `onClick` handler.',
            table: {
                type: { summary: '{ label: string; onClick: () => void }' },
                defaultValue: { summary: 'undefined' },
            },
        },
        onClose: {
            control: false,
            description:
                'Optional callback invoked when the close button is clicked. When provided, a close button is rendered in the alert. When omitted, no close button is shown.',
            table: {
                type: { summary: '() => void' },
                defaultValue: { summary: 'undefined' },
            },
        },
        title: {
            control: 'text',
            description: 'Optional heading rendered above the alert body content.',
        },
        variant: {
            control: 'select',
            options: ['default', 'error', 'info', 'success', 'warning'],
            description: 'Controls the visual style of the alert to convey semantic meaning.',
            table: {
                defaultValue: { summary: 'default' },
            },
        },
        className: {
            control: 'text',
            description: 'Additional CSS class names to apply to the alert container.',
        },
    },
    args: {
        children: 'This is an alert message.',
    },
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        children: 'This is a default alert',
        title: 'Heads up!',
    },
};

export const Dismissible: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Alerts do not manage their own visibility state, but when `onClose` is provided, a close button appears in the alert. Clicking it invokes the callback — use this to remove the alert from the UI or update state in the parent component.',
            },
            source: {
                code: `
const [visible, setVisible] = useState(true);

return (
    <div className='flex flex-col items-center gap-4 w-[400px]'>
        {visible ? (
            <Alert title="I'm dismissible" onClose={() => setVisible(false)}>
                Click the close button to dismiss this alert.
            </Alert>
        ) : (
            <Button onClick={() => setVisible(true)} type='button'>
                Reset alert
            </Button>
        )}
    </div>
);`,
                language: 'tsx',
                type: 'code',
            },
        },
    },
    render: function Render() {
        const [visible, setVisible] = useState(true);

        return (
            <div className='flex flex-col items-center gap-4 w-[400px]'>
                {visible ? (
                    <Alert title="I'm dismissible" onClose={() => setVisible(false)}>
                        Click the close button to dismiss this alert.
                    </Alert>
                ) : (
                    <Button onClick={() => setVisible(true)} type='button'>
                        Reset alert
                    </Button>
                )}
            </div>
        );
    },
};

export const WithAction: Story = {
    args: {
        action: {
            label: 'Alert',
            onClick: () => alert('You clicked it!'),
        },
        children: "Click 'ALERT' to call action.",
        title: 'With Action',
    },
    parameters: {
        docs: {
            description: {
                story: 'Use the `action` prop to render a labelled button inside the alert. Provide a `label` string and an `onClick` handler — this is useful for surfacing a single contextual action alongside the alert message, such as a retry, undo, or navigation link.',
            },
        },
    },
};

export const Variants: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Each `variant` applies a distinct visual style to communicate semantic meaning. Use `default` for general messages, `info` for informational context, `success` to confirm a completed action, `warning` to signal potential issues, and `error` to indicate something went wrong.',
            },
            source: {
                code: `
<div className='flex flex-col gap-4 w-[480px]'>
    <Alert variant='default' title='Default'>
        A general-purpose alert with no specific severity.
    </Alert>
    <Alert variant='info' title='Info'>
        Here is some useful information for you.
    </Alert>
    <Alert variant='success' title='Success'>
        Your action was completed successfully.
    </Alert>    
    <Alert variant='warning' title='Warning'>
        Proceed with caution — this action may have unintended effects.
    </Alert>
    <Alert variant='error' title='Error'>
        Something went wrong. Please try again.
    </Alert>
</div>`,
                language: 'tsx',
                type: 'code',
            },
        },
    },
    render: () => (
        <div className='flex flex-col gap-4 w-[480px]'>
            <Alert variant='default' title='Default'>
                A general-purpose alert with no specific severity.
            </Alert>
            <Alert variant='info' title='Info'>
                Here is some useful information for you.
            </Alert>
            <Alert variant='success' title='Success'>
                Your action was completed successfully.
            </Alert>
            <Alert variant='warning' title='Warning'>
                Proceed with caution — this action may have unintended effects.
            </Alert>
            <Alert variant='error' title='Error'>
                Something went wrong. Please try again.
            </Alert>
        </div>
    ),
};
