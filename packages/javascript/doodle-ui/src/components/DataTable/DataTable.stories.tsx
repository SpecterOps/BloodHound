import type { Meta, StoryObj } from '@storybook/react';
import { DataTable } from './DataTable';
import ExampleDataTable from './StorybookExample/ExampleDataTable';
import { getColumns, getData } from './StorybookExample/utils';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta: Meta<typeof DataTable> = {
    title: 'Components/DataTable',
    component: DataTable,
    parameters: {
        layout: 'fullscreen',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    // args: {},
} satisfies Meta<typeof DataTable>;

export default meta;
type Story = StoryObj<typeof meta>;

export const DataTableWithPagination: Story = {
    render: () => {
        return <ExampleDataTable />;
    },
};

export const ImpactTable: Story = {
    render: () => {
        return (
            <DataTable
                columns={getColumns()}
                data={getData(10)}
                TableHeadProps={{ className: 'font-bold text-base' }}
                TableBodyProps={{ className: 'text-xs font-roboto' }}
            />
        );
    },
};
