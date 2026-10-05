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
import { FileStatus } from '../FileUploadDialog/types';
import FileStatusListItem from './FileStatusListItem';

const sampleFile = new File(['{}'], 'directory-users.json', { type: 'application/json' });
const meta: Meta<typeof FileStatusListItem> = {
    title: 'Components/FileStatusListItem',
    component: FileStatusListItem,
    tags: ['autodocs'],
    args: {
        file: { file: sampleFile, status: FileStatus.READY },
        percentCompleted: 0,
        onRemove: fn(),
        onRefresh: fn(),
    },
    argTypes: { file: { control: false } },
    decorators: [
        (Story) => (
            <div className='max-w-xl'>
                <Story />
            </div>
        ),
    ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Ready: Story = {};
export const Uploading: Story = {
    args: { file: { file: sampleFile, status: FileStatus.UPLOADING }, percentCompleted: 45 },
};
export const Complete: Story = { args: { file: { file: sampleFile, status: FileStatus.DONE } } };
export const Failed: Story = {
    args: { file: { file: sampleFile, status: FileStatus.FAILURE, errors: ['Upload request failed.'] } },
};
export const Remove: Story = {
    play: async ({ canvasElement, args }) => {
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Remove item' }));
        await expect(args.onRemove).toHaveBeenCalled();
    },
};
export const Retry: Story = {
    ...Failed,
    play: async ({ canvasElement, args }) => {
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Retry upload' }));
        await expect(args.onRefresh).toHaveBeenCalledWith(args.file);
    },
};
