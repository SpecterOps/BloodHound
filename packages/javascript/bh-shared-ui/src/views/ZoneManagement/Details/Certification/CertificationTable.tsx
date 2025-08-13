import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';

const CertificationTable: FC = () => {
    const mockPending = '9';
    const mockData = [
        {
            severity: '1',
            id: '1',
            type: 'User',
            memberName: 'Mock Member',
            domain: 'Mock Domain',
            zone: 'Tier 0',
            firstSeen: '8/15/25',
        },
        {
            severity: '1',
            id: '2',
            type: 'User',
            memberName: 'Mock Member 2',
            domain: 'Mock Domain 2',
            zone: 'Tier 0',
            firstSeen: '8/15/25',
        },
        {
            severity: '1',
            id: '3',
            type: 'User',
            memberName: 'Mock Member 3',
            domain: 'Mock Domain 3',
            zone: 'Tier 0',
            firstSeen: '8/15/25',
        },
        {
            severity: '1',
            id: '4',
            type: 'User',
            memberName: 'Mock Member 4',
            domain: 'Mock Domain 4',
            zone: 'Tier 0',
            firstSeen: '8/15/25',
        },
    ];

    return (
        <div className='bg-neutral-light-2 dark:bg-neutral-dark-2'>
            <div className='flex items-center'>
                <h1 className='text-xl font-bold'>Certifications</h1>
                <p>{`${mockPending} pending`}</p>
            </div>
            <Table className='w-full'>
                <TableHeader>
                    <TableRow>
                        <TableHead className='w-32'>Severity</TableHead>
                        <TableHead className='w-32'>Type</TableHead>
                        <TableHead className='w-48'>Member Name</TableHead>
                        <TableHead className='w-32'>Domain</TableHead>
                        <TableHead className='w-32'>Zone</TableHead>
                        <TableHead className='w-48'>First Seen</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {mockData.map((data) => (
                        <TableRow className='' key={data.id}>
                            <TableCell className=''>{data.severity}</TableCell>
                            <TableCell className=''>{data.type}</TableCell>
                            <TableCell className=''>{data.memberName}</TableCell>
                            <TableCell className=''>{data.domain}</TableCell>
                            <TableCell className=''>{data.zone}</TableCell>
                            <TableCell className=''>{data.firstSeen}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};

export default CertificationTable;
