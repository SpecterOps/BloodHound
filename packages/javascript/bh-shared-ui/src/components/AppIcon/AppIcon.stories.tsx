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

const AppIconGallery = () => (
    <div className='space-y-8'>
        <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            {fullLogoNames.map((name) => {
                const Icon = AppIcon[name];

                return (
                    <div key={name} className='flex flex-col items-center gap-2 rounded border border-neutral-3 p-4'>
                        <Icon size={260} className='text-primary' />
                        <span>{name}</span>
                    </div>
                );
            })}
        </div>
        <div className='grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6'>
            {otherIconNames.map((name) => {
                const Icon = AppIcon[name];

                return (
                    <div key={name} className='flex flex-col items-center gap-2 rounded border border-neutral-3 p-4'>
                        <Icon size={32} />
                        <span className='text-center text-sm'>{name}</span>
                    </div>
                );
            })}
        </div>
    </div>
);

const meta = {
    title: 'Components/AppIcon',
    component: AppIconGallery,
    tags: ['autodocs'],
    parameters: {
        docs: {
            description: {
                component: 'Use an icon with `<AppIcon.CaretDown size={12} />`.',
            },
        },
    },
} satisfies Meta<typeof AppIconGallery>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Gallery: Story = {};
