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
import { RiskBadge } from './RiskBadge';

const colors = {
    critical: '#B553EC',
    high: '#FF3838',
    moderate: '#FFA86D',
    low: '#FFD64C',
    resolved: '#2DCCFF',
    deprecated: '#D9D9D9',
} as const;

const colorOptions = Object.entries(colors).map(([key, hex]) => ({
    label: key,
    value: key,
    hex,
}));

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/RiskBadge',
    component: RiskBadge,
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component: `
RiskBadges are used to communicate **severity of risk** using color to indicate such risk.

### When to use
- To display **risk severity** (e.g., critical, high, moderate, low)
- To show **exposure** (e.g., percentage of a certain asset that is exposed)

RiskBadges are best used within charts, graphs and paired with severity legend.
                `,
            },
            canvas: { sourceState: 'shown' },
        },
    },
    argTypes: {
        type: { type: 'string', control: 'select', options: ['labeled', 'sm-labeled', 'sm-circle', 'md-circle'] },
        color: {
            control: 'select',
            options: colorOptions.map((c) => c.value),
            mapping: colors,
        },
        outlined: { type: 'boolean' },
    },
    args: {
        type: 'md-circle',
        color: 'primary',
        outlined: false,
        label: 'Label',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
} satisfies Meta<typeof RiskBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/writing-stories/args
export const Default: Story = {
    args: {
        type: 'md-circle',
        color: 'primary',
        outlined: false,
    },
};

export const Labeled: Story = {
    args: {
        type: 'labeled',
        color: colors.critical,
        label: 'Critical',
        outlined: false,
    },
};

export const SmallLabeled: Story = {
    args: {
        type: 'sm-labeled',
        color: colors.moderate,
        label: '50%',
        outlined: false,
    },
};

export const OutlinedLabel: Story = {
    args: {
        type: 'labeled',
        color: colors.low,
        label: 'Low',
        outlined: true,
    },
};

export const Small: Story = {
    args: {
        type: 'sm-circle',
        color: 'tertiary',
        outlined: false,
    },
};

export const Outlined: Story = {
    args: {
        type: 'md-circle',
        color: 'primary',
        outlined: true,
    },
};
