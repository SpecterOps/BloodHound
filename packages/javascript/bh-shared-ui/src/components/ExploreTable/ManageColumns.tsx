import { Button } from '@bloodhoundenterprise/doodleui';
import React from 'react';
import { cn } from '../../utils/theme';

interface ManageColumnsProps {
    open: boolean;
    onClose: () => void;
}

const ManageColumns: React.FC<ManageColumnsProps> = (props) => {
    const { open, onClose } = props;
    return (
        <div className={cn({ hidden: !open })}>
            Manage Columns
            <Button onClick={onClose}>X</Button>
        </div>
    );
};

export default ManageColumns;
