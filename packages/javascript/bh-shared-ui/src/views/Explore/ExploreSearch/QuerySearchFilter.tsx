import { Button, Input } from '@bloodhoundenterprise/doodleui';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { InputLabel } from '@mui/material';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { useState } from 'react';
import { AppIcon } from '../../../components';
import { QueryLineItem } from '../../../types';
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
    selectedQuery: QueryLineItem | undefined;
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
        selectedQuery,
        source,
    } = props;
    const [categoriesOpen, setCategoriesOpen] = useState<boolean>(false);
    const [sourcesOpen, setSourcesOpen] = useState<boolean>(false);

    const [showImportDialog, setShowImportDialog] = useState<boolean>(false);

    const handleInput = (val: string) => {
        doFuzzySearch(val);
    };

    const doFuzzySearch = (term: string) => {
        queryFilterHandler(term, platform, categoryFilter, source);
    };
    const handlePlatformFilter = (val: string) => {
        queryFilterHandler(searchTerm, val, categoryFilter, source);
    };

    const handleCategoryChange = (event: SelectChangeEvent<typeof categoryFilter>) => {
        const {
            target: { value },
        } = event;

        // clear filter
        if (value.includes('')) {
            queryFilterHandler(searchTerm, platform, [], source);
            setCategoriesOpen(false);
            return;
        }

        // On autofill we get a stringified value.
        const newVal = typeof value === 'string' ? value.split(',') : value;
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
