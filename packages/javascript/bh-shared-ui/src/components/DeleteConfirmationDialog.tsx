import { Typography } from '@mui/material';
import React from 'react';
import ConfirmationDialog from './ConfirmationDialog';

const DeleteConfirmationDialog: React.FC<{
    open: boolean;
    itemName: string;
    itemType: string;
    onClose: (response: boolean) => void;
    isLoading?: boolean;
    error?: string;
}> = ({ open, itemName, onClose, isLoading, error, itemType }) => {
    return (
        <ConfirmationDialog
            open={open}
            title={`Delete ${itemName}?`}
            text={
                <>
                    Continuing onwards will delete {itemName} and all associated configurations and findings.
                    <Typography className='font-bold' component={'span'}>
                        <br />
                        Warning: This change is irreversible.
                    </Typography>
                </>
            }
            challengeTxt={`Delete this ${itemType}`}
            onClose={onClose}
            error={error}
            isLoading={isLoading}
        />
    );
};

export default DeleteConfirmationDialog;
