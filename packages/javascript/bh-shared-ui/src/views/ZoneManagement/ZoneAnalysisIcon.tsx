import { TooltipProvider, TooltipRoot, TooltipTrigger, TooltipPortal, TooltipContent } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppIcon } from '../../components';
import clsx from 'clsx';
import { usePrivilegeZoneAnalysis } from '../../hooks';

type ZoneAnalysisIconProps = {
    iconClasses?: string | null,
    size?: number;
    tooltip?: boolean,
    wrapperClasses?: string,
};

export const ZoneAnalysisIcon: FC<ZoneAnalysisIconProps> = ({ iconClasses, size = 24, tooltip, wrapperClasses }) => {
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();

    const iconProps = {
        size,
        'data-testid': 'analysis_disabled_icon',
        'aria-label': 'Analysis disabled for this tier',
        role: 'img',
        className: clsx(iconClasses, 'mb-0.5 mr-2 text-[#ED8537]')
    };

    if (!privilegeZoneAnalysisEnabled) {
        return null
    }

    return tooltip ? (
        <TooltipProvider>
            <TooltipRoot>
                <TooltipTrigger>
                    <div className={clsx(wrapperClasses)} >
                        <AppIcon.DataAlert
                            {...iconProps} />
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
            {...iconProps}
        />
    )
};