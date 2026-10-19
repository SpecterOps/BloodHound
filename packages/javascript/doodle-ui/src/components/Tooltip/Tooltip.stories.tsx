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
import { Button } from '../Button';
import { Tooltip, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger } from './Tooltip';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Tooltip',
    component: Tooltip,
    args: {
        contentWidth: 'md',
        delayDuration: 300,
        tooltip: 'Helpful context',
        tooltipPosition: 'top',
    },
    argTypes: {
        children: {
            control: false,
            description:
                'A custom trigger element. Use a semantic, focusable element such as a button when the tooltip must be available to keyboard users.',
            table: { category: 'Trigger' },
        },
        contentProps: {
            control: false,
            description:
                'Advanced Base UI popup and positioning props. Use contentWidth and tooltipPosition for the supported sizing and side options.',
            table: { category: 'Advanced' },
        },
        contentWidth: {
            control: 'select',
            description: 'Sets the maximum width of the tooltip or matches the width of its trigger.',
            options: ['match_trigger', 'lg', 'md', 'sm'],
            table: {
                category: 'Appearance',
                defaultValue: { summary: 'md' },
                type: { summary: "'match_trigger' | 'lg' | 'md' | 'sm'" },
            },
        },
        defaultOpen: {
            control: 'boolean',
            description: 'Sets the initial open state when the tooltip manages its own state.',
            table: { category: 'State', defaultValue: { summary: 'false' } },
        },
        delayDuration: {
            control: { min: 0, type: 'number' },
            description: 'Delay in milliseconds before the tooltip opens for pointer users.',
            table: { category: 'Behavior', defaultValue: { summary: '300' } },
        },
        onOpenChange: {
            control: false,
            description: 'Called when the tooltip requests an open-state change.',
            table: { category: 'State' },
        },
        open: {
            control: 'boolean',
            description: 'Controls whether the tooltip is open.',
            table: { category: 'State' },
        },
        renderTrigger: {
            control: false,
            description: 'Advanced render callback for customizing the generated trigger.',
            table: { category: 'Advanced' },
        },
        rootProps: {
            control: false,
            description: 'Additional Base UI Tooltip Root props.',
            table: { category: 'Advanced' },
        },
        tooltip: {
            control: 'text',
            description:
                'Supplementary content associated with the trigger through aria-describedby. Keep it concise and do not repeat the trigger accessible name.',
            table: { category: 'Content' },
        },
        tooltipPosition: {
            control: 'select',
            description:
                'Sets the preferred side of the trigger. Collision avoidance may flip the tooltip when the requested side lacks space.',
            options: ['top', 'bottom', 'left', 'right'],
            table: {
                category: 'Appearance',
                defaultValue: { summary: 'top' },
                type: { summary: "'top' | 'bottom' | 'left' | 'right'" },
            },
        },
        triggerProps: {
            control: false,
            description: 'Props applied to the Base UI tooltip trigger.',
            table: { category: 'Advanced' },
        },
    },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component: `Tooltip displays concise supplementary information when its trigger is hovered or receives keyboard focus. It is built on Base UI and preserves the Doodle UI appearance, sizing, and positioning API.

### Accessibility

Tooltip content describes the trigger through <code>aria-describedby</code>, so assistive technology announces the content once without replacing the trigger's accessible name. Do not repeat the trigger label in the tooltip. For example, a button named <code>Copy</code> might use <code>Copies the value to the clipboard</code> as its tooltip.

Use a semantic, focusable custom trigger—normally a button or link—when keyboard users need access to the tooltip. When no child is supplied, Tooltip renders a focusable information icon for non-interactive contextual help. The icon does not have button semantics and does not submit forms.

~~~tsx
<Tooltip tooltip='Copies the value to the clipboard'>
    <button type='button'>Copy</button>
</Tooltip>
~~~

### Position and size

<code>tooltipPosition</code> accepts <code>top</code>, <code>bottom</code>, <code>left</code>, or <code>right</code> and defaults to <code>top</code>. These physical directions intentionally do not change in right-to-left layouts. Base UI collision avoidance may move the tooltip when there is not enough room on the preferred side.

<code>contentWidth</code> defaults to <code>md</code>. Choose <code>sm</code>, <code>md</code>, <code>lg</code>, or <code>match_trigger</code>. Keep tooltip copy short; use another disclosure pattern for rich or interactive content.

### Advanced composition

Use the high-level Tooltip component for normal usage. <code>TooltipProvider</code>, <code>TooltipRoot</code>, <code>TooltipTrigger</code>, <code>TooltipPortal</code>, and <code>TooltipContent</code> are available when custom composition is required.`,
            },
        },
    },
    tags: ['autodocs'],
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const BasicTooltip: Story = {
    args: {
        tooltip: 'This is a basic tooltip description',
    },
    parameters: {
        docs: {
            description: {
                story: 'Hover the information icon or focus it with the keyboard to show the tooltip.',
            },
        },
    },
};

