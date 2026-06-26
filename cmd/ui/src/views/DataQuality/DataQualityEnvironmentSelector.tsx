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
    cn,
    DropdownTrigger,
    getOpenGraphEnvironmentInfoMap,
    optionIconStyles,
    optionStyles,
    popoverContentStyles,
} from 'bh-shared-ui';
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
import { DataQualityEnvironment } from 'js-client-library';
import React, { ReactNode, useMemo, useState } from 'react';

export type DataQualitySelection = {
    type: string;
    id: string | null;
    environmentKind: string;
    sourceKind: string;
    selectionType: 'environment' | 'aggregate';
};

const selectionKey = (environment: DataQualityEnvironment) => `${environment.type}:${environment.environment_kind}`;

const aggregateLabel = (
    environment: Pick<DataQualityEnvironment, 'environment_kind' | 'type'>,
    environmentInfo: ReturnType<typeof getOpenGraphEnvironmentInfoMap>
) => environmentInfo[environment.type]?.aggregationDisplayName || `All ${environment.environment_kind} Environments`;

const selectedText = (selected: DataQualitySelection | null, environments: DataQualityEnvironment[]): string => {
    if (!selected) return 'Select Environment';

    const environmentInfo = getOpenGraphEnvironmentInfoMap(environments);
    if (selected.selectionType === 'aggregate') {
        const selectedEnvironment = environments.find(
            (environment) =>
                environment.type === selected.type && environment.environment_kind === selected.environmentKind
        );

        return selectedEnvironment
            ? aggregateLabel(selectedEnvironment, environmentInfo)
            : environmentInfo[selected.type]?.aggregationDisplayName || `All ${selected.environmentKind} Environments`;
    }

    return environments.find((environment) => environment.id === selected.id)?.name || 'Select Environment';
};

const searchMatchesEnvironment = (environment: DataQualityEnvironment, search: string) => {
    const normalizedSearch = search.trim().toLowerCase();
    if (!normalizedSearch) return true;

    return [environment.name, environment.id, environment.environment_kind, environment.source_kind].some((value) =>
        value.toLowerCase().includes(normalizedSearch)
    );
};

const DataQualityEnvironmentSelector: React.FC<{
    align?: 'center' | 'start' | 'end';
    environments?: DataQualityEnvironment[];
    errorMessage?: ReactNode;
    isError?: boolean;
    isLoading?: boolean;
    onSelect: (selection: DataQualitySelection) => void;
    selected: DataQualitySelection | null;
    variant?: ButtonProps['variant'];
}> = ({
    align = 'start',
    environments = [],
    errorMessage = '',
    isError = false,
    isLoading = false,
    onSelect,
    selected,
    variant,
}) => {
    const [open, setOpen] = useState<boolean>(false);
    const [search, setSearch] = useState('');

    const collectedEnvironments = useMemo(
        () =>
            environments
                .filter((environment) => environment.collected)
                .sort((first, second) => first.name.localeCompare(second.name)),
        [environments]
    );

    const aggregateEnvironments = useMemo(() => {
        const environmentByKey = new Map<string, DataQualityEnvironment>();
        for (const environment of collectedEnvironments) {
            if (!environmentByKey.has(selectionKey(environment))) {
                environmentByKey.set(selectionKey(environment), environment);
            }
        }
        return [...environmentByKey.values()].sort((first, second) =>
            first.environment_kind.localeCompare(second.environment_kind)
        );
    }, [collectedEnvironments]);

    const filteredEnvironments = useMemo(
        () => collectedEnvironments.filter((environment) => searchMatchesEnvironment(environment, search)),
        [collectedEnvironments, search]
    );

    const environmentInfo = getOpenGraphEnvironmentInfoMap(environments);
    const selectedEnvironmentName = selectedText(selected, environments);

    const handleOpenChange = (nextOpen: boolean) => {
        setOpen(nextOpen);
        if (!nextOpen) {
            setSearch('');
        }
    };

    const handleClose = () => handleOpenChange(false);

    const handlePlatformClick = (environment: DataQualityEnvironment) => {
        onSelect({
            type: environment.type,
            id: null,
            environmentKind: environment.environment_kind,
            sourceKind: environment.source_kind,
            selectionType: 'aggregate',
        });
        handleClose();
    };

    const handleEnvironmentClick = (environment: DataQualityEnvironment) => {
        onSelect({
            type: environment.type,
            id: environment.id,
            environmentKind: environment.environment_kind,
            sourceKind: environment.source_kind,
            selectionType: 'environment',
        });
        handleClose();
    };

    if (isLoading) return <Skeleton className='rounded-md w-10' />;

    if (isError) return <Alert severity='error'>{errorMessage}</Alert>;

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
                className={cn(popoverContentStyles, 'gap-2 p-2 w-96')}>
                <TextField
                    autoFocus
                    fullWidth
                    label='Search'
                    onChange={(event) => setSearch(event.target.value)}
                    value={search}
                    variant='standard'
                    inputProps={{ 'aria-label': 'Search environments' }}
                    sx={{ px: 2, pb: 2 }}
                />
                <ul className='border-b border-neutral-light-5 pb-2 mb-2'>
                    {aggregateEnvironments.map((environment) => (
                        <li key={`${selectionKey(environment)}-aggregate`}>
                            <Button
                                className={cn(optionStyles, 'flex justify-between items-center gap-2')}
                                onClick={() => handlePlatformClick(environment)}
                                variant='text'>
                                {aggregateLabel(environment, environmentInfo)}
                                <FontAwesomeIcon
                                    className={optionIconStyles}
                                    icon={environmentInfo[environment.type]?.icon}
                                    size='sm'
                                />
                            </Button>
                        </li>
                    ))}
                </ul>
                <div className='max-h-80 overflow-y-auto'>
                    <ul>
                        {filteredEnvironments.map((environment) => (
                            <li key={`${selectionKey(environment)}:${environment.id}`}>
                                <Button
                                    className={cn(optionStyles, 'flex justify-between items-center gap-2')}
                                    onClick={() => handleEnvironmentClick(environment)}
                                    variant='text'>
                                    <TooltipProvider>
                                        <TooltipRoot>
                                            <TooltipTrigger>
                                                <span className='uppercase max-w-96 truncate'>{environment.name}</span>
                                            </TooltipTrigger>
                                            <TooltipPortal>
                                                <TooltipContent side='left' className='dark:bg-neutral-dark-5 border-0'>
                                                    <span className='uppercase'>{environment.id}</span>
                                                </TooltipContent>
                                            </TooltipPortal>
                                        </TooltipRoot>
                                    </TooltipProvider>
                                    <FontAwesomeIcon
                                        className={optionIconStyles}
                                        icon={environmentInfo[environment.type]?.icon}
                                        size='sm'
                                    />
                                </Button>
                            </li>
                        ))}
                    </ul>
                </div>
            </PopoverContent>
        </Popover>
    );
};

export default DataQualityEnvironmentSelector;
