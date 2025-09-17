import { Button } from '@bloodhoundenterprise/doodleui';
import { Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material';
import { FC } from 'react';

interface ConfirmDeleteQueryDialogProps {
    open: boolean;
    queryId?: number;
    handleClose: () => void;
    deleteHandler?: (id: number) => void;
}

const ConfirmDeleteQueryDialog: FC<ConfirmDeleteQueryDialogProps> = ({ open, queryId, handleClose, deleteHandler }) => {
    return (
        <Dialog open={open} onClose={handleClose} maxWidth={'xs'} fullWidth>
            <DialogTitle>Delete Query</DialogTitle>
            <DialogContent>
                <DialogContentText>Are you sure you want to delete this query?</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={handleClose}>
                    Cancel
                </Button>
                <Button
                    onClick={() => {
                        if (deleteHandler) deleteHandler(queryId!);
                        handleClose();
                    }}
                    color='primary'
                    autoFocus>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ConfirmDeleteQueryDialog;
