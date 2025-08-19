// Copyright 2023 Specter Ops, Inc.
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
    Popover,
    PopoverContent,
    PopoverTrigger,
    Skeleton,
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from '@bloodhoundenterprise/doodleui';
import { faCloud, faGlobe } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, TextField } from '@mui/material';
import clsx from 'clsx';
import { Environment } from 'js-client-library';
import React, { ReactNode, useState } from 'react';
import { AppIcon } from '../../../components/AppIcon';
import { useAvailableEnvironments } from '../../../hooks/useAvailableEnvironments';
import { cn } from '../../../utils/theme';
import { SelectedEnvironment, SelectorValueTypes } from './types';

const selectedText = (selected: SelectedEnvironment, environments: Environment[] | undefined): string => {
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
        } else {
            return 'Select Environment';
        }
    }
};

const SimpleEnvironmentSelector: React.FC<{
    selected: SelectedEnvironment;
    errorMessage: ReactNode;
    buttonPrimary?: boolean;
    onSelect?: (newValue: { type: SelectorValueTypes; id: string | null }) => void;
}> = ({ selected, errorMessage, buttonPrimary = true, onSelect = () => {} }) => {
    const [open, setOpen] = useState<boolean>(false);
    const [searchInput, setSearchInput] = useState<string>('');
    const { data, isLoading, isError } = useAvailableEnvironments();

    if (isLoading) return <Skeleton className='rounded-md w-10' />;

    if (isError) return <Alert severity='error'>{errorMessage}</Alert>;

    const handleClose = () => setOpen(false);

    const filteredEnvironments = data?.filter(
        (environment: Environment) =>
            environment.name.toLowerCase().includes(searchInput.toLowerCase()) && environment.collected
    );

    const selectedEnvironmentName = selectedText(selected, data);

    return (
        <Popover
            open={open}
            onOpenChange={() => {
                setOpen((prev) => !prev);
            }}>
            <PopoverTrigger asChild>
                <Button
                    variant={'primary'}
                    className={cn({
                        'bg-transparent rounded-md border uppercase shadow-outer-0 hover:bg-neutral-3 text-black dark:text-white hover:text-white truncate':
                            !buttonPrimary,
                        'w-full': buttonPrimary,
                    })}
                    data-testid='data-quality_context-selector'>
                    <span className={cn('inline-flex justify-between gap-4 items-center', { 'w-full': buttonPrimary })}>
                        <span>{selectedEnvironmentName}</span>
                        <span
                            className={cn({
                                'rotate-180 transition-transform': open,
                                'justify-self-end': buttonPrimary,
                            })}>
                            <AppIcon.CaretDown size={12} />
                        </span>
                    </span>
                </Button>
            </PopoverTrigger>
            <PopoverContent
                data-testid='data-quality_context-selector-popover'
                align='start'
                className='flex flex-col gap-2 p-4 border border-neutral-light-5 w-80'>
                <div className='flex px-0 mb-2'>
                    <TextField
                        autoFocus={true}
                        value={searchInput}
                        onChange={(e) => {
                            setSearchInput(e.target.value);
                        }}
                        variant='standard'
                        fullWidth
                        label='Search'
                        data-testid={'data-quality_context-selector-search'}
                    />
                </div>
                <ul>
                    {filteredEnvironments
                        ?.sort((a: Environment, b: Environment) => {
                            return a.name.localeCompare(b.name);
                        })
                        .map((environment: Environment, index: number) => {
                            return (
                                <li
                                    key={environment.id}
                                    className={clsx(
                                        index === filteredEnvironments.length - 1 &&
                                            'border-b border-neutral-light-5 pb-2 mb-2'
                                    )}>
                                    <Button
                                        variant={'text'}
                                        className='flex justify-between items-center gap-2 w-full'
                                        onClick={() => {
                                            onSelect({ type: environment.type, id: environment.id });
                                            handleClose();
                                        }}>
                                        <TooltipProvider>
                                            <TooltipRoot>
                                                <TooltipTrigger>
                                                    <span className='uppercase max-w-96 truncate'>
                                                        {environment.name}
                                                    </span>
                                                </TooltipTrigger>
                                                <TooltipPortal>
                                                    <TooltipContent
                                                        side='left'
                                                        className='dark:bg-neutral-dark-5 border-0'>
                                                        <span className='uppercase'>{environment.name}</span>
                                                    </TooltipContent>
                                                </TooltipPortal>
                                            </TooltipRoot>
                                        </TooltipProvider>
                                        <FontAwesomeIcon
                                            icon={environment.type === 'azure' ? faCloud : faGlobe}
                                            size='sm'
                                        />
                                    </Button>
                                </li>
                            );
                        })}
                    <li>
                        <Button
                            className='flex justify-between items-center gap-2 w-full'
                            onClick={() => {
                                onSelect({ type: 'active-directory-platform', id: null });
                                handleClose();
                            }}
                            disabled={!data?.filter((env) => env.type === 'active-directory').length}
                            variant={'text'}>
                            All Active Directory Domains
                            <FontAwesomeIcon icon={faGlobe} size='sm' />
                        </Button>
                    </li>
                    <li>
                        <Button
                            onClick={() => {
                                onSelect({ type: 'azure-platform', id: null });
                                handleClose();
                            }}
                            variant={'text'}
                            disabled={!data?.filter((env) => env.type === 'azure').length}
                            className='flex justify-between items-center gap-2 w-full'>
                            All Azure Tenants
                            <FontAwesomeIcon icon={faCloud} size='sm' />
                        </Button>
                    </li>
                </ul>
            </PopoverContent>
        </Popover>
    );
};

export default SimpleEnvironmentSelector;
