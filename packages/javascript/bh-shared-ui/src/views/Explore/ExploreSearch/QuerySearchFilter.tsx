import { TextField } from '@mui/material';
import { useState } from 'react';

interface QuerySearchProps {
    searchHandler: (searchTerm: string) => void;
}

const QuerySearchFilter = (props: QuerySearchProps) => {
    const { searchHandler } = props;
    const [searchTerm, setSearchTerm] = useState('');
    const handleInput = (val: string) => {
        setSearchTerm(val);
        doFuzzySearch(val);
    };

    const doFuzzySearch = (searchTerm: string) => {
        searchHandler(searchTerm);
    };

    return (
        <>
            <div>Query Search Filter</div>
            <TextField
                id='query-search'
                label='Search'
                variant='standard'
                value={searchTerm}
                onChange={(event: React.ChangeEvent<HTMLInputElement>) => handleInput(event.target.value)}
            />
        </>
    );
};

export default QuerySearchFilter;
