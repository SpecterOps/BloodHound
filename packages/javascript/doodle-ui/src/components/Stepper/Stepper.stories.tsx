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
import { type ComponentProps, useState } from 'react';
import { Button } from '../Button';
import { Typography } from '../Typography';
import { BasicStepper, useStepper } from './index';

const meta = {
    title: 'Components/Stepper',
    component: BasicStepper,
    parameters: {
        layout: 'padded',
        docs: {
            description: {
                component:
                    'BasicStepper is a preset for simple, ordered flows. Steps can include an optional description below the title. ' +
                    'Use value/onValueChange to share state with controls outside the stepper, or useStepper() in step content for local navigation. ' +
                    'Use the composable Stepper primitives for custom layouts, per-step indicators, and richer workflows. ' +
                    'Disabled steps reject clicks and useStepper navigation requests without calling onValueChange. ' +
                    'Parent controls must enable the destination before setting value; update both together after validation succeeds. ' +
                    'Values are 1-based. Inactive panels stay hidden while their children unmount, unless forceMount is enabled. ' +
                    'Content uses normal document layout. In a container with a constrained height, use className="flex-1" to keep navigation visible while the panel scrolls.',
            },
        },
    },
    argTypes: {
        forceMount: {
            description: 'Keep inactive step content mounted and hidden to retain form registration and validation.',
            control: 'boolean',
            table: { defaultValue: { summary: 'false' } },
        },
        panelHeader: {
            description: 'Shared content above the step panels that remains mounted during navigation.',
            control: false,
        },
        navClassName: {
            description: 'Additional classes on the navigation, such as clearance for a drawer close button.',
            control: 'text',
        },
        steps: {
            description:
                'Step titles, optional descriptions, content, loading state, and disabled state. Enable disabled steps before navigating to them.',
            control: {
                type: 'object',
            },
        },
        defaultValue: {
            description: 'The default step to be selected (1-based index).',
            control: {
                type: 'number',
            },
            table: { type: { summary: 'number' }, defaultValue: { summary: '1' } },
        },
        value: {
            description:
                'The controlled active step (1-based). The parent accepts navigation requests via onValueChange and must enable the destination before selecting it.',
            control: { type: 'number' },
            table: { type: { summary: 'number' } },
        },
        onValueChange: {
            description: 'Called when a trigger or useStepper().setActiveStep requests an enabled step.',
            control: false,
            table: { type: { summary: '(nextStep: number) => void' } },
        },
        indicators: { table: { type: { summary: 'StepIndicators' } } },
    },
    decorators: [
        (Story) => (
            <div className='mx-auto w-full max-w-[796px]'>
                <Story />
            </div>
        ),
    ],
    tags: ['autodocs'],
} satisfies Meta<typeof BasicStepper>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {
        steps: [
            { title: 'Profile', content: 'Profile content' },
            { title: 'Secret', content: 'Secret Content' },
            { title: 'Schedule', content: 'Schedule Content' },
        ],
    },
};

export const WithDescriptions: Story = {
    args: {
        steps: [
            { title: 'Profile', description: 'Name and details', content: 'Profile content' },
            { title: 'Secret', description: 'Access credentials', content: 'Secret content' },
            { title: 'Schedule', content: 'Schedule content' },
        ],
    },
};

function PanelHeaderStepper(args: ComponentProps<typeof BasicStepper>) {
    return (
        <BasicStepper
            {...args}
            panelHeader={
                <div className='mb-6 space-y-2'>
                    <Typography variant='h2'>Create a collection plan</Typography>
                    <Typography variant='body2'>
                        Configure the profile, credentials, and schedule for your collection.
                    </Typography>
                </div>
            }
        />
    );
}

export const WithPanelHeader: Story = {
    args: Default.args,
    render: (args) => <PanelHeaderStepper {...args} />,
    parameters: {
        docs: {
            description: {
                story: 'Use panelHeader for a shared title, description, or summary above the step content. Select another step to see the header stay in place while the content changes.',
            },
            source: {
                code: `import { BasicStepper, Typography } from 'doodle-ui';

<BasicStepper
    steps={[
        { title: 'Profile', content: 'Profile content' },
        { title: 'Secret', content: 'Secret content' },
        { title: 'Schedule', content: 'Schedule content' },
    ]}
    panelHeader={
        <div className='mb-6 space-y-2'>
            <Typography variant='h2'>Create a collection plan</Typography>
            <Typography variant='body2'>
                Configure the profile, credentials, and schedule for your collection.
            </Typography>
        </div>
    }
/>`,
            },
        },
    },
};

export const LongerFlow: Story = {
    args: {
        steps: [
            { title: 'Account', content: 'Account Content' },
            { title: 'Profile', content: 'Profile Content' },
            { title: 'Preferences', content: 'Preferences Content' },
            { title: 'Review', content: 'Review Content' },
            { title: 'Complete', content: 'Complete Content' },
        ],
    },
};

