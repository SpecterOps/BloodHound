import { Input } from '@bloodhoundenterprise/doodleui';
import React from 'react';
import { AppIcon } from './AppIcon';

export const SearchInput: React.FC<{ value: string; onInputChange: (search: string) => void }> = ({
    value,
    onInputChange,
}) => {
    return (
        <div className='px-2 flex items-center w-1/7 border-b-2 border-neutral-dark-1 dark:border-neutral-light-1'>
            <AppIcon.MagnifyingGlass />
            <Input
                placeholder='Search'
                className='border-none bg-transparent dark:bg-transparent placeholder-neutral-dark-1 dark:placeholder-neutral-light-1 focus-visible:ring-0 focus-visible:ring-offset-0 px-2'
                onChange={(e) => onInputChange(e.target.value)}
                value={value}
            />
        </div>
    );
};
