import { Button } from '@bloodhoundenterprise/doodleui';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, InputLabel, TextField } from '@mui/material';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { useState } from 'react';
interface QuerySearchProps {
    searchHandler: (searchTerm: string) => void;
    filterHandler: (filterValue: string) => void;
}

const QuerySearchFilter = (props: QuerySearchProps) => {
    const { searchHandler, filterHandler } = props;
    const [searchTerm, setSearchTerm] = useState('');
    const [platform, setPlatform] = useState('');
    const handleInput = (val: string) => {
        setSearchTerm(val);
        doFuzzySearch(val);
    };

    const doFuzzySearch = (searchTerm: string) => {
        searchHandler(searchTerm);
    };

    const [age, setAge] = useState('');

    const handleChange = (event: SelectChangeEvent) => {
        setAge(event.target.value);
    };
    // const handlePlatformChange = (event: SelectChangeEvent) => {
    //     setPlatform(event.target.value);
    // };

    const handleFilter = (val: string) => {
        setPlatform(val);
        doFilter(val);
    };
    const doFilter = (platform: string) => {
        filterHandler(platform);
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
                        <InputLabel id='demo-simple-select-helper-label'>Platforms</InputLabel>

                        <Select
                            labelId='demo-simple-select-helper-label'
                            id='demo-simple-select-helper'
                            value={platform}
                            label='Platforms'
                            onChange={(e) => handleFilter(e.target.value)}>
                            <MenuItem value=''>All</MenuItem>
                            <MenuItem value='Active Directory'>Active Directory</MenuItem>
                            <MenuItem value='Azure'>Azure</MenuItem>
                            <MenuItem value='Saved Queries'>Saved Queries</MenuItem>
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
