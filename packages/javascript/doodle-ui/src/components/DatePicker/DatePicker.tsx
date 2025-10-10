import { faCalendarDay } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Button,
    Calendar,
    CalendarProps,
    Input,
    InputProps,
    Popover,
    PopoverContent,
    PopoverTrigger,
} from 'components';
import { cn } from 'components/utils';
import { InputHTMLAttributes, forwardRef, useState } from 'react';

interface DatePickerProps extends InputHTMLAttributes<HTMLInputElement>, InputProps {
    calendarProps?: CalendarProps;
    InputElement?: React.ForwardRefExoticComponent<InputProps & React.RefAttributes<HTMLInputElement>>;
    onOpenChange?: (open: boolean) => void;
}

export const DatePicker = forwardRef<HTMLInputElement, DatePickerProps>(
    ({ calendarProps, InputElement, onOpenChange, ...props }, ref) => {
        const [calendarOpen, setCalendarOpen] = useState<boolean>(false);

        const handleOpenChange = (open: boolean) => {
            setCalendarOpen(open);
            onOpenChange?.(open);
        };

        return (
            <Popover onOpenChange={handleOpenChange} open={calendarOpen}>
                <div className='relative'>
                    {InputElement ? (
                        <InputElement
                            {...props}
                            className={cn('bg-neutral-light-1 dark:bg-neutral-dark-1 peer', props.className)}
                            ref={ref}
                        />
                    ) : (
                        <Input
                            variant={'outlined'}
                            {...props}
                            className={cn('bg-neutral-light-1 dark:bg-neutral-dark-1 peer', props.className)}
                        />
                    )}
                    <PopoverTrigger asChild>
                        <Button
                            variant={'text'}
                            className='absolute right-2 top-2 p-0 h-6 opacity-50 peer-hover:opacity-100'
                            aria-label='Choose Date'>
                            <FontAwesomeIcon aria-hidden='true' size='xl' icon={faCalendarDay} />
                        </Button>
                    </PopoverTrigger>
                </div>
                <PopoverContent
                    align='end'
                    alignOffset={-7}
                    className='w-auto p-3 border border-neutral-light-5 dark:border-neutral-light-1 z-[1500]'>
                    <Calendar classNames={{ day_selected: 'bg-primary text-neutral-light-1' }} {...calendarProps} />
                </PopoverContent>
            </Popover>
        );
    }
);

DatePicker.displayName = 'DatePicker';
