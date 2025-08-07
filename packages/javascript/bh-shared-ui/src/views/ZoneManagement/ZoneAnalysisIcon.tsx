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
import { usePrivilegeZoneAnalysis } from '../../hooks';

type ZoneAnalysisIconProps = {
    iconClasses?: string | null;
    size?: number;
    tooltip?: boolean;
    wrapperClasses?: string;
};

export const ZoneAnalysisIcon: FC<ZoneAnalysisIconProps> = ({ iconClasses, size = 24, tooltip, wrapperClasses }) => {
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();
    const isLabelPage = location.pathname.includes('/label');
    const iconProps = {
        size,
        'aria-label': 'Analysis disabled for this tier',
        role: 'img',
        className: clsx(
            iconClasses,
            !privilegeZoneAnalysisEnabled && 'mb-0.5 text-[#ED8537]',
            privilegeZoneAnalysisEnabled && 'text-[#8E8C95]',
            'mr-2'
        ),
    };

    if (isLabelPage) return null;

    if (privilegeZoneAnalysisEnabled === false) {
        return tooltip ? (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={clsx(wrapperClasses)}>
                            <AppIcon.DataAlert {...iconProps} data-testid="analysis_upgrade_icon" />
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
            <AppIcon.DataAlert {...iconProps} />
        )
    }

    if (privilegeZoneAnalysisEnabled) {
        return tooltip ? (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <div className={clsx(wrapperClasses)}>
                            <AppIcon.Disabled {...iconProps} data-testid="analysis_disabled_icon" />
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
            <AppIcon.Disabled {...iconProps} />
        );
    }
};
