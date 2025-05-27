import { Button } from '@bloodhoundenterprise/doodleui';
import React from 'react';

interface TableViewProps {
    open: boolean;
    onClose: () => void;
}

const TableView: React.FC<TableViewProps> = (props) => {
    const { open, onClose } = props;

    if (!open) return null;
    console.log('SHOW ME');
    return (
        <div className='absolute bottom-4 left-4 right-4 h-1/2 bg-pink-300 flex justify-center items-center'>
            {' '}
            TABLE VIEW <Button onClick={onClose}>CLOSE</Button>
        </div>
    );
};

export default TableView;

/**
 * TODO:
 * placeholder manage columns
 * trigger for manage columns
 * not required for initial PR:
 *   search container should overlap table -- not required in initial p
 *   entity panel shouldnt intersect
 */
