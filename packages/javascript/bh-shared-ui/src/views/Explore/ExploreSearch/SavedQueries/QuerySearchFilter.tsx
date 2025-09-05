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
import { Button, Input } from '@bloodhoundenterprise/doodleui';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FormControl, InputLabel, MenuItem, Select, SelectChangeEvent } from '@mui/material';
import { useState } from 'react';
import { AppIcon } from '../../../../components';
import { useSavedQueriesContext } from '../../providers';
import ImportQueryDialog from './ImportQueryDialog';
interface QuerySearchProps {
    queryFilterHandler: (searchTerm: string, platform: string, categories: string[], source: string) => void;
    exportHandler: () => void;
    deleteHandler: (id: number) => void;
    categories: string[];
    searchTerm: string;
    platform: string;
    categoryFilter: string[];
    source: string;
}

const QuerySearchFilter = (props: QuerySearchProps) => {
    const {
        queryFilterHandler,
        exportHandler,
        deleteHandler,
        categories,
        searchTerm,
        platform,
        categoryFilter,
        source,
    } = props;
    const { selectedQuery } = useSavedQueriesContext();

    const [categoriesOpen, setCategoriesOpen] = useState<boolean>(false);
    const [sourcesOpen, setSourcesOpen] = useState<boolean>(false);

    const [showImportDialog, setShowImportDialog] = useState<boolean>(false);

    const handleInput = (val: string) => {
        queryFilterHandler(val, platform, categoryFilter, source);
    };

    const handlePlatformFilter = (val: string) => {
        queryFilterHandler(searchTerm, val, categoryFilter, source);
    };

    const handleCategoryChange = (event: SelectChangeEvent<typeof categoryFilter>) => {
        const raw = event.target.value;
        const newVal = typeof raw === 'string' ? raw.split(',') : raw;
        if (newVal.includes('')) {
            queryFilterHandler(searchTerm, platform, [], source);
            setCategoriesOpen(false);
            return;
        }
        queryFilterHandler(searchTerm, platform, newVal, source);
    };

    const handleSourceFilter = (val: string) => {
        queryFilterHandler(searchTerm, platform, categoryFilter, val);
    };

    const exportEnabled = selectedQuery?.id ? true : false;

    const importHandler = () => {
        setShowImportDialog(true);
    };

    return (
        <>
            <div className='mb-2'>
                <div className='mb-4 flex w-full'>
                    <div className='flex-grow relative'>
                        <Input
                            type='text'
                            id='query-search'
                            placeholder='Search'
                            value={searchTerm}
                            className='w-full bg-transparent dark:bg-transparent rounded-none border-neutral-dark-5 border-t-0 border-x-0'
                            onChange={(event: React.ChangeEvent<HTMLInputElement>) => handleInput(event.target.value)}
                        />
                        <AppIcon.MagnifyingGlass size={16} className='absolute right-2 top-[50%] -mt-[8px]' />
                    </div>
                    <div className='flex items-center ml-4'>
                        <Button variant='secondary' size='medium' onClick={importHandler}>
                            Import
                        </Button>
                        <Button
                            disabled={!exportEnabled}
                            className='ml-2'
                            variant='secondary'
                            size='medium'
                            onClick={exportHandler}>
                            Export
                        </Button>
                        <Button
                            aria-label='delete'
                            disabled={!exportEnabled}
                            className='ml-2'
                            variant='icon'
                            onClick={() => deleteHandler(selectedQuery?.id as number)}>
                            <FontAwesomeIcon icon={faTrash} />
                        </Button>
                    </div>
                </div>
                <div className='flex w-full items-center justify-between flex-row'>
                    <FormControl size='small' className='w-full'>
                        <InputLabel id='platforms-label'>Platforms</InputLabel>
                        <Select
                            labelId='platforms-label'
                            id='demo-simple-select-helper'
                            value={platform}
                            label='Platforms'
                            onChange={(e) => handlePlatformFilter(e.target.value)}>
                            <MenuItem value=''>All</MenuItem>
                            <MenuItem value='Active Directory'>Active Directory</MenuItem>
                            <MenuItem value='Azure'>Azure</MenuItem>
                            <MenuItem value='Saved Queries'>Saved Queries</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size='small' className='w-full ml-2'>
                        <InputLabel id='category-filter-label'>Categories</InputLabel>
                        <Select
                            labelId='category-filter-label'
                            id='category-filter'
                            value={categoryFilter}
                            label='categories'
                            open={categoriesOpen}
                            onOpen={() => setCategoriesOpen(true)}
                            onClose={() => setCategoriesOpen(false)}
                            multiple
                            onChange={handleCategoryChange}>
                            <MenuItem value=''>All Categories</MenuItem>
                            {categories.map((category) => (
                                <MenuItem key={category} value={category}>
                                    {category}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                    <FormControl size='small' className='w-full ml-2'>
                        <InputLabel id='source-filter-label'>Source</InputLabel>
                        <Select
                            labelId='source-filter-label'
                            id='source-filter'
                            value={source || ''}
                            label='source'
                            open={sourcesOpen}
                            onOpen={() => setSourcesOpen(true)}
                            onClose={() => setSourcesOpen(false)}
                            onChange={(e) => handleSourceFilter(e.target.value)}>
                            <MenuItem value=''>All Sources</MenuItem>
                            <MenuItem value='prebuilt'>Prebuilt</MenuItem>
                            <MenuItem value='owned'>Owned</MenuItem>
                            <MenuItem value='shared'>Shared</MenuItem>
                        </Select>
                    </FormControl>
                </div>
            </div>
            <ImportQueryDialog open={showImportDialog} onClose={() => setShowImportDialog(false)} />
        </>
    );
};

export default QuerySearchFilter;
