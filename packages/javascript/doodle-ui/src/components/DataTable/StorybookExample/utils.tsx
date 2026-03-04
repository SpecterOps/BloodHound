import { faker } from '@faker-js/faker';
import { faCaretDown, faCaretUp, faEllipsis, faGem, faUser, faUserGroup } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ColumnDef } from '@tanstack/react-table';

export const getData = (entries: number): Finding[] =>
    Array(entries)
        .fill(undefined)
        .map(() => ({
            id: faker.animal.petName(),
            nonTierZeroPrincipal: `${faker.person.firstName()} ${faker.person.lastName()}`,
            tierZeroPrincipal: `${faker.person.firstName()} ${faker.person.lastName()}`,
            firstSeen: faker.date.month(),
            email: faker.internet.email(),
            userName: faker.internet.username(),
            buzzAdjective: faker.company.buzzAdjective(),
            buzzNoun: faker.company.buzzNoun(),
            buzzVerb: faker.company.buzzVerb(),
            nullValue: null,
        }));

interface Finding {
    id: string;
    nonTierZeroPrincipal: string;
    tierZeroPrincipal: string;
    firstSeen: string;
}

export const getColumns = (sortOrder?: string | null, handleSort?: (sortBy: string) => void) => {
    const columns: ColumnDef<Finding>[] = [
        {
            accessorKey: '',
            id: 'action-menu',
            size: 50,
            minSize: 50,
            cell: () => (
                <button className='pl-4'>
                    <FontAwesomeIcon icon={faEllipsis} className='rotate-90 dark:text-neutral-light-1' />
                </button>
            ),
        },
        {
            accessorKey: 'nonTierZeroPrincipal',
            size: 250,
            minSize: 50,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Non Tier Zero Principal</span>;
            },
            cell: ({ row }) => {
                return (
                    <div className='flex justify-start items-center max-w-2 dark:text-neutral-light-1 text-nowrap text-black'>
                        <span className='p-2 border border-black rounded-full bg-[#17E625] mr-2 size-4 flex justify-center items-center'>
                            <FontAwesomeIcon icon={faUser} className='text-xs' />
                        </span>
                        {row.getValue('nonTierZeroPrincipal')}
                    </div>
                );
            },
            meta: {
                label: 'Non Tier Zero Principal',
            },
        },
        {
            size: 220,
            minSize: 120,
            accessorKey: 'tierZeroPrincipal',
            header: () => (
                <div className='flex justify-start items-center dark:text-neutral-light-1 text-nowrap'>
                    <span className='p-2 rounded-full bg-black mr-2 size-4 flex justify-center items-center'>
                        <FontAwesomeIcon icon={faGem} className='text-xs' color='white' />
                    </span>
                    Tier Zero Principal
                    {sortOrder !== undefined ? (
                        <button onClick={() => handleSort?.('tierZeroPrincipal')} className='p-2'>
                            {!sortOrder && <FontAwesomeIcon icon={faCaretDown} color='black' />}
                            {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} color='dodgerblue' />}
                            {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} color='dodgerblue' />}
                        </button>
                    ) : null}
                </div>
            ),
            cell: ({ row }) => {
                return (
                    <div className='flex justify-start items-center dark:text-neutral-light-1 text-black'>
                        <span className='p-2 border border-black rounded-full bg-[#DBE617] mr-2 size-4 flex justify-center items-center'>
                            <FontAwesomeIcon icon={faUserGroup} className='text-xs' />
                        </span>
                        {row.getValue('tierZeroPrincipal')}
                    </div>
                );
            },
            meta: {
                label: 'Tier Zero Principal',
            },
        },
        {
            accessorKey: 'email',
            size: 220,
            minSize: 120,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Email</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('email')}</span>;
            },
        },
        {
            accessorKey: 'userName',
            size: 350,
            minSize: 50,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Username</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('userName')}</span>;
            },
            meta: {
                label: 'User Name',
            },
        },
        {
            accessorKey: 'firstSeen',
            size: 350,
            minSize: 50,
            enableResizing: false,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Non-Resizable Col</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('firstSeen')}</span>;
            },
            meta: {
                label: 'First Seen',
            },
        },

        {
            accessorKey: 'buzzAdjective',
            size: 350,
            minSize: 50,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Buzz Adjective</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('buzzAdjective')}</span>;
            },
            meta: {
                label: 'Buzz Adjective',
            },
        },
        {
            accessorKey: 'buzzNoun',
            size: 350,
            minSize: 50,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Buzz Noun</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('buzzNoun')}</span>;
            },
            meta: {
                label: 'Buzz Noun',
            },
        },
        {
            accessorKey: 'buzzVerb',
            size: 350,
            minSize: 50,
            header: () => {
                return <span className='dark:text-neutral-light-1'>Buzz Verb</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('buzzVerb')}</span>;
            },
            meta: {
                label: 'Buzz Verb',
            },
        },
    ];
    return columns;
};
