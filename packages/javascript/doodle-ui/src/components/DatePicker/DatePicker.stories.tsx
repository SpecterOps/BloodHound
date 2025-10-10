import InputMask from '@mona-health/react-input-mask';
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from 'components/Button';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from 'components/Form';
import { Input, InputProps } from 'components/Input';
import { DateTime } from 'luxon';
import { forwardRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { DatePicker } from './DatePicker';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/DatePicker',
    component: DatePicker,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof DatePicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    render: () => {
        return <DatePicker />;
    },
};

export const InputVariantUnderlined: Story = {
    render: () => {
        return <DatePicker variant={'underlined'} />;
    },
};

const now = new Date();
const fromDate = new Date(now);
fromDate.setFullYear(now.getFullYear() - 1);

export const BasicSingle = () => {
    const [selected, setSelected] = useState<Date | undefined>(now);

    return (
        <DatePicker
            calendarProps={{
                fromDate: fromDate,
                toDate: now,
                mode: 'single',
                selected,
                onSelect: (value) => {
                    setSelected(value);
                },
            }}
        />
    );
};

const formatString = 'yyyy-MM-dd';

export const SingleMasked = () => {
    const form = useForm<{ date: string }>({ defaultValues: { date: DateTime.local().toFormat(formatString) } });

    const InputElement = forwardRef<HTMLInputElement, InputProps>((props, ref) => {
        return (
            <InputMask
                {...props}
                id={props.name}
                mask='9999-99-99'
                maskPlaceholder=''
                placeholder='yyyy-mm-dd'
                ref={ref}>
                <Input
                    id={props.name}
                    variant={'outlined'}
                    className='rounded bg-neutral-light-1 dark:bg-neutral-dark-1'
                />
            </InputMask>
        );
    });

    return (
        <Form {...form}>
            <form
                onSubmit={form.handleSubmit(() => {
                    alert('this is your date: ' + form.getValues('date'));
                })}
                className='space-y-8'>
                <FormField
                    control={form.control}
                    name='date'
                    rules={{ required: 'No date selected. Please select a date' }}
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Date</FormLabel>
                            <FormControl>
                                <DatePicker
                                    {...field}
                                    InputElement={InputElement}
                                    calendarProps={{
                                        fromDate: fromDate,
                                        toDate: now,
                                        mode: 'single',
                                        selected: DateTime.fromFormat(field.value, formatString).toJSDate(),
                                        onSelect: (value: Date | undefined) => {
                                            form.setValue(
                                                'date',
                                                value ? DateTime.fromJSDate(value).toFormat(formatString) : ''
                                            );
                                        },
                                    }}
                                />
                            </FormControl>
                            <FormDescription>Select a Date</FormDescription>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <Button type='submit'>Submit</Button>
            </form>
        </Form>
    );
};
