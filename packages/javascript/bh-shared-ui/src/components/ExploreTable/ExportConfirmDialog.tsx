import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogTitle,
    RadioGroup,
    RadioItem,
    VisuallyHidden,
} from 'doodle-ui';
import { FC, useState } from 'react';
import { ExportColumns } from './explore-table-utils';

type ExportConfirmDialogProps = {
    open: boolean;
    onCancel: () => void;
    onConfirm: (columns: 'selected' | 'all') => void;
};

const ExportConfirmDialog: FC<ExportConfirmDialogProps> = ({ open, onCancel, onConfirm }) => {
    const [exportColumns, setExportColumns] = useState<ExportColumns>('all');

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
                    <RadioGroup value={exportColumns} onValueChange={(val) => setExportColumns(val as ExportColumns)}>
                        <RadioItem value='all' label='All Columns'></RadioItem>
                        <RadioItem value='selected' label='Selected Columns'></RadioItem>
                    </RadioGroup>
                </div>
                <DialogActions>
                    <Button onClick={onCancel} variant={'secondary'}>
                        Cancel
                    </Button>
                    <Button onClick={() => onConfirm(exportColumns)} variant={'primary'}>
                        Confirm
                    </Button>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

export default ExportConfirmDialog;
