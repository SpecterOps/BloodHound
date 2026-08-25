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
import { Alert } from 'doodle-ui';
import { SnackbarContent, useSnackbar, VariantType } from 'notistack';
import React from 'react';

interface NotificationSnackbarProps {
    id: string | number;
    message: React.ReactNode;
    variant?: VariantType | null;
    title?: string;
}

export const NotificationSnackbar = React.forwardRef<HTMLDivElement, NotificationSnackbarProps>(
    ({ id, message, variant, title }, ref) => {
        const { closeSnackbar } = useSnackbar();
        return (
            <SnackbarContent ref={ref} className='justify-center'>
                <Alert variant={variant} title={title} onClose={() => closeSnackbar(id)}>
                    {message}
                </Alert>
            </SnackbarContent>
        );
    }
);

NotificationSnackbar.displayName = 'NotificationSnackbar';
