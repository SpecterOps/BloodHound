import { Meta, StoryObj } from '@storybook/react';

import { useState } from 'react';
import { Pagination } from './Pagination';
// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Pagination',
    component: Pagination,
    parameters: {
        layout: 'center',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof Pagination>;

export default meta;
type Story = StoryObj<typeof meta>;

const PaginationController: React.FC<Omit<Story['args'], 'onPageChange' | 'onRowsPerPageChange'>> = (props) => {
    const [state, setState] = useState({ ...props });

    const onPageChange = (page: number) => {
        setState((prev) => ({ ...prev, page }));
    };

    const handleRowsPerPageChange = (rows: number) => {
        const rowsPerPage = rows ?? 10;
        setState((prev) => ({ ...prev, rowsPerPage }));
    };

    return <Pagination {...state} onPageChange={onPageChange} onRowsPerPageChange={handleRowsPerPageChange} />;
};

export const PaginationControls: Story = {
    args: {
        page: 0,
        rowsPerPage: 10,
        count: 123,
        onPageChange() {},
        onRowsPerPageChange() {},
    },
    render: (props) => {
        return <PaginationController {...props} />;
    },
};
