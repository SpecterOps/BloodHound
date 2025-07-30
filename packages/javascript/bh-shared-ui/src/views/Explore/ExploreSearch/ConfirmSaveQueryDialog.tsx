import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
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
                <DialogTitle>Update Query</DialogTitle>
                <DialogContent>
                    <div>{dialogContent}</div>

                    <DialogActions>
                        <Button variant='secondary' onClick={handleCancel}>
                            Cancel
                        </Button>
                        <Button variant='primary' onClick={handleApply}>
                            Ok
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ConfirmUpdateQueryDialog;
