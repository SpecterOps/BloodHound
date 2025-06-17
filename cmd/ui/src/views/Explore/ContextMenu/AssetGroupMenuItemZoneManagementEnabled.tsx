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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Dialog, DialogActions, DialogContent, DialogTitle, MenuItem } from '@mui/material';
import { FC, useState } from 'react';
import { Link } from 'react-router-dom';

const AssetGroupMenuItem: FC<{
    assetGroupId: number;
    assetGroupName: string;
    isCurrentMember: boolean;
    onAddNode?: (assetGroupId: string | number) => void;
    showConfirmationOnAdd?: boolean;
    confirmationOnAddMessage?: string;
}> = ({
    assetGroupId,
    assetGroupName,
    isCurrentMember,
    onAddNode = () => {},
    showConfirmationOnAdd = false,
    confirmationOnAddMessage = '',
}) => {
    const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);

    const handleAddNode = () => {
        onAddNode(assetGroupId);
        setConfirmDialogOpen(false);
    };

    const handleOnCancel = () => {
        setConfirmDialogOpen(false);
    };

    // selected node is not a member of the group
    if (!isCurrentMember) {
        return (
            <>
                <MenuItem onClick={showConfirmationOnAdd ? () => setConfirmDialogOpen(true) : handleAddNode}>
                    Add to {assetGroupName}
                </MenuItem>
                {showConfirmationOnAdd && (
                    <ConfirmNodeChangesDialog
                        onCancel={handleOnCancel}
                        onAccept={handleAddNode}
                        open={confirmDialogOpen}
                        dialogContent={confirmationOnAddMessage}
                    />
                )}
            </>
        );
    } else {
        return (
            <MenuItem component={Link} to={`/tier-management/details/tag/${assetGroupId}`}>
                Remove from {assetGroupName}
            </MenuItem>
        );
    }
};

const ConfirmNodeChangesDialog: FC<{
    open: boolean;
    onCancel: () => void;
    onAccept: () => void;
    dialogContent: string;
}> = ({ open, onCancel, onAccept, dialogContent }) => {
    return (
        <Dialog open={open}>
            <DialogTitle>Confirm Selection</DialogTitle>
            <DialogContent>{dialogContent}</DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={onCancel}>
                    Cancel
                </Button>
                <Button variant='primary' onClick={onAccept}>
                    Ok
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default AssetGroupMenuItem;
