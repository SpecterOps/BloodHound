import { TooltipProvider, TooltipRoot, TooltipTrigger, TooltipPortal, TooltipContent } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppIcon } from '../../components';
import clsx from 'clsx';
import { usePrivilegeZoneAnalysis } from '../../hooks';

type TierAnalysisIconProps = {
    iconClasses?: string | null,
    size?: number;
    tooltip?: boolean,
    wrapperClasses?: string,
};

export const TierAnalysisIcon: FC<TierAnalysisIconProps> = ({ iconClasses, size = 24, tooltip, wrapperClasses }) => {
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();

    if (!privilegeZoneAnalysisEnabled) {
        return
    }

    return tooltip ? (
        <TooltipProvider>
            <TooltipRoot>
                <TooltipTrigger>
                    <div className={clsx(wrapperClasses)} >
                        <AppIcon.DataAlert
                            size={size}
                            data-testid='analysis_disabled_icon'
                            className={clsx(iconClasses, 'mb-0.5 mr-2 text-[#ED8537]')} />
                    </div>
                </TooltipTrigger>
                <TooltipPortal>
                    <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                        Analysis disabled
                    </TooltipContent>
                </TooltipPortal>
            </TooltipRoot>
        </TooltipProvider>
    ) : (
        <AppIcon.DataAlert
            size={size}
            className={clsx(iconClasses, 'mb-0.5 mr-2 text-[#ED8537]')}
            data-testid='analysis_disabled_icon'
        />
    )
};