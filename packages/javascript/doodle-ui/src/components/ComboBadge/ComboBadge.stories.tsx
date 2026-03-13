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
import { faArrowTrendUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { RadialGauge } from '../RadialGauge';
import { ComboBadge } from './ComboBadge';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/ComboBadge',
    component: ComboBadge,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        label: {
            type: 'string',
            control: 'text',
        },
        adornment: {
            type: 'string',
            control: 'text',
        },
    },
    args: {},
} satisfies Meta<typeof ComboBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SimpleImpl: Story = {
    args: {
        label: '2 📈',
        adornment: '2+',
        type: 'slideLeft',
        ariaLabel: 'x value increased by 2 points',
    },
};

export const ExtendedAdornment: Story = {
    args: {
        label: '2 📈',
        adornment: 'An unreasonably long adornment',
        type: 'slideLeft',
        ariaLabel: 'x value increased by 2 points',
    },
};

export const SlideRight: Story = {
    args: {
        label: '2 📈',
        adornment: '2+',
        type: 'slideRight',
        ariaLabel: 'x value increased by 2 points',
    },
};

export const NoAdornment: Story = {
    args: {
        label: '2 📈',
        type: 'slideRight',
        ariaLabel: 'something',
    },
};

export const InlineAdornmentFigmaSpec: Story = {
    args: {
        label: (
            <>
                2% <FontAwesomeIcon icon={faArrowTrendUp} className='ml-1' />
            </>
        ),
        adornment: '2+',
        type: 'inlineSlideLeft',
        ariaLabel: 'x value increased by 2 points',
        className: 'ml-2',
    },

    render: (props) => {
        return (
            <div className='flex justify-center items-center'>
                <RadialGauge value={50} color='primary' /> 8.1K Exposed Principles
                <ComboBadge {...props} />
            </div>
        );
    },
};
