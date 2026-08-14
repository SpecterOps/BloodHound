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
import { faInfo, faListUl, faStar, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { expect, fn, within } from '@storybook/test';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { Button, IconButton as IconButtonComponent, TextButton as TextButtonComponent } from './Button';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Button',
    component: Button,
    parameters: {
        // Optional parameter to center the component in the Canvas. More info: https://storybook.js.org/docs/configure/story-layout
        layout: 'centered',
        docs: {
            description: {
                component: `
### Use a Button to perform an action. 
Use a [Link](http://localhost:6006/?path=/docs/components-link--docs) to navigate to another location.

###Use a Button when a user remains in the current context and triggers an action, such as:
- Submitting or resetting a form
- Saving, creating, deleting, or updating data
- Opening a dialog, menu, popover, or drawer
- Expanding or collapsing content
- Starting, stopping, or retrying a process
- Moving through a multi-step interface without changing the URL.

#### **Avoid choosing by appearance**
Buttons and Links may be styled similarly, but their semantics should reflect their behavior:
- Do not use a Link solely because the control looks like text.
- Do not use a Button solely because the control looks prominent.
- Do not place a Button inside a Link or a Link inside a Button.

Using the correct element provides expected keyboard behavior and helps assistive technology communicate what will happen.
                `,
            },
        },
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        render: {
            description: 'Use as a replacement for the `asChild` prop',
        },
        variant: {
            // TODO - remove transparent option
            options: ['primary', 'secondary'],
            control: 'select',
        },
        // TODO - remove fontColor
        fontColor: {
            description: '**Deprecated:** Use `TextButton` instead. This prop will be removed in BED-7642.',
            control: false,
            table: {
                type: {
                    summary: "'primary' | null",
                },
            },
        },
        size: {
            options: ['small', 'medium', 'large'],
            control: 'select',
            description: 'deprecated',
            table: {
                category: 'Deprecated',
            },
        },
        disabled: {
            control: 'boolean',
            description: 'Disables button interactions.',
        },
    },
    // Use `fn` to spy on the onClick arg, which will appear in the actions panel once invoked: https://storybook.js.org/docs/essentials/actions#action-args
    args: { onClick: fn() },
} satisfies Meta<typeof Button>;

export default meta;
type ButtonStory = StoryObj<typeof meta>;
type TextButtonStory = StoryObj<typeof TextButtonComponent>;
type IconButtonStory = StoryObj<typeof IconButtonComponent>;

// More on writing stories with args: https://storybook.js.org/docs/writing-stories/args
export const DefaultType: ButtonStory = {
    args: {
        children: 'Type is Button',
        disabled: false,
        variant: 'primary',
        size: 'medium',
    },
    argTypes: {
        type: {
            description: 'Sets the native HTML button type.',
            options: ['button', 'submit', 'reset'],
            control: {
                type: 'select',
            },
            table: {
                category: 'HTML attributes',
                defaultValue: {
                    summary: 'button',
                },
                type: {
                    summary: "'button' | 'submit' | 'reset'",
                },
            },
        },
    },
    render: ({ children, ...buttonProps }) => {
        return <Button {...buttonProps}>{children}</Button>;
    },
    play: async ({ canvasElement, args }) => {
        const canvas = within(canvasElement);
        const button = await canvas.findByRole('button');

        // Button defaults to `type="button"` so it does not
        // unintentionally submit a containing form.
        expect(button).toHaveAttribute('type', args.type ?? 'button');
    },
};

