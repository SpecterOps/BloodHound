// Copyright 2025 Specter Ops, Inc.
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
    DialogTitle,
} from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';

type AnalyzeNowConfirmDialogProps = {
    open: boolean;
    onCancel: () => void;
    onConfirm: () => void;
};

const AnalyzeNowConfirmDialog: FC<AnalyzeNowConfirmDialogProps> = ({ open, onCancel, onConfirm }) => {
    return (
        <Dialog open={open}>
            <DialogContent maxWidth='sm'>
                <DialogTitle>Confirm re-run analysis</DialogTitle>
                <DialogDescription>
                    Analysis may take some time, during which your data will be in flux. Proceed with analysis?
                </DialogDescription>
                <DialogActions>
                    <Button onClick={onCancel} variant={'secondary'}>
                        Cancel
                    </Button>
                    <Button onClick={onConfirm} variant={'primary'}>
                        Confirm
                    </Button>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

export default AnalyzeNowConfirmDialog;
