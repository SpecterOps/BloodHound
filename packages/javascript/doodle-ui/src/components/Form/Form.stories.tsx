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
import { SubmitHandler, useForm } from 'react-hook-form';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '.';
import { Button } from '../Button';
import { Input } from '../Input';
import { Switch } from '../Switch';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta: Meta<typeof Form> = {
    title: 'Components/Form',
    component: Form,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof Form>;

export default meta;
type Story = StoryObj<typeof meta>;

interface FormInputs {
    username: string;
    checked: boolean;
}

export const Default: Story = {
    render: () => {
        const form = useForm<FormInputs>({ defaultValues: { username: 'doodle', checked: true } });

        const onSubmit: SubmitHandler<FormInputs> = (data) => {
            alert(JSON.stringify(data));
        };
        return (
            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-8'>
                    <FormField
                        control={form.control}
                        name='username'
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Username</FormLabel>
                                <FormControl>
                                    <Input placeholder='Username' {...field} />
                                </FormControl>
                                <FormDescription>This is your public display name.</FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name='checked'
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Toggle me!</FormLabel>
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                                <FormDescription>Switch it up</FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <Button type='submit'>Submit</Button>
                </form>
            </Form>
        );
    },
};
