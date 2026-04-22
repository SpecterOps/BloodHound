// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    RadioGroup,
    RadioItem,
    VisuallyHidden,
} from 'doodle-ui';
import { FC, useEffect, useState } from 'react';
import { ExportColumns } from './explore-table-utils';

type ExportConfirmDialogProps = {
    open: boolean;
    onCancel: () => void;
    onConfirm: (columns: 'selected' | 'all') => void;
};

const ExportConfirmDialog: FC<ExportConfirmDialogProps> = ({ open, onCancel, onConfirm }) => {
    const [exportColumns, setExportColumns] = useState<ExportColumns>('all');
    useEffect(() => {
        if (open) {
            setExportColumns('all');
        }
    }, [open]);
    return (
        <Dialog open={open} onOpenChange={() => onCancel()}>
            <DialogPortal>
                <DialogContent maxWidth='sm'>
                    <DialogTitle>Download Table</DialogTitle>
                    <VisuallyHidden asChild>
                        <DialogDescription>
                            Download table content. You can choose to export all columns or selected columns only.
                        </DialogDescription>
                    </VisuallyHidden>

                    <div className='mb-3'>
                        <RadioGroup
                            value={exportColumns}
                            onValueChange={(val) => setExportColumns(val as ExportColumns)}>
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
            </DialogPortal>
        </Dialog>
    );
};

export default ExportConfirmDialog;
