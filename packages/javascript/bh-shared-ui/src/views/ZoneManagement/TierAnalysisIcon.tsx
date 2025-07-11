import { TooltipProvider, TooltipRoot, TooltipTrigger, TooltipPortal, TooltipContent } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppIcon } from '../../components';
import clsx from 'clsx';

type TierAnalysisIconProps = {
    iconClasses?: string | null,
    size: number;
    wrapperClasses?: string,
};

export const TierAnalysisIcon: FC<TierAnalysisIconProps> = ({ iconClasses, size, wrapperClasses }) => {

    return (
        <TooltipProvider>
            <TooltipRoot>
                <TooltipTrigger>
                    <div className={clsx(wrapperClasses, 'flex flex-row items-center')} >
                        <AppIcon.DataAlert
                            size={size}
                            data-testid='analysis_disabled_icon'
                            className={clsx(iconClasses, 'mr-2 text-[#ED8537]')} />
                    </div>
                </TooltipTrigger>
                <TooltipPortal>
                    <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                        Analysis disabled
                    </TooltipContent>
                </TooltipPortal>
            </TooltipRoot>
        </TooltipProvider>
    )
};