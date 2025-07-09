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
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
} from '@bloodhoundenterprise/doodleui';
import { MenuItem } from '@mui/material';
import { FC } from 'react';
import { Link } from 'react-router-dom';

const AssetGroupMenuItem: FC<{
    assetGroupId: number;
    assetGroupName: string;
    isCurrentMember: boolean;
    removeNodePath: string;
    onAddNode?: (assetGroupId: string | number) => void;
    onShowConfirmation?: () => void;
    onCancelConfirmation?: () => void;
    showConfirmationOnAdd?: boolean;
    confirmationOnAddMessage?: string;
}> = ({
    assetGroupId,
    assetGroupName,
    isCurrentMember,
    removeNodePath,
    onAddNode = () => {},
    onShowConfirmation = () => {},
    onCancelConfirmation = () => {},
    showConfirmationOnAdd = false,
    confirmationOnAddMessage = '',
}) => {
    const handleAddNode = () => {
        onAddNode(assetGroupId);
    };

    const handleOnCancel = () => {
        onCancelConfirmation();
    };

    // selected node is not a member of the group
    if (!isCurrentMember) {
        if (showConfirmationOnAdd) {
            return (
                <>
                    <MenuItem onClick={onShowConfirmation}>Add to {assetGroupName}</MenuItem>
                    <ConfirmNodeChangesDialog
                        onCancel={handleOnCancel}
                        onAccept={handleAddNode}
                        dialogContent={confirmationOnAddMessage}
                    />
                </>
            );
        } else {
            return <MenuItem onClick={handleAddNode}>Add to {assetGroupName}</MenuItem>;
        }
    } else {
        return (
            <MenuItem component={Link} to={removeNodePath}>
                Remove from {assetGroupName}
            </MenuItem>
        );
    }
};

const ConfirmNodeChangesDialog: FC<{
    onCancel: () => void;
    onAccept: () => void;
    dialogContent: string;
}> = ({ onCancel, onAccept, dialogContent }) => {
    return (
        <DialogPortal>
            <DialogContent>
                <DialogTitle>Confirm Selection</DialogTitle>
                <DialogDescription>{dialogContent}</DialogDescription>
                <DialogActions>
                    <Button variant='tertiary' onClick={onCancel}>
                        Cancel
                    </Button>
                    <Button variant='primary' onClick={onAccept}>
                        Ok
                    </Button>
                </DialogActions>
            </DialogContent>
        </DialogPortal>
    );
};

export default AssetGroupMenuItem;
