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
import SearchResultItem from './SearchResultItem';

const meta: Meta<typeof SearchResultItem> = {
    title: 'Components/SearchResultItem',
    component: SearchResultItem,
    tags: ['autodocs'],
    args: {
        item: {
            label: 'ADMINISTRATOR@EXAMPLE.COM',
            objectId: 'S-1-5-21-1000-500',
            kind: 'User',
            distinguishedName: 'CN=Administrator,CN=Users,DC=EXAMPLE,DC=COM',
        },
        index: 0,
        keyword: 'admin',
        getItemProps: () => ({}),
    },
    argTypes: { getItemProps: { control: false } },
    decorators: [
        (Story) => (
            <ul className='max-w-xl'>
                <Story />
            </ul>
        ),
    ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Highlighted: Story = { args: { highlightedIndex: 0 } };
export const DistinguishedName: Story = { args: { showDistinguishedName: true } };
export const ObjectIdFallback: Story = {
    args: { item: { label: '', objectId: 'S-1-5-21-1000-500', kind: 'User' }, keyword: '500' },
};
export const LongName: Story = {
    args: {
        item: { label: 'SERVICE-ACCOUNT-WITH-A-LONG-NAME@PRODUCTION.EXAMPLE.COM', objectId: 'service-1', kind: 'User' },
        showDistinguishedName: true,
    },
};
