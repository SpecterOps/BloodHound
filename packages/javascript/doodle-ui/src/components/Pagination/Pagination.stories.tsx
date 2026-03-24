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
