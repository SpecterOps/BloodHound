import { FC } from 'react';
import { Typography, Button } from '@mui/material';
import { Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import { faTriangleExclamation } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const CitrixRDPConfirmDialog: FC<{
    open: boolean;
    dialogDescription: string;
    handleCancel: () => void;
    handleConfirm: () => void;
}> = ({ open, dialogDescription, handleCancel, handleConfirm }) => {
    return (
        <Dialog
            open={open}
            maxWidth='md'
            aria-labelledby='citrix-rdp-alert-dialog-title'
            aria-describedby='citrix-rdp-alert-dialog-description'>
            <DialogTitle id='citrix-rdp-alert-dialog-title'>Confirm environment configuration</DialogTitle>
            <DialogContent sx={{ paddingBottom: 0 }}>
                <Typography
                    variant='body1'
                    component='div'
                    sx={{ display: 'flex', alignItems: 'center', paddingBottom: '16px' }}>
                    <FontAwesomeIcon icon={faTriangleExclamation} size='2x' />
                    <Typography sx={{ marginLeft: '20px' }}>{dialogDescription}</Typography>
                </Typography>
                <Typography>
                    Select <b>`Confirm`</b> to proceed and to start analysis.
                </Typography>
                <Typography>
                    Select <b>`Cancel`</b> to return to previous configuration.
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => handleCancel()}>Cancel</Button>
                <Button onClick={() => handleConfirm()}>Confirm</Button>
            </DialogActions>
        </Dialog>
    );
};

export default CitrixRDPConfirmDialog;