function ControlledStepper({ defaultValue = 1, onValueChange, ...args }: ComponentProps<typeof BasicStepper>) {
    const [value, setValue] = useState(defaultValue);
    return (
        <div className='space-y-6'>
            <BasicStepper
                {...args}
                value={value}
                onValueChange={(nextStep) => {
                    setValue(nextStep);
                    onValueChange?.(nextStep);
                }}
            />
            <div className='flex items-center gap-4'>
                <Button disabled={value <= 1 || args.steps[value - 2]?.isDisabled} onClick={() => setValue(value - 1)}>
                    Back
                </Button>
                <Button
                    disabled={value >= args.steps.length || args.steps[value]?.isDisabled}
                    onClick={() => setValue(value + 1)}>
                    Next
                </Button>
                <p role='status'>
                    Step {value} of {args.steps.length}
                </p>
            </div>
        </div>
    );
}

export const Controlled: Story = {
    args: Default.args,
    render: (args) => <ControlledStepper {...args} />,
    parameters: {
        docs: {
            description: {
                story: 'Back and Next sit outside BasicStepper and share the active step through controlled props.',
            },
        },
    },
};

function StepContent({ title, isLast }: { title: string; isLast: boolean }) {
    const { activeStep, setActiveStep } = useStepper();
    return (
        <div className='space-y-4 text-center'>
            <p>{title} content</p>
            {!isLast && <Button onClick={() => setActiveStep(activeStep + 1)}>Continue</Button>}
        </div>
    );
}

function ContentNavigationStepper(args: ComponentProps<typeof BasicStepper>) {
    return (
        <BasicStepper
            {...args}
            steps={args.steps.map((step, index) => ({
                ...step,
                content: <StepContent title={step.title} isLast={index === args.steps.length - 1} />,
            }))}
        />
    );
}

export const ContentNavigation: Story = {
    args: Default.args,
    render: (args) => <ContentNavigationStepper {...args} />,
    parameters: {
        docs: {
            description: {
                story: 'Continue is inside step content and uses the existing useStepper hook. No parent state is needed.',
            },
        },
    },
};

export const DisabledStep: Story = {
    args: {
        steps: [
            { title: 'Profile', content: 'Profile content' },
            { title: 'Secret', content: 'Secret content', isDisabled: true },
            { title: 'Schedule', content: 'Schedule content' },
        ],
    },
    parameters: {
        docs: {
            description: {
                story: 'Secret is unavailable to clicks, keyboard activation, and useStepper navigation requests. Arrow keys skip it. Enable it before navigating to it.',
            },
        },
    },
};

export const DisabledInitialStep: Story = {
    args: {
        steps: [
            { title: 'Profile', content: 'Profile content', isDisabled: true },
            { title: 'Secret', content: 'Secret content' },
            { title: 'Schedule', content: 'Schedule content' },
        ],
    },
    parameters: {
        docs: {
            description: {
                story: 'If the selected step is disabled, Tab reaches the first enabled trigger. Selection stays unchanged until an enabled step is activated.',
            },
        },
    },
};

function DisabledNavigationStepper({ steps, onValueChange, ...props }: ComponentProps<typeof BasicStepper>) {
    const [value, setValue] = useState(1);
    const [isSecretEnabled, setIsSecretEnabled] = useState(false);

    return (
        <div className='space-y-6'>
            <BasicStepper
                {...props}
                value={value}
                onValueChange={(nextStep) => {
                    setValue(nextStep);
                    onValueChange?.(nextStep);
                }}
                steps={steps.map((step, index) => ({
                    ...step,
                    isDisabled: index === 1 ? !isSecretEnabled : step.isDisabled,
                    content: index === 0 ? <StepContent title={step.title} isLast={false} /> : step.content,
                }))}
            />
            <p role='status'>
                {isSecretEnabled
                    ? 'Secret is available.'
                    : 'Complete Profile to unlock Secret. Continue is blocked until then.'}
            </p>
            <Button
                disabled={isSecretEnabled}
                onClick={() => {
                    setIsSecretEnabled(true);
                    setValue(2);
                }}>
                Complete Profile
            </Button>
        </div>
    );
}

export const DisabledNavigation: Story = {
    args: DisabledStep.args,
    render: (args) => <DisabledNavigationStepper {...args} />,
    parameters: {
        docs: {
            description: {
                story: 'Continue uses useStepper and cannot enter disabled Secret. Complete Profile enables Secret and sets the controlled value in the same update, as a parent would after successful validation.',
            },
        },
    },
};

export const Loading: Story = {
    args: {
        steps: [
            { title: 'Profile', content: 'Profile content' },
            { title: 'Secret', content: 'Secret Content', isLoading: true },
            { title: 'Schedule', content: 'Schedule Content', isDisabled: true },
        ],
        defaultValue: 2,
    },
};

export const CompletedSteps: Story = {
    args: {
        steps: [
            { title: 'Profile', content: 'Profile content' },
            { title: 'Secret', content: 'Secret content' },
            { title: 'Schedule', content: 'Schedule content', isCompleted: true },
        ],
    },
    parameters: {
        docs: {
            description: {
                story: 'Completed steps are visually distinct but can still be selected.',
            },
        },
    },
};
