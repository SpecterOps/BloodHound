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
