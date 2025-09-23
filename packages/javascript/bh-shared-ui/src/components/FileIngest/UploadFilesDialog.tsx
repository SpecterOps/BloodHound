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
import { useMountEffect, usePermissions } from '../../hooks';
import { useFileUploadDialogContext } from '../../hooks/useFileUploadDialogContext';
import { useNotifications } from '../../providers';
import { Permission } from '../../utils';

export const UploadFilesDialog = () => {
    const { setShowFileIngestDialog } = useFileUploadDialogContext();

    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.GRAPH_DB_INGEST);

    const { addNotification, dismissNotification } = useNotifications();
    const notificationKey = 'file-upload-permission';

    const effect: React.EffectCallback = () => {
        if (!hasPermission) {
            addNotification(
                `Your user role does not grant permission to upload data. Please contact your administrator for details.`,
                notificationKey,
                {
                    persist: true,
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );
        }

        return () => dismissNotification(notificationKey);
    };

    useMountEffect(effect);

    const toggleFileUploadDialog = () => setShowFileIngestDialog((prev) => !prev);

    return (
        <Button
            onClick={() => toggleFileUploadDialog()}
            data-testid='file-ingest_button-upload-files'
            disabled={!hasPermission}>
            Upload File(s)
        </Button>
    );
};
