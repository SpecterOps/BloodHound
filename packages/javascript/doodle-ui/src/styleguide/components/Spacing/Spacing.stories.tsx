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
import { Spacing } from './Spacing';

const storyDescription = `<div><h4>Visual reference for Tailwind spacing values.</h4><p>Usage: <code>className="w-4"</code></p></div>`;
const meta = {
    title: 'Styleguide/Spacing',
    component: Spacing,
    tags: ['autodocs'],
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                story: storyDescription,
            },
        },
    },
} satisfies Meta<typeof Spacing>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
