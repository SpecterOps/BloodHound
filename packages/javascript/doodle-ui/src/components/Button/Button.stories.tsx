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
import { faListUl, faStar } from '@fortawesome/free-solid-svg-icons';
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
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        variant: {
            // TODO - remove transparent option
            options: ['primary', 'secondary'],
            control: 'select',
        },
        // TODO - remove fontColor
        fontColor: { options: ['primary', 'default'], control: 'select' },
        size: { options: ['small', 'medium', 'large'], control: 'select' },
        disabled: {
            control: 'boolean',
            description: 'Disables button interactions.',
        },
        render: {
            control: false,
            table: {
                disable: true,
            },
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
    },
    parameters: {
        controls: {
            exclude: ['size', 'fontColor'],
        },
    },
    render: ({ ...buttonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <IconButtonComponent {...buttonProps} aria-label='Filter'>
                    <AppIcon.FilterOutline size={24} />
                </IconButtonComponent>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-6'>
                <div className='flex flex-col items-center gap-4'>
                    <IconButtonComponent aria-label='Star Icon' variant='primary'>
                        <FontAwesomeIcon icon={faStar} />
                    </IconButtonComponent>
                    Primary
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButtonComponent aria-label='Gear Icon' variant='secondary'>
                        <FontAwesomeIcon icon={faStar} />
                    </IconButtonComponent>
                    Secondary
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButtonComponent aria-label='Gear Icon' variant='primary' disabled>
                        <AppIcon.FilterOutline />
                    </IconButtonComponent>
                    Disabled
                </div>
            </div>
        </>
    ),
};

export const DefaultType: ButtonStory = {
    render: () => {
        return <Button>Type is Button</Button>;
    },
    play: async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        const button = await canvas.findByRole('button');
        // Assert that the default type is button
        // We default to this type so that the element does not submit forms if a type is not passed
        expect(button).toHaveAttribute('type', 'button');
    },
};
