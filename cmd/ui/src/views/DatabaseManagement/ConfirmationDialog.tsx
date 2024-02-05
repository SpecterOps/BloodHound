import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
    Typography,
    useTheme,
} from '@mui/material';
import { FC } from 'react';

const ConfirmationDialog: FC<{ open: boolean; handleClose: () => void }> = ({ open, handleClose }) => {
    const theme = useTheme();

    const handleConfirm = () => {};

    return open ? (
        <Dialog maxWidth='lg' open>
            <DialogTitle>Confirm deleting data</DialogTitle>
            <DialogContent dividers>
                <Box display='flex' flexDirection='column' gap={2}>
                    <Typography variant='body1'>Continuing will delete all data from your environment.</Typography>
                    <Typography
                        variant='body1'
                        color={theme.palette.error.main}
                        fontWeight={theme.typography.fontWeightMedium}>
                        This change is irreversible.
                    </Typography>
                    <Typography variant='body1'>Please input the phrase prior to click confirm.</Typography>
                    <TextField placeholder='Please delete my data'></TextField>
                </Box>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose}>Cancel</Button>
                <Button onClick={handleConfirm}>Confirm</Button>
            </DialogActions>
        </Dialog>
    ) : null;
};

export default ConfirmationDialog;
