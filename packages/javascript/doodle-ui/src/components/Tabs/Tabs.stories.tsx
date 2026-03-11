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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '.';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Tabs',
    component: Tabs,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof Tabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    render: () => {
        return (
            <Tabs defaultValue='tiers' className='w-full'>
                <TabsList className='w-full'>
                    <TabsTrigger value='tiers'>Tiers</TabsTrigger>
                    <TabsTrigger value='labels'>Labels</TabsTrigger>
                    <TabsTrigger value='certifications'>Certifications</TabsTrigger>
                    <TabsTrigger value='history'>History</TabsTrigger>
                </TabsList>
                <TabsContent value='tiers'>
                    <div className='flex justify-center'>Tiers</div>
                </TabsContent>
                <TabsContent value='labels'>
                    <div className='flex justify-center'>Labels</div>
                </TabsContent>
                <TabsContent value='certifications'>
                    <div className='flex justify-center'>Certifications</div>
                </TabsContent>
                <TabsContent value='history'>
                    <div className='flex justify-center'>History</div>
                </TabsContent>
            </Tabs>
        );
    },
};

export const StaticLayout: Story = {
    render: () => {
        return (
            <Tabs defaultValue='tiers' className='w-full'>
                <TabsList className='w-full *:h-10 *:after:font-bold *:after:-translate-y-6 *:after:invisible'>
                    <TabsTrigger value='tiers' className='after:content-["Tiers"]'>
                        Tiers
                    </TabsTrigger>
                    <TabsTrigger value='labels' className='after:content-["Labels"]'>
                        Labels
                    </TabsTrigger>
                    <TabsTrigger value='certifications' className='after:content-["Certifications"]'>
                        Certifications
                    </TabsTrigger>
                    <TabsTrigger value='history' className='after:content-["History"]'>
                        History
                    </TabsTrigger>
                </TabsList>
                <TabsContent value='tiers'>
                    <div className='flex justify-center'>Tiers</div>
                </TabsContent>
                <TabsContent value='labels'>
                    <div className='flex justify-center'>Labels</div>
                </TabsContent>
                <TabsContent value='certifications'>
                    <div className='flex justify-center'>Certifications</div>
                </TabsContent>
                <TabsContent value='history'>
                    <div className='flex justify-center'>History</div>
                </TabsContent>
            </Tabs>
        );
    },
};
