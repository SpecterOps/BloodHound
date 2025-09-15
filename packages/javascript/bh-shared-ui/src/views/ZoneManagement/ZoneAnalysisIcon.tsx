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
import clsx from 'clsx';
import { FC } from 'react';
import { AppIcon } from '../../components';
import { usePrivilegeZoneAnalysis, useZonePathParams } from '../../hooks';

type ZoneAnalysisIconProps = {
    iconClasses?: string | null;
    size?: number;
    tooltip?: boolean;
    wrapperClasses?: string;
    analysisEnabled?: boolean | null;
};

export const ZoneAnalysisIcon: FC<ZoneAnalysisIconProps> = ({
    iconClasses,
    size = 24,
    tooltip,
    wrapperClasses,
    analysisEnabled,
}) => {
    const { hasLabelId } = useZonePathParams();
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();
    const ariaLabel = privilegeZoneAnalysisEnabled === false ? 'Upgrade available' : 'Analysis disabled';
    const iconProps = {
        size,
        'aria-label': ariaLabel,
        role: 'img',
        className: clsx(
            iconClasses,
            !privilegeZoneAnalysisEnabled && 'mb-0.5 text-[#ED8537]',
            privilegeZoneAnalysisEnabled && 'text-[#8E8C95]',
            'mr-2'
        ),
    };

    if (hasLabelId) return null;

    if (privilegeZoneAnalysisEnabled === false) {
        return tooltip ? (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={wrapperClasses}>
                            <AppIcon.DataAlert {...iconProps} data-testid='analysis_upgrade_icon' />
                        </div>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent className='max-w-80 dark:bg-neutral-dark-5 border-0'>
                            Upgrade Available
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>
        ) : (
            <AppIcon.DataAlert {...iconProps} data-testid='analysis_upgrade_icon' />
        );
    }

    if (privilegeZoneAnalysisEnabled && !analysisEnabled) {
        return tooltip ? (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={wrapperClasses}>
                            <AppIcon.Disabled {...iconProps} data-testid='analysis_disabled_icon' />
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
            <AppIcon.Disabled {...iconProps} data-testid='analysis_disabled_icon' />
        );
    }
};
