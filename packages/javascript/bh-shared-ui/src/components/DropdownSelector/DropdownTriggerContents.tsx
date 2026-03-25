import { Button, ButtonProps } from '@bloodhoundenterprise/doodleui';
import { cn } from '../../utils';
import { AppIcon } from '../AppIcon';
import { triggerStyles } from './constants';

type DropdownTriggerContentsProps = {
    open: boolean;
    selectedText: JSX.Element | string;
    buttonProps?: ButtonProps;
    StartAdornment?: React.FC;
    EndAdornment?: React.FC;
    testId?: string;
    variant?: ButtonProps['variant'];
    readOnly?: boolean;
};

const DropdownTriggerContents = ({
    open,
    selectedText,
    buttonProps,
    StartAdornment,
    EndAdornment,
    testId,
    variant,
    readOnly,
}: DropdownTriggerContentsProps) => {
    const buttonPrimary = variant === 'primary';

    return (
        <Button
            variant={variant}
            className={cn(
                'uppercase',
                {
                    'w-full text-sm': buttonPrimary,
                    [triggerStyles]: !buttonPrimary,
                    'bg-primary text-white border-transparent': open,
                },
                buttonProps?.className
            )}
            size='small'
            data-testid={testId ? testId : 'dropdown_context-selector'}>
            <span
                className={cn('flex justify-center items-center max-w-full', {
                    'justify-between': StartAdornment,
                })}>
                <div className='flex items-center truncate'>
                    {StartAdornment && <StartAdornment />}
                    <p className='truncate font-bold mr-2'>{selectedText}</p>
                </div>
                {EndAdornment ? (
                    <EndAdornment />
                ) : (
                    <span
                        className={cn({
                            'rotate-180 transition-transform': open,
                            'justify-self-end': buttonPrimary,
                            hidden: readOnly,
                        })}>
                        <AppIcon.CaretDown size={12} />
                    </span>
                )}
            </span>
        </Button>
    );
};

export default DropdownTriggerContents;
