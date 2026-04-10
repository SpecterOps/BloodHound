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
import { RadioGroup, RadioItem } from './RadioGroup';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/RadioGroup',
    component: RadioGroup,
    parameters: {
        layout: 'centered',
    },
    // This story will not appear in Storybook's sidebar or docs page: https://storybook.js.org/docs/writing-stories/tags
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof RadioGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
    render: () => {
        const [radioValue, setRadioValue] = useState('a');
        return (
            <>
                <code className='mb-4 p-2 bg-slate-200 dark:bg-slate-700 rounded block'>Value = {radioValue}</code>
                <RadioGroup value={radioValue} onValueChange={(value) => setRadioValue(value)}>
                    <RadioItem value='a' label='Value a' />
                    <RadioItem value='b' label='Value b' />
                    <RadioItem value='c' label='Value c' />
                    <RadioItem value='d' label='Disabled' disabled />
                </RadioGroup>
            </>
        );
    },
};
