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
    Button,
    IconButton,
    Input,
    Label,
    Menu,
    MenuContent,
    MenuItem,
    MenuTrigger,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from 'doodle-ui';
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

    const [showImportDialog, setShowImportDialog] = useState<boolean>(false);

    const handleInput = (val: string) => {
        queryFilterHandler(val, platform, categoryFilter, source);
    };

    const handlePlatformFilter = (val: string) => {
        queryFilterHandler(searchTerm, val, categoryFilter, source);
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
                            variant='outlined'
                            placeholder='Search'
                            value={searchTerm}
                            onChange={(event: React.ChangeEvent<HTMLInputElement>) => handleInput(event.target.value)}
                        />
                        <AppIcon.MagnifyingGlass
                            size={16}
                            className='absolute right-2 top-[50%] -mt-[8px] pointer-events-none'
                        />
                    </div>
                    <div className='flex items-center ml-4 gap-2'>
                        <Button variant='secondary' onClick={importHandler}>
                            Import
                        </Button>
                        <Button disabled={!exportEnabled} variant='secondary' onClick={exportHandler}>
                            Export
                        </Button>
                        <IconButton
                            aria-label='delete'
                            disabled={!deleteEnabled}
                            onClick={() => deleteHandler(selectedQuery?.id as number)}>
                            <FontAwesomeIcon icon={faTrash} />
                        </IconButton>
                    </div>
                </div>
                <div className='grid w-full grid-cols-3 items-center gap-2'>
                    <div className='min-w-0'>
                        <Label htmlFor='platform-filter'>Platforms</Label>
                        <Select
                            value={platform || '__all__'}
                            onValueChange={(value) => handlePlatformFilter(value === '__all__' ? '' : value)}>
                            <SelectTrigger id='platform-filter'>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value='__all__'>All</SelectItem>
                                <SelectItem value='Active Directory'>Active Directory</SelectItem>
                                <SelectItem value='Azure'>Azure</SelectItem>
                                <SelectItem value='Saved Queries'>Saved Queries</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <div className='min-w-0'>
                        <Label htmlFor='category-filter'>Categories</Label>
                        <Menu>
                            <MenuTrigger asChild>
                                <Button id='category-filter' variant='secondary' className='w-full justify-between'>
                                    <span
                                        className='truncate'
                                        title={categoryFilter.length ? categoryFilter.join(', ') : 'All Categories'}>
                                        {categoryFilter.length ? categoryFilter.join(', ') : 'All Categories'}
                                    </span>
                                </Button>
                            </MenuTrigger>
                            <MenuContent
                                className='max-h-[var(--radix-dropdown-menu-content-available-height)] overflow-y-auto'
                                aria-label='Categories'>
                                <MenuItem onSelect={() => queryFilterHandler(searchTerm, platform, [], source)}>
                                    All Categories
                                </MenuItem>
                                {categories.map((category) => (
                                    <MenuItem
                                        key={category}
                                        role='menuitemcheckbox'
                                        aria-checked={categoryFilter.includes(category)}
                                        onSelect={(event) => {
                                            event.preventDefault();
                                            const values = categoryFilter.includes(category)
                                                ? categoryFilter.filter((value) => value !== category)
                                                : [...categoryFilter, category];
                                            queryFilterHandler(searchTerm, platform, values, source);
                                        }}>
                                        {categoryFilter.includes(category) ? '✓ ' : ''}
                                        {category}
                                    </MenuItem>
                                ))}
                            </MenuContent>
                        </Menu>
                    </div>
                    <div className='min-w-0'>
                        <Label htmlFor='source-filter'>Source</Label>
                        <Select
                            value={source || '__all__'}
                            onValueChange={(value) => handleSourceFilter(value === '__all__' ? '' : value)}>
                            <SelectTrigger id='source-filter'>
                                <SelectValue />
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
