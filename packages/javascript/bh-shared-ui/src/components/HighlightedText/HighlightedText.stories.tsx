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
import HighlightedText from './HighlightedText';

const meta: Meta<typeof HighlightedText> = {
    title: 'Components/HighlightedText',
    component: HighlightedText,
    tags: ['autodocs'],
    args: { text: 'ADMINISTRATOR@EXAMPLE.COM', search: 'admin' },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const CaseInsensitiveMatch: Story = {};
export const NoMatch: Story = { args: { search: 'service' } };
export const NoSearch: Story = { args: { search: undefined } };
export const LiteralSpecialCharacters: Story = {
    args: { text: 'service.account[1]@example.com', search: 'account[1]' },
};
