import {
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from '@bloodhoundenterprise/doodleui';
import { IconName, findIconDefinition } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import clsx from 'clsx';
import { AssetGroupTag } from 'js-client-library';
import { FC, HTMLProps } from 'react';
import { cn } from '../..';
import { AppIcon } from '../../components';
import { useHighestPrivilegeTagId, usePZPathParams, usePrivilegeZoneAnalysis } from '../../hooks';

type ZoneIconProps = {
    zone: AssetGroupTag | undefined;
    size?: number;
    tooltipMessage?: string;
    iconClasses?: HTMLProps<HTMLElement>['className'];
    wrapperClasses?: HTMLProps<HTMLElement>['className'];
};

export const ZoneIcon: FC<ZoneIconProps> = ({ zone, size = 24, tooltipMessage, iconClasses, wrapperClasses }) => {
    const { hasLabelId } = usePZPathParams();
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();
    const { tagId: topTagId } = useHighestPrivilegeTagId();

    if (!zone) return null;

    const { analysis_enabled, glyph } = zone;
    const ariaLabel = !privilegeZoneAnalysisEnabled ? 'Upgrade available' : 'Analysis disabled';
    const iconProps = {
        size,
        'aria-label': ariaLabel,
        role: 'img',
        className: clsx(
            iconClasses,
            !privilegeZoneAnalysisEnabled && 'mb-0.5 text-[#ED8537]',
            privilegeZoneAnalysisEnabled && 'text-[#8E8C95]'
        ),
    };

    const upgradeIcon = <AppIcon.DataAlert {...iconProps} data-testid='analysis_upgrade_icon' />;
    const disabledIcon = (
        <AppIcon.Disabled {...iconProps} width={size} height={size} data-testid='analysis_disabled_icon' />
    );
    // const tierZeroIcon = <AppIcon.TierZero {...iconProps} data-testid='tier_zero_icon' />;

    if (hasLabelId) return null;

    if (zone.id === topTagId) return <AppIcon.TierZero className='mr-2' size={18} />;
    // {zone.id === topTagId && <AppIcon.TierZero className='mr-2' size={18} />}

    if (privilegeZoneAnalysisEnabled && analysis_enabled) {
        // TODO need to check for Tier Zero and use <AppIcon.TierZero className='ml-2' size={18} />

        const iconDefiniton = findIconDefinition({ prefix: 'fas', iconName: glyph as IconName });

        return (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={cn('min-w-4 w-4 mr-2 flex justify-center', wrapperClasses)}>
                            {iconDefiniton ? <FontAwesomeIcon icon={iconDefiniton} /> : <AppIcon.Zones size={size} />}
                        </div>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                            {tooltipMessage ||
                                (zone.glyph ? `glyph of ${zone.glyph} for ${zone.name}` : `${zone.name}`)}
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>
        );
    }
    // TODO create case for Hygiene
    return (
        <TooltipProvider>
            <TooltipRoot>
                <TooltipTrigger>
                    <div className={cn('min-w-4 w-4 mr-2 flex justify-center', wrapperClasses)}>
                        {!privilegeZoneAnalysisEnabled ? upgradeIcon : disabledIcon}
                    </div>
                </TooltipTrigger>
                <TooltipPortal>
                    <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                        {!privilegeZoneAnalysisEnabled ? 'Upgrade available' : 'Analysis disabled'}
                    </TooltipContent>
                </TooltipPortal>
            </TooltipRoot>
        </TooltipProvider>
    );
};
