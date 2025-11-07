// Copyright 2025 Specter Ops, Inc.
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

import { faker } from '@faker-js/faker/locale/en';
import type { Meta, StoryObj } from '@storybook/react';
import { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow } from './Table';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Table',
    component: Table,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof Table>;

export default meta;
type Story = StoryObj<typeof meta>;

const invoices = [
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'incomplete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'incomplete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'incomplete',
        description: faker.lorem.sentence(),
    },
    {
        id: faker.string.uuid(),
        todo: faker.hacker.phrase(),
        status: 'complete',
        description: faker.lorem.sentence(),
    },
];

export const Default: Story = {
    render: () => (
        <Table className='w-full'>
            <TableCaption>A list of your recent TODOs.</TableCaption>
            <TableHeader>
                <TableRow>
                    <TableHead className='w-48'>Todo</TableHead>
                    <TableHead className='w-48'>Description</TableHead>
                    <TableHead className='w-32'>Status</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {invoices.map((invoice) => (
                    <TableRow key={invoice.id}>
                        <TableCell className='font-medium'>{invoice.todo}</TableCell>
                        <TableCell>{invoice.description}</TableCell>
                        <TableCell>{invoice.status}</TableCell>
                    </TableRow>
                ))}
            </TableBody>
            <TableFooter>
                <TableRow>
                    <TableCell colSpan={3}>Completed</TableCell>
                    <TableCell className='text-right'>
                        {invoices.filter((s) => s.status === 'complete').length}
                    </TableCell>
                </TableRow>
            </TableFooter>
        </Table>
    ),
};
