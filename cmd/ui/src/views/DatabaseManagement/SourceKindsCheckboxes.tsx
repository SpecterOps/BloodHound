import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { Checkbox, FormControlLabel } from '@mui/material';
import { type FC } from 'react';
import { useQuery } from 'react-query';

// TODO: This goes away once the API is in place
const STUB_SOURCE_KINDS = [
    {
        id: 1,
        name: 'Base',
    },
    {
        id: 2,
        name: 'AZBase',
    },
    {
        id: 3,
        name: 'ACustomBase',
    },
    {
        id: 0,
        name: 'Sourceless',
    },
];

// The default source kind names are replaced with friendlier ones
const KIND_LABEL_MAP: Record<string, string> = {
    Base: 'Active Directory',
    AZBase: 'Azure',
};

// Displayed while source kinds are loading
const LOADING_CHECKBOXES = (
    <>
        <div className='pl-5 flex items-center'>
            <Checkbox />
            <Skeleton className='h-4 w-[200px]' />
        </div>
        <div className='pl-5 flex items-center'>
            <Checkbox />
            <Skeleton className='h-4 w-[200px]' />
        </div>
        <div className='pl-5 flex items-center'>
            <Checkbox />
            <Skeleton className='h-4 w-[200px]' />
        </div>
    </>
);

export const SourceKindsCheckboxes: FC<{
    checked: number[];
    disabled: boolean;
    onChange: (checked: number[]) => void;
}> = ({ checked, disabled, onChange }) => {
    const {
        data: sourceKinds,
        isLoading,
        isSuccess,
    } = useQuery({
        queryKey: ['source-kinds'],
        // TODO: Use the API once it's available
        // queryFn: ({ signal }) => apiClient.getSourceKinds({ signal }).then((res) => res.data.data.kinds),
        queryFn: () => STUB_SOURCE_KINDS,
    });

    let amountChecked = 'none';

    if (isSuccess && checked.length > 0) {
        amountChecked = sourceKinds.length === checked.length ? 'all' : 'some';
    }

    const toggleAllChecked = () => {
        if (sourceKinds) {
            onChange(['none', 'some'].includes(amountChecked) ? sourceKinds.map((item) => item.id) : []);
        }
    };

    const toggleSourceKind = (id: number) => () => {
        const newChecked = checked.includes(id) ? checked.filter((item) => item !== id) : [...checked, id];
        onChange(newChecked);
    };

    return (
        <>
            <FormControlLabel
                label='All graph data'
                control={
                    <Checkbox
                        checked={amountChecked === 'all'}
                        disabled={disabled}
                        indeterminate={amountChecked === 'some'}
                        name='All GraphData'
                        onChange={toggleAllChecked}
                    />
                }
            />

            {isLoading && LOADING_CHECKBOXES}

            {isSuccess &&
                sourceKinds.map((item) => (
                    <FormControlLabel
                        className='pl-8'
                        control={
                            <Checkbox
                                checked={checked.includes(item.id)}
                                onChange={toggleSourceKind(item.id)}
                                name={item.name}
                                disabled={disabled}
                            />
                        }
                        label={(KIND_LABEL_MAP[item.name] ?? item.name) + ' data'}
                        key={item.id}
                    />
                ))}
        </>
    );
};