export const InlineTooltip: Story = {
    args: {
        tooltip: 'Example of a tooltip in context',
        triggerProps: {
            className: 'ml-2',
        },
    },
    render: (args) => {
        return (
            <div className='flex font-bold items-center justify-center'>
                Severity
                <Tooltip {...args} />
            </div>
        );
    },
    parameters: {
        docs: {
            description: {
                story: 'The built-in information icon can sit inline with related content and remains keyboard focusable.',
            },
        },
    },
};

export const LongTooltip: Story = {
    args: {
        contentWidth: 'lg',
        tooltip: `Lorem ipsum dolor sit amet consectetur adipisicing elit. Repudiandae voluptate assumenda eius ipsum ab, eum earum corrupti provident sit quaerat quasi? Repudiandae minima esse nisi nihil nam quam vel ex?`,
    },
    render: (args) => {
        return <Tooltip {...args} />;
    },
    parameters: {
        docs: {
            description: {
                story: 'Choose a larger width only when concise tooltip content needs additional wrapping space.',
            },
        },
    },
};

export const CustomTrigger: Story = {
    args: {
        contentWidth: 'lg',
        tooltip: 'This tooltip has a custom trigger to demonstrate that you can add a tooltip to most things',
    },
    render: (args) => {
        return (
            <Tooltip {...args}>
                <Button>Example</Button>
            </Tooltip>
        );
    },
    parameters: {
        docs: {
            description: {
                story: 'Pass a semantic, focusable child to use it as the tooltip trigger.',
            },
        },
    },
};

export const Positions: Story = {
    args: {
        tooltip: 'Positioned tooltip',
    },
    parameters: {
        docs: {
            description: {
                story: 'The four supported physical positions are top, bottom, left, and right. Collision avoidance is disabled in this demonstration so each requested side remains visible.',
            },
        },
    },
    render: ({ tooltip, ...args }) => {
        return (
            <div className='grid grid-cols-2 gap-16 p-16'>
                {(['top', 'bottom', 'left', 'right'] as const).map((tooltipPosition) => (
                    <Tooltip
                        {...args}
                        contentProps={{ avoidCollisions: false }}
                        key={tooltipPosition}
                        tooltip={tooltip}
                        tooltipPosition={tooltipPosition}>
                        <Button>{tooltipPosition}</Button>
                    </Tooltip>
                ))}
            </div>
        );
    },
};

export const DeconstructedTooltip: Story = {
    args: {
        tooltip: '',
    },
    render: () => {
        return (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <button>Example</button>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent>
                            This is an example of a deconstructed tooltip. You can control much more than the shipped
                            Tooltip component. However, the shipped tooltip should cover most cases.
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>
        );
    },
    parameters: {
        docs: {
            description: {
                story: 'The exported component parts provide advanced composition while retaining the Doodle UI styling and accessibility relationship.',
            },
        },
    },
};
