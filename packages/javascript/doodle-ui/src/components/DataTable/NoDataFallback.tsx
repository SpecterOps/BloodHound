import React from 'react';
import { TableCell, TableRow } from '../Table';

interface NoDataFallbackProps {
    fallback: string | React.ReactNode;
    colSpan: number;
}

const NoDataFallback: React.FC<NoDataFallbackProps> = ({ fallback, colSpan }) => {
    if (fallback) return fallback;

    return (
        <TableRow>
            <TableCell colSpan={colSpan} className='h-24 text-center'>
                No results.
            </TableCell>
        </TableRow>
    );
};

export default NoDataFallback;
