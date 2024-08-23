import { FC } from 'react';
import { Typography, Button, useTheme, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import { faTriangleExclamation } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

type CitrixRDPConfirmDialogProps = {
    open: boolean;
    dialogDescription: string;
    handleCancel: () => void;
    handleConfirm: () => void;
};

const CitrixRDPConfirmDialog: FC<CitrixRDPConfirmDialogProps> = ({
    open,
    dialogDescription,
    handleCancel,
    handleConfirm,
}) => {
    const theme = useTheme();

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
