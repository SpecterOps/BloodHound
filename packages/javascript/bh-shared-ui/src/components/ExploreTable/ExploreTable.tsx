import { Button } from '@bloodhoundenterprise/doodleui';
import React from 'react';
import useToggle from '../../hooks/useToggle';
import { cn } from '../../utils/theme';
import ManageColumns from './ManageColumns';

interface ExploreTableProps {
    open: boolean;
    onClose: () => void;
}

const ExploreTable: React.FC<ExploreTableProps> = (props) => {
    const { open, onClose } = props;

    const [openManageColumns, toggleOpenManageColumns] = useToggle(false);

    if (!open) return null;

    return (
        <div className='absolute bottom-4 left-4 right-4 h-1/2 bg-pink-300 flex justify-center items-center'>
            <div className={cn({ hidden: openManageColumns })}>
                <Button onClick={toggleOpenManageColumns}>Manage Columns</Button>
                <Button onClick={onClose}>CLOSE</Button>
            </div>
            <ManageColumns open={openManageColumns} onClose={toggleOpenManageColumns} />
        </div>
    );
};

export default ExploreTable;

/**
 * TODO:
 * not required for initial PR:
 *   search container should overlap table -- not required in initial p
 *   entity panel shouldnt intersect
 */
