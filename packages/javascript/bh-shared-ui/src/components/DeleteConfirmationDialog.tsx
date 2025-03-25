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

import { Typography } from '@mui/material';
import React from 'react';
import ConfirmationDialog from './ConfirmationDialog';

const DeleteConfirmationDialog: React.FC<{
    open: boolean;
    itemName: string;
    itemType: string;
    onClose: (response: boolean) => void;
    isLoading?: boolean;
    error?: string;
}> = ({ open, itemName, onClose, isLoading, error, itemType }) => {
    return (
        <ConfirmationDialog
            open={open}
            title={`Delete ${itemName}?`}
            text={
                <>
                    Continuing onwards will delete {itemName} and all associated configurations and findings.
                    <Typography className='font-bold' component={'span'}>
                        <br />
                        Warning: This change is irreversible.
                    </Typography>
                </>
            }
            challengeTxt={`Delete this ${itemType}`}
            onClose={onClose}
            error={error}
            isLoading={isLoading}
        />
    );
};

export default DeleteConfirmationDialog;
