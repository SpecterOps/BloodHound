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
import { Button } from 'doodle-ui';
import PageWithTitle from './PageWithTitle';

function Content() {
    return (
        <section className='rounded border border-neutral-3 p-6'>
            Page content appears below the title and description.
        </section>
    );
}
function Actions() {
    return (
        <div className='flex gap-2'>
            <Button variant='secondary'>Export</Button>
            <Button>Create</Button>
        </div>
    );
}
function Description() {
    return <p>Review and manage collection environments.</p>;
}
const meta: Meta<typeof PageWithTitle> = {
    title: 'Components/PageWithTitle',
    component: PageWithTitle,
    tags: ['autodocs'],
    args: { title: 'Environments', children: <Content />, fullWidth: false },
    argTypes: {
        title: {
            control: 'text',
            description: 'Page heading and browser tab title.',
            table: { type: { summary: 'string' } },
        },
        pageDescription: {
            control: false,
            description: 'Content below the heading and actions. See With Actions And Description.',
            table: { type: { summary: 'JSX.Element' } },
        },
        actions: {
            control: false,
            description: 'Actions beside the heading. See With Actions And Description.',
            table: { type: { summary: 'React.ReactNode' } },
        },
        children: {
            control: false,
            description: 'Page content below the header.',
            table: { type: { summary: 'React.ReactNode' } },
        },
        fullWidth: {
            control: 'boolean',
            description: 'Use the available width instead of limiting the container to the xl breakpoint.',
            table: { type: { summary: 'boolean' }, defaultValue: { summary: 'false' } },
        },
        className: {
            control: 'text',
            description: 'Additional classes on the page container.',
            table: { type: { summary: 'string' } },
        },
    },
    parameters: { layout: 'fullscreen' },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const TitleAndContent: Story = {};
export const WithActionsAndDescription: Story = { args: { actions: <Actions />, pageDescription: <Description /> } };
export const FullWidth: Story = { args: { fullWidth: true, actions: <Actions />, pageDescription: <Description /> } };
export const WithoutTitle: Story = { args: { title: undefined } };
