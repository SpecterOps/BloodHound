import { faker } from '@faker-js/faker';
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
