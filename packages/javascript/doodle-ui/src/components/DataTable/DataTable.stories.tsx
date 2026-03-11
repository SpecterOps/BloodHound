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
import { useState } from 'react';
import { Button } from '../Button';
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
    argTypes: {
        enableResizing: {
            options: [true, false],
            control: 'boolean',
            table: { defaultValue: { summary: 'false' } },
        },
    },
    args: { enableResizing: false },
} satisfies Meta<typeof DataTable>;

const DATA = getData(10);

export default meta;
type Story = StoryObj<typeof meta>;

export const DataTableWithPagination: Story = {
    render: () => {
        return <ExampleDataTable />;
    },
};

export const ImpactTable: Story = {
    args: { enableResizing: false },
    render: ({ ...args }) => {
        return (
            <DataTable
                columns={getColumns()}
                data={getData(10)}
                TableHeadProps={{ className: 'font-bold text-base text-nowrap pl-2 pr-4' }}
                TableBodyProps={{ className: 'text-xs font-roboto' }}
                TableCellProps={{ className: 'pl-2  pr-4' }}
                columnPinning={{
                    left: ['action-menu'],
                }}
                enableResizing={args.enableResizing}
            />
        );
    },
};

const metaLabelNote = `
// Note: add meta label to columnDef in order to get proper measurements on the header content. 
// This is used in the Double Click to Expand functionality
header: ...,
label: ...,
meta: { label: 'Tier Zero Principal' }
// Otherwise we grab the header.id which may not match the label text
`;

export const ResizableColumns: Story = {
    args: {
        enableResizing: true,
    },
    parameters: {
        docs: {
            description: {
                story: `<ul className='list-disc'>
                            <li>Resizing is disabled by default</li>
                            <li>
                                Resizing can be enabled by adding the top level&nbsp;
                                <code>enableResizing</code> prop
                            </li>
                            <li>Pinned Columns are resizable</li>

                            <li>
                                Individual Columns can have resizing disabled by setting &nbsp;
                                <code>enableResizing: false </code> in the columnDef
                            </li>
                            <li>Double Click on Resize Indicator expand column to accommodate full width of content</li>
                        </ul>
                     `,
            },
            source: {
                code: metaLabelNote,
            },
            canvas: {
                sourceState: 'shown', // Show code blocks open by default
            },
        },
    },
    render: ({ ...args }) => {
        return (
            <DataTable
                columns={getColumns()}
                data={getData(10)}
                enableResizing={args.enableResizing}
                TableHeadProps={{ className: 'font-bold text-base text-nowrap pl-2 pr-4' }}
                TableBodyProps={{ className: 'text-xs font-roboto' }}
                TableCellProps={{ className: 'pl-2  pr-4' }}
                columnPinning={{
                    left: ['action-menu'],
                }}
            />
        );
    },
};

export const ResizeAndGrowLastColumn: Story = {
    args: { enableResizing: true },
    parameters: {
        docs: {
            description: {
                story: `<ul className='list-disc'>
                                <li>
                                    Combine <code>enableResizing</code> with <code>growLastColumn</code>.
                                </li>
                                <li>
                                    This combination yeilds a table that stretches to full width, without horizontal
                                    scrolling.
                                </li>
                            </ul>`,
            },
        },
    },
    render: ({ ...args }) => {
        return (
            <DataTable
                columns={getColumns().filter((_, i) => i < 5)}
                data={getData(10)}
                enableResizing={args.enableResizing}
                TableHeadProps={{ className: 'font-bold text-base text-nowrap pl-2 pr-4' }}
                TableBodyProps={{ className: 'text-xs font-roboto' }}
                TableCellProps={{ className: 'pl-2  pr-4' }}
                columnPinning={{
                    left: ['action-menu'],
                }}
                growLastColumn
            />
        );
    },
};

export const ResizeAndPinnedColumns: Story = {
    args: { enableResizing: true },
    parameters: {
        docs: {
            description: {
                story: `<ul className='list-disc'>
                            <li>Resizable with Pinned Columns.</li>
                            <li>Pinned columns are resizable as well.</li>
                        </ul>`,
            },
        },
    },
    render: ({ ...args }) => {
        return (
            <DataTable
                columns={getColumns()}
                data={getData(10)}
                enableResizing={args.enableResizing}
                TableHeadProps={{ className: 'font-bold text-base text-nowrap pl-2 pr-4' }}
                TableBodyProps={{ className: 'text-xs font-roboto' }}
                TableCellProps={{ className: 'pl-2  pr-4' }}
                columnPinning={{
                    left: ['action-menu', 'nonTierZeroPrincipal', 'email'],
                }}
            />
        );
    },
};

export const ResetColumnSizingButton: Story = {
    args: { enableResizing: true },
    parameters: {
        docs: {
            description: {
                story: `<div>
                        <p className='mb-4'>
                            In order to set up a resize button on your table, columnSizing needs to be&nbsp;
                            <code>controlled</code> by setting the below props:
                        </p>
                        <ul className='list-disc px-8 py-4 border rounded rounded-lg mb-4'>
                            <li>
                                <code>columnSizing={columnSizing}</code>
                            </li>
                            <li>
                                <code>onColumnSizingChange={setColumnSizing}</code>
                            </li>
                        </ul>

                        <div className='mb-4'>
                            And connect to state:
                            <ul className='list-disc px-8 py-4 border rounded rounded-lg'>
                                <li>
                                    <code>const [columnSizing, setColumnSizing] = useState({});</code>
                                </li>
                            </ul>
                        </div>
                        </div>`,
            },
        },
    },
    render: ({ ...args }) => {
        const [columnSizing, setColumnSizing] = useState({});
        const resetColumns = () => {
            setColumnSizing({});
        };
        return (
            <>
                <Button onClick={resetColumns} size='small' className='m-4'>
                    Reset Column Sizing
                </Button>
                <DataTable
                    TableHeadProps={{ className: 'font-bold text-base text-nowrap pl-2 pr-4' }}
                    TableBodyProps={{ className: 'text-xs font-roboto' }}
                    TableCellProps={{ className: 'pl-2  pr-4' }}
                    data={DATA}
                    enableResizing={args.enableResizing}
                    columns={getColumns()}
                    columnPinning={{
                        left: ['action-menu', 'nonTierZeroPrincipal'],
                    }}
                    columnSizing={columnSizing}
                    onColumnSizingChange={setColumnSizing}
                />
            </>
        );
    },
};
