import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
} from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';

const ConfirmUpdateQueryDialog: FC<{
    open: boolean;
    handleCancel: () => void;
    handleApply: () => void;
    dialogContent: string;
}> = ({ open, handleApply, handleCancel, dialogContent }) => {
    return (
        <Dialog open={open}>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle>Update Query</DialogTitle>
                    <DialogDescription>{dialogContent}</DialogDescription>
                    <DialogActions>
                        <Button variant='text' onClick={handleCancel}>
                            Cancel
                        </Button>
                        <Button variant='text' onClick={handleApply}>
                            Ok
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ConfirmUpdateQueryDialog;
