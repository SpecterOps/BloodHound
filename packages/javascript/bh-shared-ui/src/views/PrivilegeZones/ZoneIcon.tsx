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
import { usePZPathParams, usePrivilegeZoneAnalysis } from '../../hooks';

type ZoneIconProps = {
    zone: AssetGroupTag;
    iconClasses?: HTMLProps<HTMLElement>['className'];
    size?: number;
    tooltip?: boolean;
    wrapperClasses?: HTMLProps<HTMLElement>['className'];
};

export const ZoneIcon: FC<ZoneIconProps> = ({ zone, size = 24, iconClasses, wrapperClasses }) => {
    const { hasLabelId } = usePZPathParams();
    const { analysis_enabled, glyph } = zone;
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();
    const ariaLabel = !privilegeZoneAnalysisEnabled ? 'Upgrade available' : 'Analysis disabled';
    const iconProps = {
        size,
        'aria-label': ariaLabel,
        role: 'img',
        className: clsx(
            iconClasses,
            !privilegeZoneAnalysisEnabled && 'mb-0.5 text-error',
            privilegeZoneAnalysisEnabled && 'text-[#8E8C95]'
        ),
    };

    const upgradeIcon = <AppIcon.DataAlert {...iconProps} data-testid='analysis_upgrade_icon' />;
    const disabledIcon = <AppIcon.Disabled {...iconProps} data-testid='analysis_disabled_icon' />;

    if (hasLabelId) return null;

    if (privilegeZoneAnalysisEnabled && analysis_enabled) {
        const iconDefiniton = findIconDefinition({ prefix: 'fas', iconName: glyph as IconName });

        return (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={cn('min-w-2 w-2 mr-4', wrapperClasses)}>
                            {iconDefiniton ? <FontAwesomeIcon icon={iconDefiniton} /> : <AppIcon.Zones />}
                        </div>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                            {zone.name}
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
                    <div className={cn('min-w-2 w-2 mr-4', wrapperClasses)}>
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
