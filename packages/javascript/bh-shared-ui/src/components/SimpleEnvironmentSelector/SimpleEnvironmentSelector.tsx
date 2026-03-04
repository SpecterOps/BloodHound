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
} from '@bloodhoundenterprise/doodleui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, TextField } from '@mui/material';
import { Environment } from 'js-client-library';
import React, { ReactNode, useMemo, useState } from 'react';
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments';
import { usePZPathParams } from '../../hooks/usePZParams/usePZPathParams';
import {
    filterAndSearchEnvironments,
    getOpenGraphEnvironmentInfoMap,
    sortEnvironmentsByName,
} from '../../utils/environments';
import { cn } from '../../utils/theme';
import { DropdownTrigger, popoverContentStyles } from '../DropdownSelector';
import { SelectedEnvironment, SelectorValueTypes } from './types';

const selectedText = (
    selected: SelectedEnvironment,
    environments: Environment[] | undefined,
    isPrivilegeZonesPage: boolean
): string => {
    if (selected.type === 'active-directory-platform') {
        return 'All Active Directory Domains';
    } else if (selected.type === 'azure-platform') {
        return 'All Azure Tenants';
    } else {
        const selectedDomain: Environment | undefined = environments?.find(
            (domain: Environment) => domain.id === selected.id
        );
        if (selectedDomain) {
            return selectedDomain.name;
        } else if (isPrivilegeZonesPage) {
            return 'All Environments';
        } else {
            return 'Select Environment';
        }
    }
};

const SimpleEnvironmentSelector: React.FC<{
    selected: SelectedEnvironment;
    align?: 'center' | 'start' | 'end';
    errorMessage?: ReactNode;
    variant?: ButtonProps['variant'];
    onSelect?: (newValue: { type: SelectorValueTypes | null; id: string | null }) => void;
}> = ({ selected, align = 'start', errorMessage = '', variant, onSelect = () => {} }) => {
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
                    // To include OpenGraph environments, remove the two platform filters below
                    'active-directory': true,
                    azure: true,
                    yes: true,
                },
            }).sort(sortEnvironmentsByName),
        [availableEnvironments, searchInput]
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
        onSelect({ type: type ? `${type}-platform` : null, id: null });
        handleClose();
    };

    const handleEnvironmentClick = (environment: Environment) => {
        onSelect({ type: environment.type, id: environment.id });
        handleClose();
    };

    if (isLoading) return <Skeleton className='rounded-md w-10' />;

    if (isError) return <Alert severity='error'>{errorMessage}</Alert>;

    const selectedEnvironmentName = selectedText(selected, availableEnvironments, isPrivilegeZonesPage);

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
                            <li key={environment.id}>
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
