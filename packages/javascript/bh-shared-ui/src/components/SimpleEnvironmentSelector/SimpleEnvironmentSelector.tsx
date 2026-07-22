// Copyright 2026 Specter Ops, Inc.
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

import { IconProp } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, TextField } from '@mui/material';
import {
    Button,
    ButtonProps,
    Popover,
    PopoverContent,
    Skeleton,
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from 'doodle-ui';
import { Environment } from 'js-client-library';
import React, { ReactNode, useMemo, useState } from 'react';
import { useFeatureFlag } from '../../hooks';
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments';
import { usePZPathParams } from '../../hooks/usePZParams/usePZPathParams';
import {
    allEnvironmentsSelected,
    filterAndSearchEnvironments,
    getOpenGraphEnvironmentInfoMap,
    sortEnvironmentsByName,
} from '../../utils/environments';
import { cn } from '../../utils/theme';
import { DropdownTrigger, optionIconStyles, optionStyles, popoverContentStyles } from '../DropdownSelector';
import { SelectedEnvironment } from './types';

const selectedText = (
    selected: SelectedEnvironment,
    environments: Environment[] | undefined,
    environmentInfo: ReturnType<typeof getOpenGraphEnvironmentInfoMap>,
    isPrivilegeZonesPage: boolean
): string => {
    const isNonOpenGraphPlatform = selected.type === 'active-directory-platform' || selected.type === 'azure-platform'; // AD and AZ don't support the new OG architecture
    const isOpenGraphPlatform = !!selected.environment_kind_id; // Only OG environments have this property
    const isPlatformAggregation = isNonOpenGraphPlatform || isOpenGraphPlatform;

    // Check if this is an aggregate platform selection (e.g., "active-directory-platform", "azure-platform", "aws-platform")
    // Aggregations dont have ids they have environment_kind_id if they are OG ones
    if (selected.type && !selected.id && isPlatformAggregation) {
        const baseType = selected.type.replace('-platform', '') as Environment['type'];
        // Return the aggregation display name from the environment info map
        return environmentInfo[baseType]?.aggregationDisplayName || `All ${baseType} Environments`;
    } else {
        const selectedDomain: Environment | undefined = environments?.find(
            (domain: Environment) => domain.id === selected.id
        );
        if (selectedDomain) {
            return selectedDomain.name;
        } else if (isPrivilegeZonesPage) {
            return allEnvironmentsSelected;
        } else {
            return 'Select Environment';
        }
    }
};

const SelectorListItemContent: React.FC<{
    displayName: string;
    displayIcon: IconProp;
    onClick: VoidFunction;
    isUpperCase?: boolean;
}> = ({ displayName, displayIcon, isUpperCase = false, onClick }) => {
    return (
        <Button
            className={cn(optionStyles, 'flex justify-between items-center gap-2')}
            onClick={onClick}
            variant={'text'}>
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <span className={cn('max-w-96 truncate', { uppercase: isUpperCase })}>{displayName}</span>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent
                            side='left'
                            className={cn('dark:bg-neutral-dark-5 border-0', { uppercase: isUpperCase })}>
                            <span>{displayName}</span>
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>
            <FontAwesomeIcon className={optionIconStyles} icon={displayIcon} size='sm' />
        </Button>
    );
};

const SimpleEnvironmentSelector: React.FC<{
    align?: 'center' | 'start' | 'end';
    errorMessage?: ReactNode;
    onSelect?: (newValue: SelectedEnvironment) => void;
    selected: SelectedEnvironment;
    variant?: ButtonProps['variant'];
}> = ({ align = 'start', errorMessage = '', onSelect = () => {}, selected, variant }) => {
    const [open, setOpen] = useState<boolean>(false);
    const [searchInput, setSearchInput] = useState<string>('');
    const { data: availableEnvironments, isLoading, isError } = useAvailableEnvironments();
    const { isPrivilegeZonesPage } = usePZPathParams();
    const { data: openGraphManagementFlag } = useFeatureFlag('opengraph_extension_management');
    const { data: openGraphFindingsFlag } = useFeatureFlag('opengraph_findings');

    const environmentInfo = getOpenGraphEnvironmentInfoMap(availableEnvironments);
    const hasOpenGraphManagement = openGraphManagementFlag?.enabled;
    const hasOpenGraphFindings = openGraphFindingsFlag?.enabled;

    const filteredEnvironments = useMemo(
        () =>
            filterAndSearchEnvironments(availableEnvironments, {
                search: searchInput,
                filters: {
                    // All environments are included when there are no specific environemt filters
                    ...(hasOpenGraphManagement && hasOpenGraphFindings
                        ? {}
                        : { 'active-directory': true, azure: true }),
                    yes: true,
                },
            }).sort(sortEnvironmentsByName),
        [availableEnvironments, hasOpenGraphManagement, hasOpenGraphFindings, searchInput]
    );

    const environmentTypes = useMemo(
        () => [...new Set(filteredEnvironments?.map((environment) => environment.type)), 're'],
        [filteredEnvironments]
    );

    const handleClose = () => setOpen(false);

    const handleOpenChange: (open: boolean) => void = (open) => setOpen(open);

    const handleChange: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement> = (e) =>
        setSearchInput(e.target.value);

    const handlePlatformClick = (
        type?: Environment['type'],
        environment_kind_id?: Environment['environment_kind_id']
    ) => {
        onSelect({ type: type ? `${type}-platform` : null, id: null, environment_kind_id });
        handleClose();
    };

    const handleEnvironmentClick = (environment: Environment) => {
        onSelect({ type: environment.type, id: environment.id });
        handleClose();
    };

    if (isLoading) return <Skeleton className='rounded-md w-10' />;

    if (isError) return <Alert severity='error'>{errorMessage}</Alert>;

    const selectedEnvironmentName = selectedText(
        selected,
        availableEnvironments,
        environmentInfo,
        isPrivilegeZonesPage
    );

    return (
        <Popover open={open} onOpenChange={handleOpenChange}>
            <DropdownTrigger
                open={open}
                selectedText={selectedEnvironmentName}
                variant={variant}
                testId='data-quality_context-selector'
            />
            <PopoverContent
                data-testid='data-quality_context-selector-popover'
                align={align}
                className={cn(popoverContentStyles, 'gap-2 p-4')}>
                <div className='flex px-0 mb-2'>
                    <TextField
                        autoFocus={true}
                        value={searchInput}
                        onChange={handleChange}
                        variant='standard'
                        fullWidth
                        label='Search'
                        data-testid={'data-quality_context-selector-search'}
                    />
                </div>
                <ul className='border-b border-neutral-light-5 pb-2 mb-2'>
                    {isPrivilegeZonesPage && (
                        <li key='all-environments'>
                            <Button
                                className={cn(optionStyles, 'flex justify-between items-center gap-2')}
                                onClick={() => handlePlatformClick()}
                                variant='text'>
                                All Environments
                            </Button>
                        </li>
                    )}
                    {environmentTypes?.map(
                        (type) =>
                            environmentInfo[type] && (
                                <li key={`${type}-platform`}>
                                    <SelectorListItemContent
                                        displayName={environmentInfo[type].aggregationDisplayName}
                                        displayIcon={environmentInfo[type].icon}
                                        onClick={() =>
                                            handlePlatformClick(type, environmentInfo[type].environment_kind_id)
                                        }
                                    />
                                </li>
                            )
                    )}
                </ul>
                <ul className='max-h-80 overflow-y-auto'>
                    {filteredEnvironments?.map((environment: Environment) => {
                        return (
                            <li key={environment.id}>
                                <SelectorListItemContent
                                    displayName={environment.name}
                                    displayIcon={environmentInfo[environment.type]?.icon}
                                    isUpperCase
                                    onClick={() => handleEnvironmentClick(environment)}
                                />
                            </li>
                        );
                    })}
                </ul>
            </PopoverContent>
        </Popover>
    );
};

export default SimpleEnvironmentSelector;
