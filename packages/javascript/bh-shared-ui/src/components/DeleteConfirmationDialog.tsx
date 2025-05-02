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
import React from 'react';
import ConfirmationDialog from './ConfirmationDialog';

const DeleteConfirmationDialog: React.FC<{
    open: boolean;
    itemName: string;
    itemType: string;
    onCancel: () => void;
    onConfirm: () => void;
    isLoading?: boolean;
    error?: string;
}> = ({ open, itemName, onCancel, isLoading, error, itemType, onConfirm }) => {
    return (
        <ConfirmationDialog
            open={open}
            title={`Delete ${itemName}?`}
            text={
                <div>
                    <p className='mb-3'>
                        Continuing onwards will delete {itemName} and all associated configurations and findings.
                    </p>
                    <p className='font-bold text-red'>Warning: This change is irreversible.</p>
                </div>
            }
            challengeTxt={`Delete this ${itemType}`}
            onCancel={onCancel}
            onConfirm={onConfirm}
            error={error}
            isLoading={isLoading}
        />
    );
};

export default DeleteConfirmationDialog;
