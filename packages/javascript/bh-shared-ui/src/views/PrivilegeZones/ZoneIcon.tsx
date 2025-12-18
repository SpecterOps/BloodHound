// Copyright 2025 Specter Ops, Inc.
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
import { HYGIENE_AGT_ID, useHighestPrivilegeTagId, usePZPathParams, usePrivilegeZoneAnalysis } from '../../hooks';

type ZoneIconProps = {
    zone?: Pick<AssetGroupTag, 'name' | 'glyph'> & { analysis_enabled?: boolean | null; id?: number };
    size?: number;
    tooltipMessage?: string;
    persistGlyph?: boolean; // use to escape logic to force render icon
    iconClasses?: HTMLProps<HTMLElement>['className'];
    wrapperClasses?: HTMLProps<HTMLElement>['className'];
};

export const ZoneIcon: FC<ZoneIconProps> = ({
    zone,
    size = 18,
    persistGlyph = false,
    tooltipMessage,
    iconClasses,
    wrapperClasses,
}) => {
    const { hasLabelId } = usePZPathParams();
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();
    const { tagId: topTagId } = useHighestPrivilegeTagId();

    if (hasLabelId) return null;

    const { analysis_enabled, glyph } = zone ?? {};
    const ariaLabel = !privilegeZoneAnalysisEnabled ? 'Upgrade available' : 'Analysis disabled';
    const iconProps = {
        size,
        'aria-label': ariaLabel,
        role: 'img',
        className: clsx(
            iconClasses,
            !privilegeZoneAnalysisEnabled && 'mb-0.5 -ml-1 text-link',
            privilegeZoneAnalysisEnabled && 'text-[#8E8C95]'
        ),
    };

    const upgradeIcon = <AppIcon.DataAlert {...iconProps} data-testid='analysis_upgrade_icon' />;
    const disabledIcon = <AppIcon.Disabled {...iconProps} data-testid='analysis_disabled_icon' />;
    const tierZeroIcon = <AppIcon.TierZero className='mr-2' size={size} data-testid='tier_zero_icon' />;
    const hygieneIcon = <AppIcon.Shield className='mr-2' />;
    const iconDefinition = findIconDefinition({ prefix: 'fas', iconName: glyph as IconName });

    if (zone?.id === topTagId) return tierZeroIcon;
    if (zone?.id === HYGIENE_AGT_ID) return hygieneIcon;

    if ((privilegeZoneAnalysisEnabled && analysis_enabled) || persistGlyph) {
        return (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={cn('min-w-4 w-4 mr-2 flex items-center', wrapperClasses)}>
                            {iconDefinition ? <FontAwesomeIcon icon={iconDefinition} /> : <AppIcon.Zones size={size} />}
                        </div>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                            {tooltipMessage ||
                                (zone?.glyph ? `glyph of ${zone.glyph} for ${zone.name}` : `${zone?.name}`)}
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>
        );
    }

    return (
        <TooltipProvider>
            <TooltipRoot>
                <TooltipTrigger>
                    <div className={cn('min-w-4 w-4 mr-2 flex items-center', wrapperClasses)}>
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
