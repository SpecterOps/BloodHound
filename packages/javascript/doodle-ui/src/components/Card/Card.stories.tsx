import { faGlobe } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { Badge } from 'components/Badge';
import { Button } from 'components/Button';
import { ChevronUp } from 'lucide-react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from './Card';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Card',
    component: Card,
    parameters: {
        // Optional parameter to center the component in the Canvas. More info: https://storybook.js.org/docs/configure/story-layout
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    // Use `fn` to spy on the onClick arg, which will appear in the actions panel once invoked: https://storybook.js.org/docs/essentials/actions#action-args
    args: {},
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/writing-stories/args
export const BasicCard: Story = {
    render: (args) => {
        return (
            <Card {...args}>
                <CardHeader>
                    <CardTitle>Card Title</CardTitle>
                    <CardDescription>Card Description</CardDescription>
                </CardHeader>
                <CardContent>
                    <p>Card Content</p>
                </CardContent>
                <CardFooter>
                    <p>Card Footer</p>
                </CardFooter>
            </Card>
        );
    },
};

export const NotificationsCard: Story = {
    render: (args) => {
        const notifications = [
            {
                title: 'Your call has been confirmed.',
                description: '1 hour ago',
            },
            {
                title: 'You have a new message!',
                description: '1 hour ago',
            },
            {
                title: 'Your subscription is expiring soon!',
                description: '2 hours ago',
            },
        ];
        return (
            <Card {...args}>
                <CardHeader>
                    <CardTitle>Notifications</CardTitle>
                    <CardDescription>You have 3 unread messages.</CardDescription>
                </CardHeader>
                <CardContent className='grid gap-4'>
                    <div>
                        {notifications.map((notification, index) => (
                            <div
                                key={index}
                                className='mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0'>
                                <span className='flex h-2 w-2 translate-y-1 rounded-full bg-sky-500' />
                                <div className='space-y-1'>
                                    <p className='text-sm font-medium leading-none'>{notification.title}</p>
                                    <p className='text-sm text-muted-foreground'>{notification.description}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                </CardContent>
                <CardFooter>
                    <div className='w-full text-center'>
                        <Button>Mark all as read</Button>
                    </div>
                </CardFooter>
            </Card>
        );
    },
};

export const BloodHoundUICountCard: Story = {
    render: (args) => {
        return (
            <Card {...args}>
                <div className='flex gap-12'>
                    <div>
                        <CardHeader>
                            <CardTitle>Findings</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className='text-4xl font-light mb-3'>20K</div>
                            <div className='flex items-center justify-center gap-1'>
                                <Badge label={'10'} icon={<ChevronUp />} color={'#02c577'} />
                                <span className='text-gray-400'>vs last month</span>
                            </div>
                        </CardContent>
                    </div>
                    <div className='pt-3 pr-3'>
                        <div className='bg-neutral-light-5 dark:bg-neutral-dark-5 w-8 h-8 rounded-full flex items-center justify-center'>
                            <FontAwesomeIcon icon={faGlobe} />
                        </div>
                    </div>
                </div>
            </Card>
        );
    },
};
