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
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments';
import { usePZPathParams } from '../../hooks/usePZParams/usePZPathParams';
import {
    allEnvironmentsSelected,
    filterAndSearchEnvironments,
    getOpenGraphEnvironmentInfoMap,
    sortEnvironmentsByName,
} from '../../utils/environments';
import { cn } from '../../utils/theme';
import { DropdownTrigger, popoverContentStyles } from '../DropdownSelector';
import { SelectedEnvironment } from './types';

const selectedText = (
    selected: SelectedEnvironment,
    environments: Environment[] | undefined,
    environmentInfo: ReturnType<typeof getOpenGraphEnvironmentInfoMap>,
    isPrivilegeZonesPage: boolean
): string => {
    // Check if this is an aggregate platform selection (e.g., "active-directory-platform", "azure-platform", "aws-platform")
    if (selected.type?.endsWith('-platform')) {
        // Extract the base environment type by removing the "-platform" suffix
        const baseType = selected.type.replace('-platform', '') as Environment['type'];
        // Return the aggregation display name from the environment info map
        return environmentInfo[baseType]?.aggregationDisplayName || `All ${baseType} Environments`;
    } else {
        const selectedDomain: Environment | undefined = environments?.find((domain: Environment) =>
            environmentMatchesSelection(domain, selected)
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

const environmentMatchesSelection = (environment: Environment, selected: SelectedEnvironment): boolean => {
    if (environment.id !== selected.id) return false;

    if (selected.type && environment.type !== selected.type) return false;

    // OG selectors can share an environment id across schema environment kinds, so compare schema IDs when present.
    if (selected.schema_extension_id !== undefined && selected.schema_extension_id !== null) {
        if (environment.schema_extension_id !== selected.schema_extension_id) return false;
    }

    if (selected.schema_environment_id !== undefined && selected.schema_environment_id !== null) {
        return environment.schema_environment_id === selected.schema_environment_id;
    }

    return true;
};

const environmentSelectionKey = (environment: Environment): string => {
    // Include schema IDs so React keys stay stable when multiple OG selectors point at the same environment id.
    return `${environment.type}:${environment.schema_extension_id ?? ''}:${environment.schema_environment_id ?? ''}:${environment.id}`;
};

const SimpleEnvironmentSelector: React.FC<{
    align?: 'center' | 'start' | 'end';
    errorMessage?: ReactNode;
    includeOpenGraph?: boolean;
    onSelect?: (newValue: SelectedEnvironment) => void;
    selected: SelectedEnvironment;
    variant?: ButtonProps['variant'];
}> = ({ align = 'start', errorMessage = '', includeOpenGraph = false, onSelect = () => {}, selected, variant }) => {
    const [open, setOpen] = useState<boolean>(false);
    const [searchInput, setSearchInput] = useState<string>('');
    const { data: availableEnvironments, isLoading, isError } = useAvailableEnvironments();
    const { isPrivilegeZonesPage } = usePZPathParams();

    const environmentInfo = getOpenGraphEnvironmentInfoMap(availableEnvironments);

    const filteredEnvironments = useMemo(
        () =>
            filterAndSearchEnvironments(availableEnvironments, {
                search: searchInput,
                filters: {
                    // All environments are included when there are no specific environemt filters
                    ...(includeOpenGraph ? {} : { 'active-directory': true, azure: true }),
                    yes: true,
                },
            }).sort(sortEnvironmentsByName),
        [availableEnvironments, includeOpenGraph, searchInput]
    );

    const environmentTypes = useMemo(
        () => [...new Set(filteredEnvironments?.map((environment) => environment.type))],
        [filteredEnvironments]
    );

    const handleClose = () => setOpen(false);

    const handleOpenChange: (open: boolean) => void = (open) => setOpen(open);

    const handleChange: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement> = (e) =>
        setSearchInput(e.target.value);

    const handlePlatformClick = (type?: Environment['type']) => {
        const platformEnvironment =
            type !== undefined ? availableEnvironments?.find((environment) => environment.type === type) : undefined;
        // Aggregate OG selections keep the schema IDs from a representative environment of that kind.
        const schemaExtensionID = platformEnvironment?.schema_extension_id ?? null;
        const schemaEnvironmentID = platformEnvironment?.schema_environment_id ?? null;

        onSelect({
            type: type ? `${type}-platform` : null,
            id: null,
            schema_extension_id: schemaExtensionID,
            schema_environment_id: schemaEnvironmentID,
        });
        handleClose();
    };

    const handleEnvironmentClick = (environment: Environment) => {
        onSelect({
            type: environment.type,
            id: environment.id,
            schema_extension_id: environment.schema_extension_id ?? null,
            schema_environment_id: environment.schema_environment_id ?? null,
        });
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
                                className='flex justify-between items-center gap-2 w-full'
                                onClick={() => handlePlatformClick()}
                                variant='text'>
                                All Environments
                            </Button>
                        </li>
                    )}
                    {environmentTypes?.map((type) => (
                        <li key={`${type}-platform`}>
                            <Button
                                className='flex justify-between items-center gap-2 w-full'
                                onClick={() => handlePlatformClick(type)}
                                variant={'text'}>
                                {environmentInfo[type]?.aggregationDisplayName}
                                <FontAwesomeIcon icon={environmentInfo[type]?.icon} size='sm' />
                            </Button>
                        </li>
                    ))}
                </ul>
                <ul className='max-h-80 overflow-y-auto'>
                    {filteredEnvironments?.map((environment: Environment) => {
                        return (
                            <li key={environmentSelectionKey(environment)}>
                                <Button
                                    className='flex justify-between items-center gap-2 w-full'
                                    onClick={() => handleEnvironmentClick(environment)}
                                    variant='text'>
                                    <TooltipProvider>
                                        <TooltipRoot>
                                            <TooltipTrigger>
                                                <span className='uppercase max-w-96 truncate'>{environment.name}</span>
                                            </TooltipTrigger>
                                            <TooltipPortal>
                                                <TooltipContent side='left' className='dark:bg-neutral-dark-5 border-0'>
                                                    <span className='uppercase'>{environment.name}</span>
                                                </TooltipContent>
                                            </TooltipPortal>
                                        </TooltipRoot>
                                    </TooltipProvider>
                                    <FontAwesomeIcon icon={environmentInfo[environment.type]?.icon} size='sm' />
                                </Button>
                            </li>
                        );
                    })}
                </ul>
            </PopoverContent>
        </Popover>
    );
};

export default SimpleEnvironmentSelector;
