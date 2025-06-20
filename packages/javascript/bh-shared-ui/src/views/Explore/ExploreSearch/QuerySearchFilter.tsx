import { Button } from '@bloodhoundenterprise/doodleui';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, InputLabel, TextField } from '@mui/material';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { useState } from 'react';
interface QuerySearchProps {
    queryFilterHandler: (searchTerm: string, platform: string, categories: string[]) => void;
    categories: string[];
    searchTerm: string;
    platform: string;
    categoryFilter: string[];
}

const QuerySearchFilter = (props: QuerySearchProps) => {
    const { queryFilterHandler, categories, searchTerm, platform, categoryFilter } = props;
    const [categoriesOpen, setCategoriesOpen] = useState(false);

    const handleInput = (val: string) => {
        doFuzzySearch(val);
    };

    const [age, setAge] = useState('');

    const handleChange = (event: SelectChangeEvent) => {
        setAge(event.target.value);
    };

    const doFuzzySearch = (term: string) => {
        queryFilterHandler(term, platform, categoryFilter);
    };
    const handlePlatformFilter = (val: string) => {
        queryFilterHandler(searchTerm, val, categoryFilter);
    };

    const handleCategoryChange = (event: SelectChangeEvent<typeof categoryFilter>) => {
        const {
            target: { value },
        } = event;

        // clear filter
        if (value.includes('')) {
            queryFilterHandler(searchTerm, platform, []);
            setCategoriesOpen(false);
            return;
        }

        // On autofill we get a stringified value.
        const newVal = typeof value === 'string' ? value.split(',') : value;
        queryFilterHandler(searchTerm, platform, newVal);
    };

    return (
        <>
            <Box className='mb-2'>
                <div className='mb-4 flex w-full'>
                    <div className='flex-grow'>
                        <TextField
                            id='query-search'
                            label='Search'
                            variant='standard'
                            value={searchTerm}
                            className='w-full'
                            onChange={(event: React.ChangeEvent<HTMLInputElement>) => handleInput(event.target.value)}
                        />
                    </div>

                    <div className='flex items-center ml-4'>
                        <Button variant='secondary' size='medium'>
                            Import
                        </Button>
                        <Button className='ml-2' variant='secondary' size='medium'>
                            Export
                        </Button>
                        <Button className='ml-2' variant='icon'>
                            <FontAwesomeIcon icon={faTrash} />{' '}
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
                        <Select
                            labelId='demo-simple-select-helper-label'
                            id='demo-simple-select-helper'
                            value={age}
                            onChange={handleChange}>
                            <MenuItem value=''>
                                <em>None</em>
                            </MenuItem>
                            <MenuItem value={10}>Ten</MenuItem>
                            <MenuItem value={20}>Twenty</MenuItem>
                            <MenuItem value={30}>Thirty</MenuItem>
                        </Select>
                    </FormControl>
                </div>
            </Box>
        </>
    );
};

export default QuerySearchFilter;
