import { Button } from '@bloodhoundenterprise/doodleui';
import React from 'react';

interface ExploreTableProps {
    open: boolean;
    onClose: () => void;
}

const ExploreTable: React.FC<ExploreTableProps> = (props) => {
    const { open, onClose } = props;

    if (!open) return null;

    return (
        <div className='absolute bottom-4 left-4 right-4 h-1/2 bg-pink-300 flex justify-center items-center'>
            {' '}
            TABLE VIEW <Button onClick={onClose}>CLOSE</Button>
        </div>
    );
};

export default ExploreTable;

/**
 * TODO:
 * placeholder manage columns
 * trigger for manage columns
 * not required for initial PR:
 *   search container should overlap table -- not required in initial p
 *   entity panel shouldnt intersect
 */
