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
import { expect, fn, userEvent, within } from '@storybook/test';
import { ComponentProps, useState } from 'react';
import FileDrop from './FileDrop';

function FileDropExample(args: ComponentProps<typeof FileDrop>) {
    const [names, setNames] = useState<string[]>([]);
    return (
        <div className='max-w-2xl'>
            <FileDrop
                {...args}
                onDrop={(files: FileList | undefined) => {
                    args.onDrop(files);
                    setNames(Array.from(files ?? [], (file) => file.name));
                }}
            />
            <p role='status' className='mt-4'>
                {names.length ? `Selected: ${names.join(', ')}` : 'No files selected.'}
            </p>
        </div>
    );
}
const meta: Meta<typeof FileDrop> = {
    title: 'Components/FileDrop',
    component: FileDrop,
    tags: ['autodocs'],
    args: { disabled: false, multiple: true, accept: ['.json', '.zip'], onDrop: fn(), className: 'w-full px-6' },
    parameters: {
        docs: {
            description: {
                component:
                    'Reusable file chooser and drop target. This example lists selected files without uploading them.',
            },
        },
    },
    render: (args) => <FileDropExample {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const MultipleFiles: Story = {};
export const SingleJsonFile: Story = { args: { multiple: false, accept: ['.json'] } };
export const Disabled: Story = { args: { disabled: true } };
export const ChooseFiles: Story = {
    play: async ({ canvasElement, args }) => {
        const canvas = within(canvasElement);
        await userEvent.upload(canvas.getByTestId('ingest-file-upload'), [
            new File(['{}'], 'users.json', { type: 'application/json' }),
            new File(['{}'], 'groups.json', { type: 'application/json' }),
        ]);
        await expect(canvas.getByRole('status')).toHaveTextContent('Selected: users.json, groups.json');
        await expect(args.onDrop).toHaveBeenCalled();
    },
};
