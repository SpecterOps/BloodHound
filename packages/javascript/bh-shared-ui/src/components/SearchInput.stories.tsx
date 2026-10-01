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
import { SearchInput } from './SearchInput';

const meta: Meta<typeof SearchInput> = {
    title: 'Components/SearchInput',
    component: SearchInput,
    tags: ['autodocs'],
    args: { id: 'search-input', value: '', onInputChange: fn() },
    render: function Render(args) {
        const [, updateArgs] = useArgs();
        return (
            <SearchInput
                {...args}
                onInputChange={(value) => {
                    args.onInputChange(value);
                    updateArgs({ value });
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = { args: { disabled: true, value: 'Example search' } };
