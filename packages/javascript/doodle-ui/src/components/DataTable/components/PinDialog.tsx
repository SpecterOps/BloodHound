import { Button, Dialog, DialogActions, DialogContent, DialogDescription, DialogPortal, DialogTitle } from 'doodle-ui';
import React, { useCallback } from 'react';

const PinDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onConfirm: (activeId: string | number, overId: string | number) => void;
    pinDialogState: {
        action: 'pin' | 'unpin' | null;
        activeId: string | number;
        overId: string | number;
        label?: string;
    };
}> = ({ open, onClose, onConfirm, pinDialogState }) => {
    const handleClose = useCallback(() => {
        onClose();
    }, [onClose]);

    const { action, activeId, overId } = pinDialogState;

    const handleConfirm = useCallback(
        (activeId: string | number, overId: string | number) => {
            onConfirm(activeId, overId);
            onClose();
        },
        [onConfirm, onClose]
    );

    return (
        <Dialog open={open} data-testid='pin-dialog'>
            <DialogPortal>
                <DialogContent onEscapeKeyDown={handleClose}>
                    <DialogTitle>{action === 'unpin' ? 'Unpin' : 'Pin'} Column</DialogTitle>
                    {pinDialogState.label && (
                        <DialogDescription>
                            The {pinDialogState.label} column will {action === 'pin' ? 'now' : 'no longer'} be pinned.
                        </DialogDescription>
                    )}
                    <DialogActions className='flex justify-end gap-4'>
                        <Button variant='secondary' onClick={handleClose}>
                            Cancel
                        </Button>
                        <Button onClick={() => handleConfirm(activeId, overId)}>Confirm</Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default PinDialog;
