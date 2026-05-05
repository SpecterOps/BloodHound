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
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    FormControl,
    InputLabel,
    MenuItem as MuiMenuItem,
    Select as MuiSelect,
    SelectChangeEvent,
} from '@mui/material';
import { Button, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from 'doodle-ui';
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
    const deleteEnabled = selectedQuery?.id && selectedQuery?.canEdit ? true : false;

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
                            disabled={!deleteEnabled}
                            className='ml-2'
                            variant='icon'
                            onClick={() => deleteHandler(selectedQuery?.id as number)}>
                            <FontAwesomeIcon icon={faTrash} />
                        </Button>
                    </div>
                </div>
                <div className='flex w-full items-center justify-between flex-row'>
                    <div className='w-full z-10'>
                        <label htmlFor='platforms' className='text-sm'>
                            Platforms
                        </label>
                        <Select
                            value={platform || undefined}
                            onValueChange={(val) => handlePlatformFilter(val === '__all__' ? '' : val)}>
                            <SelectTrigger id='platforms' className='w-full mt-1'>
                                <SelectValue placeholder='All' />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value='__all__'>All</SelectItem>
                                <SelectItem value='Active Directory'>Active Directory</SelectItem>
                                <SelectItem value='Azure'>Azure</SelectItem>
                                <SelectItem value='Saved Queries'>Saved Queries</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <FormControl size='small' className='w-full ml-2 z-10'>
                        <InputLabel id='category-filter-label'>Categories</InputLabel>
                        <MuiSelect
                            labelId='category-filter-label'
                            id='category-filter'
                            className='z-10'
                            value={categoryFilter}
                            label='categories'
                            open={categoriesOpen}
                            onOpen={() => setCategoriesOpen(true)}
                            onClose={() => setCategoriesOpen(false)}
                            multiple
                            onChange={handleCategoryChange}>
                            <MuiMenuItem value=''>All Categories</MuiMenuItem>
                            {categories.map((category) => (
                                <MuiMenuItem key={category} value={category}>
                                    {category}
                                </MuiMenuItem>
                            ))}
                        </MuiSelect>
                    </FormControl>
                    <div className='w-full ml-2 z-10'>
                        <label htmlFor='source-filter' className='text-sm'>
                            Source
                        </label>
                        <Select
                            value={source || undefined}
                            onValueChange={(val) => handleSourceFilter(val === '__all__' ? '' : val)}>
                            <SelectTrigger id='source-filter' className='w-full mt-1'>
                                <SelectValue placeholder='All Sources' />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value='__all__'>All Sources</SelectItem>
                                <SelectItem value='prebuilt'>Prebuilt</SelectItem>
                                <SelectItem value='personal'>Personal</SelectItem>
                                <SelectItem value='shared'>Shared</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
            </div>
            <ImportQueryDialog open={showImportDialog} onClose={() => setShowImportDialog(false)} />
        </>
    );
};

export default QuerySearchFilter;
