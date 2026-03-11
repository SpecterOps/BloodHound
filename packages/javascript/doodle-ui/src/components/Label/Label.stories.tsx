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

import { Label } from './Label';

/**
 * Renders an accessible label associated with controls.
 */
const meta = {
    title: 'Components/Label',
    component: Label,
    tags: ['autodocs'],
    argTypes: {
        children: {
            control: { type: 'text' },
        },
        size: { control: 'select', options: ['small', 'medium', 'large'] },
    },
    args: {
        children: 'Email',
        htmlFor: 'email',
    },
} satisfies Meta<typeof Label>;

export default meta;

type Story = StoryObj<typeof Label>;

/**
 * The default form of the label.
 */
export const Default: Story = {};
