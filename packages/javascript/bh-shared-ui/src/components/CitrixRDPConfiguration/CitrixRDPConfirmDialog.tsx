import { FC } from 'react';
import { Typography, Button, useTheme, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

type CitrixRDPConfirmDialogProps = {
    open: boolean;
    futureSwitchState: boolean;
    handleCancel: () => void;
    handleConfirm: () => void;
};
export const dialogTitle = 'Confirm environment configuration';
const enabledDialogDescription =
    'Analysis has been added with Citrix Configuration, this will ensure that BloodHound can account for Direct Access RDP connections. \n\nCompensating controls handled within Citrix are not handled by BloodHound at this time.';
const disabledDialogDescription =
    'Analysis has been removed with Citrix Configuration, this will result in BloodHound performing analysis to account for this change';

const CitrixRDPConfirmDialog: FC<CitrixRDPConfirmDialogProps> = ({
    open,
    futureSwitchState,
    handleCancel,
    handleConfirm,
}) => {
    const theme = useTheme();

    return (
        <Dialog
            open={open}
            maxWidth='sm'
            aria-labelledby='citrix-rdp-alert-dialog-title'
            aria-describedby='citrix-rdp-alert-dialog-description'>
            <DialogTitle id='citrix-rdp-alert-dialog-title' sx={{ fontSize: '20px' }}>
                {dialogTitle}
            </DialogTitle>
            <DialogContent sx={{ paddingBottom: 0 }}>
                <Typography variant='body2' sx={{ paddingBottom: '16px', whiteSpace: 'break-spaces' }}>
                    {futureSwitchState ? enabledDialogDescription : disabledDialogDescription}
                </Typography>
                <Typography variant='body2'>
                    Select <b>`Confirm`</b> to proceed and to start analysis.
                </Typography>
                <Typography variant='body2'>
                    Select <b>`Cancel`</b> to return to previous configuration.
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button sx={{ color: theme.palette.color.primary }} onClick={() => handleCancel()}>
                    Cancel
                </Button>
                <Button sx={{ color: theme.palette.primary.main }} onClick={() => handleConfirm()}>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default CitrixRDPConfirmDialog;
