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

import { useArgs } from '@storybook/preview-api';
import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { DateRangeInputs } from './DateRangeInputs';

const meta: Meta<typeof DateRangeInputs> = {
    title: 'Components/DateRangeInputs',
    component: DateRangeInputs,
    tags: ['autodocs'],
    args: {
        onChange: fn(),
        onValidation: fn(),
    },
    render: function Render(args) {
        const [, updateArgs] = useArgs();

        return (
            <DateRangeInputs
                {...args}
                onChange={(changed) => {
                    args.onChange(changed);
                    updateArgs({
                        ...(Object.hasOwn(changed, 'start_time') && { start: changed.start_time }),
                        ...(Object.hasOwn(changed, 'end_time') && { end: changed.end_time }),
                    });
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Populated: Story = {
    args: {
        start: '2025-06-01T00:00:00.000-07:00',
        end: '2025-06-30T23:59:59.999-07:00',
    },
};

export const InvalidRange: Story = {
    args: {
        start: '2025-06-30T00:00:00.000-07:00',
        end: '2025-06-01T23:59:59.999-07:00',
    },
};