export const Primary: ButtonStory = {
    args: {
        variant: 'primary',
        children: 'Next',
        disabled: false,
    },
    render: ({ children, ...buttonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <Button {...buttonProps}>{children}</Button>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <Button>
                    {children}
                    <FontAwesomeIcon icon={faStar} />
                </Button>
                <Button variant='primary' disabled>
                    Disabled
                </Button>
                <Button variant='primary' disabled>
                    <FontAwesomeIcon icon={faStar} />
                    Disabled
                </Button>
            </div>
        </>
    ),
};

export const Secondary: ButtonStory = {
    args: {
        variant: 'secondary',
        children: 'Secondary',
        disabled: false,
    },
    render: ({ variant, children, ...buttonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <Button variant={variant} {...buttonProps}>
                    {children}
                </Button>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <Button variant='secondary'>
                    {children}
                    <FontAwesomeIcon icon={faStar} />
                </Button>
                <Button variant='secondary' disabled>
                    Disabled
                </Button>
                <Button variant='secondary' disabled>
                    <FontAwesomeIcon icon={faStar} />
                    Disabled
                </Button>
            </div>
        </>
    ),
};

export const TextButton: TextButtonStory = {
    args: {
        children: 'Help Text',
        disabled: false,
        fontColor: 'default',
    },
    argTypes: {
        fontColor: {
            control: 'select',
            options: ['primary', 'default'],
        },
    },
    render: ({ children, ...textButtonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <TextButtonComponent {...textButtonProps}>{children}</TextButtonComponent>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <TextButtonComponent>
                    {children}
                    <FontAwesomeIcon icon={faListUl} />
                </TextButtonComponent>
                <TextButtonComponent disabled>Disabled</TextButtonComponent>
                <TextButtonComponent disabled>
                    <FontAwesomeIcon icon={faListUl} />
                    Disabled
                </TextButtonComponent>
            </div>
        </>
    ),
};

export const IconButton: IconButtonStory = {
    args: {
        variant: 'default',
        disabled: false,
    },
    argTypes: {
        variant: {
            options: ['default', 'primary', 'secondary'],
            control: 'select',
        },
        children: {
            control: false,
            table: {
                disable: true,
            },
        },
        'aria-label': {
            description: 'Required accessible name describing the action performed by the icon button.',
            control: 'text',
            table: {
                category: 'Accessibility',
                type: {
                    summary: 'string',
                },
            },
        },
        size: {
            description:
                'Sets the icon width and height in pixels. Defaults to 16. The button remains square and grows when necessary to contain larger icons.',
            control: {
                type: 'number',
                min: 8,
                step: 1,
            },
            table: {
                category: 'Appearance',
                defaultValue: {
                    summary: '16',
                },
                type: {
                    summary: 'number',
                },
            },
        },
    },
    parameters: {
        controls: {
            exclude: ['fontColor'],
        },
        docs: {
            description: {
                // story: 'Icon size can be adjusted via `className`. The default is `h-8 w-8` (32×32).<br />However, adjusting the size of the icon will also adjust the overall dimensions.',
                story: `** Accessible label **
- Because an icon usually does not provide an accessible name, every \`IconButton\` requires an \`aria-label\`. The label should describe the action performed by the button.

\`\`\`tsx
<IconButton aria-label='Open settings'>
    <SettingsIcon />
</IconButton>
\`\`\`

Do not use the icon's name as the label when it does not describe the action. For example, prefer \`"Show filter options"\` over \`"Filter icon"\`.

*** Coming Soon *** -
Tooltip for IconButton`,
            },
        },
    },
    render: ({ ...buttonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <IconButtonComponent {...buttonProps} aria-label='Filter' size={24}>
                    <AppIcon.FilterOutline />
                </IconButtonComponent>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <div className='flex flex-col items-center gap-4'>
                    <IconButtonComponent aria-label='Star Icon' variant='primary'>
                        <FontAwesomeIcon icon={faInfo} />
                    </IconButtonComponent>
                    Primary
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButtonComponent aria-label='Gear Icon' variant='secondary'>
                        <FontAwesomeIcon icon={faTrash} />
                    </IconButtonComponent>
                    Secondary
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButtonComponent aria-label='Gear Icon' variant='primary' disabled size={24}>
                        <AppIcon.FilterOutline />
                    </IconButtonComponent>
                    Disabled
                </div>
            </div>
        </>
    ),
};
