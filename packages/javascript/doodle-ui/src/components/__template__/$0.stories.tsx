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
import { $0 } from './$0';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/$0',
    component: $0,
    parameters: {
        layout: 'centered',
    },
    // This story will not appear in Storybook's sidebar or docs page: https://storybook.js.org/docs/writing-stories/tags
    tags: ['!dev'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof $0>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
