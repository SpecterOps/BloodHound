import { FC } from 'react';
import { Box, Typography } from '@mui/material';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import { Button } from '@bloodhoundenterprise/doodleui';
import { faTriangleExclamation } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const CitrixRDPConfirmDialog: FC<{
    open: boolean;
    handleCancel: () => void;
    handleConfirm: () => void;
    isLoading: boolean;
}> = ({ open, handleCancel, handleConfirm, isLoading }) => {
    return (
        <Dialog
            open={open}
            maxWidth='md'
            aria-labelledby='citrix-rdp-alert-dialog-title'
            aria-describedby='citrix-rdp-alert-dialog-description'>
            <DialogTitle id='citrix-rdp-alert-dialog-title'>Confirm environment configuration</DialogTitle>
            <DialogContent sx={{ paddingBottom: 0 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', paddingBottom: '16px' }}>
                    <FontAwesomeIcon icon={faTriangleExclamation} size='2x' />
                    <Typography sx={{ marginLeft: '20px' }}>
                        Analysis has been added with Citrix Configuration, this will ensure that BloodHound can account
                        for Direct Access RDP connections. Compensating controls handled within Citrix are not handled
                        by BloodHound at this time.
                    </Typography>
                </Box>
                <Typography>
                    Select <b>`Confirm`</b> to proceed and to start analysis.
                </Typography>
                <Typography>
                    Select <b>`Cancel`</b> to return to previous configuration.
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={() => handleCancel()} disabled={isLoading}>
                    Cancel
                </Button>
                <Button onClick={() => handleConfirm()} disabled={isLoading}>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default CitrixRDPConfirmDialog;
