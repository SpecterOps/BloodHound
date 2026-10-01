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
import { Controls, Description, Primary, Title } from '@storybook/blocks';
import { useArgs } from '@storybook/preview-api';
import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import type { UserMinimal } from 'js-client-library';
import { rest } from 'msw';
import { UserMinimalSelect } from './UserMinimalSelect';

const users: UserMinimal[] = [
    { id: 'alex', first_name: 'Alex', last_name: 'Morgan', email_address: 'alex@example.com' },
    { id: 'sam', first_name: 'Sam', last_name: 'Lee', email_address: 'sam@example.com' },
];

const userHandlers = (responseUsers = users) => [
    rest.get('/api/v2/bloodhound-users-minimal', (_request, response, context) =>
        response(context.json({ data: { users: responseUsers } }))
    ),
];

const meta: Meta<typeof UserMinimalSelect> = {
    title: 'Components/SelectMenus/UserMinimalSelect',
    component: UserMinimalSelect,
    tags: ['autodocs'],
    args: { user: '', onSelect: fn() },
    parameters: {
        layout: 'centered',
        msw: { handlers: () => userHandlers() },
        docs: {
            // MSW 1 workers share responses across docs iframes. Mount one live
            // example at a time; the other scenarios are available in the sidebar.
            page: () => (
                <>
                    <Title />
                    <Description />
                    <Primary />
                    <Controls />
                </>
            ),
            story: { inline: false, height: '300px' },
            description: {
                component:
                    'Fetches active users through the real useUsersMinimal hook. MSW mocks GET /api/v2/bloodhound-users-minimal. Select a user to update the controlled value; None clears the selection. Loading and request failures have no distinct visual state in this component.',
            },
        },
    },
    render: function Render(args) {
        const [, updateArgs] = useArgs();
        return (
            <UserMinimalSelect
                {...args}
                onSelect={(value) => {
                    args.onSelect(value);
                    updateArgs({ user: value === '-none-' ? '' : value });
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Selected: Story = { args: { user: 'sam' } };

export const Empty: Story = {
    parameters: {
        msw: { handlers: () => userHandlers([]) },
        docs: { description: { story: 'When no users are returned, the menu contains only None.' } },
    },
};
