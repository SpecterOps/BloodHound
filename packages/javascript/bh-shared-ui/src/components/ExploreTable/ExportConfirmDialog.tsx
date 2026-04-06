import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogTitle,
    Switch,
    VisuallyHidden,
} from 'doodle-ui';
import { FC, useState } from 'react';

type ExportConfirmDialogProps = {
    open: boolean;
    onCancel: () => void;
    onConfirm: (columns: 'selected' | 'all') => void;
};

const ExportConfirmDialog: FC<ExportConfirmDialogProps> = ({ open, onCancel, onConfirm }) => {
    const [exportAllColumns, setExportAllColumns] = useState(false);

    return (
        <Dialog open={open}>
            <DialogContent maxWidth='sm'>
                <DialogTitle>Download Table</DialogTitle>
                <VisuallyHidden asChild>
                    <DialogDescription>
                        Download table content. You can choose to export all columns or selected columns only.
                    </DialogDescription>
                </VisuallyHidden>

                <div className='mb-3'>
                    <Switch
                        label={exportAllColumns ? 'All Columns' : 'Selected Columns'}
                        checked={exportAllColumns}
                        size='large'
                        onCheckedChange={() => setExportAllColumns((prev) => !prev)}></Switch>
                </div>
                <DialogActions>
                    <Button onClick={onCancel} variant={'secondary'}>
                        Cancel
                    </Button>
                    <Button onClick={() => onConfirm(exportAllColumns ? 'all' : 'selected')} variant={'primary'}>
                        Confirm
                    </Button>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

export default ExportConfirmDialog;
