import { DataTable } from '@bloodhoundenterprise/doodleui';
import { FC, useState } from 'react';

const CertificationTable: FC = () => {
    const mockPending = '9';
    const mockData = [
        {
            id: 205,
            object_id: 'E4E6B0BB-0403-4F6A-9CC1-12138BB62220',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'DOMAIN CONTROLLERS@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:24.021781Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 304,
            object_id: '401DD3EB-3B3B-4CCF-A33C-AEB97C976B25',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'GROUPS@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:24.005204Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 310,
            object_id: '707F30C0-CCE8-4C77-9EEB-5DD0F975F636',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'SERVICEACCOUNTS@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:24.007698Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 352,
            object_id: 'D1348511-7BE5-42A1-8FC9-D5A5C4DD23A8',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'GROUPS@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:24.026395Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 356,
            object_id: '39ECBE77-D692-4D47-93D6-11925D4DB5A1',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'USERS@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:23.968067Z',
            certified_by: 'SYSTEM',
            certified: 3,
        },
        {
            id: 430,
            object_id: 'FC647E72-CD1C-4D7E-960D-454F981E34AB',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'TIER0@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:23.964493Z',
            certified_by: 'SYSTEM',
            certified: 3,
        },
        {
            id: 486,
            object_id: '51FB8637-28BC-4816-9A51-984160B207FA',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'USERS@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:23.987303Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 560,
            object_id: '79583BC0-65A5-49A3-8FCB-9EF814445913',
            environment_id: 'S-1-5-21-2697957641-2271029196-387917394',
            primary_kind: 'OU',
            name: 'TIER1@PHANTOM.CORP',
            created_at: '2025-09-03T17:11:23.987914Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 634,
            object_id: 'F2DF5C82-FE82-447F-92D4-BDA619225F76',
            environment_id: 'S-1-5-21-2845847946-3451170323-4261139666',
            primary_kind: 'OU',
            name: 'DOMAIN CONTROLLERS@GHOST.CORP',
            created_at: '2025-09-03T17:11:24.008708Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 1182,
            object_id: '16D432AC-31A9-47A8-9449-36C2BFF9417B',
            environment_id: 'S-1-5-21-3702535222-3822678775-2090119576',
            primary_kind: 'OU',
            name: 'DOMAIN CONTROLLERS@WRAITH.CORP',
            created_at: '2025-09-03T17:11:24.004605Z',
            certified_by: '',
            certified: 0,
        },
        {
            id: 1744,
            object_id: '1862497B-9BDD-4DCB-AEA8-D0D709DF5AFB',
            environment_id: 'S-1-5-21-3702535222-3822678775-2090119576',
            primary_kind: 'OU',
            name: 'MYUSERS@WRAITH.CORP',
            created_at: '2025-09-03T17:11:23.968627Z',
            certified_by: 'SYSTEM',
            certified: 3,
        },
    ];

    const [selectedRows, setSelectedRows] = useState<number[]>([]);

    const toggleRow = (id: number) => {
        setSelectedRows((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
    };

    const toggleAll = () => {
        if (selectedRows.length === mockData.length) {
            setSelectedRows([]);
        } else {
            setSelectedRows(mockData.map((row) => row.id));
        }
    };

    const allSelected = selectedRows.length === mockData.length;
    const someSelected = selectedRows.length > 0 && !allSelected;

    const columns = [
        {
            header: () => (
                <div className='pl-8'>
                    <input
                        type='checkbox'
                        checked={allSelected}
                        ref={(el) => {
                            if (el) el.indeterminate = someSelected;
                        }}
                        onChange={toggleAll}
                    />
                </div>
            ),
            id: 'bulk-certify',
            cell: ({ row }: { row: { original: (typeof mockData)[number] } }) => (
                <div className='pl-8'>
                    <input
                        type='checkbox'
                        checked={selectedRows.includes(row.original.id)}
                        onChange={() => toggleRow(row.original.id)}
                    />
                </div>
            ),
        },
        {
            header: () => <div className='pl-8 text-left'>Type</div>,
            id: 'type',
        },
        {
            header: () => <div className='pl-8 text-left'>Member Name</div>,
            id: 'name',
        },
        {
            header: () => <div className='pl-8 text-left'>Domain</div>,
            id: 'domain',
        },
        {
            header: () => <div className='pl-8 text-left'>Zone</div>,
            id: 'zone',
        },
        {
            header: () => <div className='pl-8 text-center'>First Seen</div>,
            id: 'first-seen',
        },
    ];

    type DataTableProps = React.ComponentProps<typeof DataTable>;

    const tableProps: DataTableProps['TableProps'] = {
        className: 'table-fixed',
        disableDefaultOverflowAuto: true,
    };

    const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
        className: 'sticky top-0 z-10 shadow-sm text-base',
    };

    const tableHeadProps: DataTableProps['TableHeadProps'] = {
        className: 'pr-2 text-center',
    };

    const tableCellProps: DataTableProps['TableCellProps'] = {
        className: 'truncate group relative pl-8',
    };

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        estimateSize: () => 79,
    };

    return (
        <div className='bg-neutral-light-2 dark:bg-neutral-dark-2'>
            <div className='flex items-center'>
                <h1 className='text-xl font-bold'>Certifications</h1>
                <p>{`${mockPending} pending`}</p>
            </div>
            <DataTable
                data={mockData ?? []}
                TableHeaderProps={tableHeaderProps}
                TableHeadProps={tableHeadProps}
                TableProps={tableProps}
                TableCellProps={tableCellProps}
                columns={columns}
                virtualizationOptions={virtualizationOptions}
            />
        </div>
    );
};

export default CertificationTable;
