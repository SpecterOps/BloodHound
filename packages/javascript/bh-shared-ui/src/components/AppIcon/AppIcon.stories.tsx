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
import { AppIcon, type AppIconOptions } from './AppIcon';

const iconNames = Object.keys(AppIcon) as AppIconOptions[];
const fullLogoNames = iconNames.filter((name) => name.endsWith('LogoFull'));
const otherIconNames = iconNames.filter((name) => !name.endsWith('LogoFull'));

interface IconUsageProps {
    icon: AppIconOptions;
    size: number;
    className: string;
}

const iconMarkup = ({ icon, size, className }: IconUsageProps) =>
    `<AppIcon.${icon} size={${size}}${className ? ` className={${JSON.stringify(className)}}` : ''} />`;

const IconUsage = ({ icon, size, className }: IconUsageProps) => {
    const Icon = AppIcon[icon];
    return <Icon size={size} className={className} />;
};

const IconCard = ({ icon, size, className }: IconUsageProps) => {
    const Icon = AppIcon[icon];
    return (
        <div className='min-w-0 flex flex-col items-center gap-3 rounded-lg border border-neutral-3 bg-neutral-1 p-4'>
            <div className='flex h-20 w-full items-center justify-center'>
                <Icon size={size} className={className} />
            </div>
            <span className='w-full break-words text-center text-sm font-semibold'>{icon}</span>
            <code className='w-full whitespace-pre-wrap break-words rounded bg-neutral-2 p-2 text-xs'>
                {iconMarkup({ icon, size, className })}
            </code>
        </div>
    );
};

const AppIconGallery = () => (
    <div className='space-y-6'>
        <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            {fullLogoNames.map((icon) => (
                <IconCard key={icon} icon={icon} size={260} className='h-auto max-w-full text-primary' />
            ))}
        </div>
        <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
            {otherIconNames.map((icon) => (
                <IconCard key={icon} icon={icon} size={32} className='' />
            ))}
        </div>
    </div>
);

const meta: Meta<IconUsageProps> = {
    title: 'Components/AppIcon',
    component: IconUsage,
    tags: ['autodocs'],
    args: { icon: 'BHCELogo', size: 32, className: 'text-primary' },
    argTypes: {
        icon: {
            options: iconNames,
            control: 'select',
            description: 'The named icon to render, for example AppIcon.BHCELogo.',
        },
        size: {
            control: { type: 'number', min: 1 },
            description: 'SVG width and height in pixels. Icons default to 16 when size is omitted.',
        },
        className: {
            control: 'text',
            description: 'Use a text color utility such as text-primary to set the SVG currentColor.',
        },
    },
    parameters: {
        docs: {
            description: {
                component: `Import \`AppIcon\` from \`bh-shared-ui\`, then render a named icon such as \`<AppIcon.BHCELogo size={32} />\`.

Change the controls in Usage and expand **Show code** to copy the matching import and JSX.
The \`icon\` control selects the component name; it is not a prop passed to the icon.
Icons accept standard SVG props in addition to \`size\`. Their paths use \`currentColor\`, so a text color class or inherited color controls their appearance.`,
            },
            source: {
                type: 'dynamic',
                language: 'tsx',
                transform: (_source: string, context: { args: IconUsageProps }) =>
                    `import { AppIcon } from 'bh-shared-ui';

${iconMarkup(context.args)}`,
            },
        },
    },
};

export default meta;
type Story = StoryObj<IconUsageProps>;

export const Usage: Story = {};

export const Gallery: Story = {
    render: AppIconGallery,
    parameters: {
        controls: { disable: true },
        docs: {
            description: {
                story: 'Find an icon below. Each card includes usable JSX; Show code includes every displayed icon.',
            },
            source: {
                type: 'code',
                transform: undefined,
                code: `import { AppIcon } from 'bh-shared-ui';

<>
${[
    ...fullLogoNames.map((icon) => iconMarkup({ icon, size: 260, className: 'h-auto max-w-full text-primary' })),
    ...otherIconNames.map((icon) => iconMarkup({ icon, size: 32, className: '' })),
]
    .map((markup) => `    ${markup}`)
    .join('\n')}
</>`,
            },
        },
    },
};
