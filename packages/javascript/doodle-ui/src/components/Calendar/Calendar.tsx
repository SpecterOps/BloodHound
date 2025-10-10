import { ChevronLeft, ChevronRight } from 'lucide-react';
import * as React from 'react';
import { DayPicker } from 'react-day-picker';

import { ButtonVariants } from 'components/Button';
import { Select, SelectContent, SelectItem, SelectPortal, SelectTrigger, SelectValue } from 'components/Select';
import { cn } from 'components/utils';

export type CalendarProps = React.ComponentProps<typeof DayPicker>;

function Calendar({
    className,
    classNames,
    showOutsideDays = true,
    ...props
}: CalendarProps & { onChange?: React.ChangeEventHandler<HTMLSelectElement> }) {
    const handleCalendarChange = (_value: string | number, e: React.ChangeEventHandler<HTMLSelectElement>) => {
        const event = {
            target: {
                value: String(_value),
            },
        } as React.ChangeEvent<HTMLSelectElement>;
        e(event);
    };

    return (
        <DayPicker
            showOutsideDays={showOutsideDays}
            className={cn('p-3', className)}
            classNames={{
                months: 'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0',
                month: 'space-y-4',
                caption_start: 'is-start',
                caption_between: 'is-between',
                caption_end: 'is-end',
                caption: 'flex justify-center pt-1 relative items-center gap-2',
                caption_label:
                    'flex h-7 font-medium dark:text-white justify-center items-center grow [.is-multiple_&]:flex',
                caption_dropdowns: 'flex justify-center gap-1 grow dropdowns pl-8 pr-9',
                multiple_months: 'is-multiple',
                vhidden: 'hidden [.is-between_&]:flex [.is-end_&]:flex [.is-start.is-end_&]:hidden',
                nav: "flex items-center [&:has([name='previous-month'])]:order-first [&:has([name='next-month'])]:order-last gap-1",
                nav_button: cn(
                    ButtonVariants({ variant: 'text' }),
                    'h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100'
                ),
                nav_button_previous: 'absolute left-1',
                nav_button_next: 'absolute right-1',
                table: 'w-full border-collapse space-y-1',
                head_row: 'flex',
                head_cell: 'rounded-md w-9 font-normal text-[0.8rem] dark:text-neutral-light-1 opacity-70',
                row: 'flex w-full mt-2',
                cell: 'h-9 w-9 text-center text-sm p-0 relative [&:has([aria-selected]:not(.day-range))]:rounded-md [&:has([aria-selected].day-range-end)]:rounded-r-md [&:has([aria-selected].day-range-start)]:rounded-l-md [&:has([aria-selected].day-outside)]:bg-neutral-light-2 [&:has([aria-selected])]:bg-neutral-light-3 dark:[&:has([aria-selected])]:bg-neutral-dark-5 dark:[&:has([aria-selected].day-outside)]:bg-neutral-dark-2 dark:[&:has([aria-selected])]:bg-neutral-dark-3 first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md focus-within:relative focus-within:z-20',
                day: cn(
                    ButtonVariants({ variant: 'text' }),
                    'h-9 w-9 p-0 font-normal aria-selected:opacity-100 hover:no-underline rounded-md'
                ),
                day_range_end: 'day-range-end day-range',
                day_range_start: 'day-range-start day-range',
                day_selected: 'bg-primary text-white dark:bg-neutral-light-1 [&&]:dark:text-neutral-dark-0',
                day_today: '[&&]:underline underline-offset-4 decoration-2',
                day_outside: 'day-outside opacity-50 aria-selected:bg-neutral-dark-1/50 aria-selected:opacity-30',
                day_disabled: 'text-neutral-dark-5 opacity-50',
                day_range_middle:
                    'aria-selected:bg-neutral-light-3 aria-selected:text-neutral-dark-1 dark:aria-selected:bg-neutral-dark-5 [&&]:dark:aria-selected:text-neutral-light-1 day-range',
                day_hidden: 'invisible',
                ...classNames,
            }}
            components={{
                IconLeft: () => <ChevronLeft className='h-4 w-4' />,
                IconRight: () => <ChevronRight className='h-4 w-4' />,
                Dropdown: ({ ...props }) => (
                    <Select
                        {...props}
                        onValueChange={(value) => {
                            if (props.onChange) {
                                handleCalendarChange(value, props.onChange);
                            }
                        }}
                        value={props.value as string}>
                        <SelectTrigger
                            className={cn(
                                ButtonVariants({ variant: 'text' }),
                                'py-2 px-1 rounded-sm border-0 h-7 w-fit font-medium text-base dark:text-white [.is-between_&]:hidden [.is-end_&]:hidden [.is-start.is-end_&]:flex'
                            )}>
                            <SelectValue placeholder={props?.caption}>{props?.caption}</SelectValue>
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {props.children &&
                                    React.Children.map(props.children, (child) => (
                                        <SelectItem
                                            value={(child as React.ReactElement)?.props?.value}
                                            className='text-sm min-w-[var(--radix-popper-anchor-width)]'>
                                            {(child as React.ReactElement)?.props?.children}
                                        </SelectItem>
                                    ))}
                            </SelectContent>
                        </SelectPortal>
                    </Select>
                ),
            }}
            {...props}
        />
    );
}
Calendar.displayName = 'Calendar';

export { Calendar };
