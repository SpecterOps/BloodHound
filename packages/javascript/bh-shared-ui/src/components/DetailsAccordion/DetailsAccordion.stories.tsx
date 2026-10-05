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
import { expect, userEvent, within } from '@storybook/test';
import { DetailsAccordion } from './DetailsAccordion';

type Detail = { id: string; title: string; description: string; disabled?: boolean };
const items: Detail[] = [
    { id: 'directory', title: 'Directory collection', description: 'Collect users, computers, and group memberships.' },
    { id: 'sessions', title: 'Session collection', description: 'Collect active sessions from reachable computers.' },
    { id: 'cloud', title: 'Cloud collection', description: 'This collection method is unavailable.', disabled: true },
];
const Header = ({ title }: Detail) => <span className='ml-2'>{title}</span>;
const Content = ({ description }: Detail) => <p className='p-4'>{description}</p>;
const Empty = () => <p className='p-4 text-neutral-5'>No collection details available.</p>;
const meta: Meta<typeof DetailsAccordion<Detail>> = {
    title: 'Components/DetailsAccordion',
    component: DetailsAccordion<Detail>,
    tags: ['autodocs'],
    parameters: {
        docs: {
            description: {
                component: `Displays a list of items with expandable details. Only one item can be expanded at a time.

- Pass components through \`Header\` and \`Content\`; each receives the item's fields as props.
- Use \`getKey\` for stable item keys and \`itemDisabled\` to prevent specific items from expanding.
- Set \`openIndex\` to expand an item initially, or \`accent\` to emphasize the headers.
- Provide \`Empty\` to render a fallback when \`items\` is \`undefined\`. An empty array renders no items.

The examples below use the same typed collection data for the header and content.`,
            },
        },
    },
    args: { items, Header, Content, Empty, getKey: (item) => item.id, itemDisabled: (item) => Boolean(item.disabled) },
    argTypes: {
        Header: { control: false },
        Content: { control: false },
        Empty: { control: false },
        getKey: { control: false },
        itemDisabled: { control: false },
    },
    decorators: [
        (Story) => (
            <div className='max-w-2xl'>
                <Story />
            </div>
        ),
    ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Collapsed: Story = {
    parameters: {
        docs: {
            description: { story: 'Expand a collection to see its details. Cloud collection is disabled.' },
        },
    },
};
export const InitiallyOpen: Story = { args: { openIndex: 0 } };
export const Accent: Story = { args: { accent: true, openIndex: 1 } };
export const SingleItem: Story = { args: { items: items[0], openIndex: 0 } };
export const MissingItems: Story = { args: { items: undefined } };
export const ExpandAndCollapse: Story = {
    play: async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        const trigger = canvas.getByRole('button', { name: 'Directory collection' });
        await userEvent.click(trigger);
        await expect(trigger).toHaveAttribute('aria-expanded', 'true');
        await expect(canvas.getByText(items[0].description)).toBeVisible();
        await expect(canvas.getByRole('button', { name: 'Cloud collection' })).toBeDisabled();
        await userEvent.click(trigger);
        await expect(trigger).toHaveAttribute('aria-expanded', 'false');
    },
};
