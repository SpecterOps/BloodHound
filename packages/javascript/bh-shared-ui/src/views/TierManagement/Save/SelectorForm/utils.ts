<<<<<<< HEAD
import { isAxiosError } from 'axios';
=======
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

import { isAxiosError } from 'js-client-library';
>>>>>>> c01a65f741b72a4b3935333b7c1231d7f95d3465
import { OptionsObject } from 'notistack';

export const handleError = (
    error: unknown,
    action: 'creating' | 'updating' | 'deleting',
    addNotification: (notification: string, key?: string, options?: OptionsObject) => void
) => {
    console.error(error);

    const key = `tier-management_${action}-selector`;

    const options: OptionsObject = { anchorOrigin: { vertical: 'top', horizontal: 'right' } };

    const message = isAxiosError(error)
        ? `An unexpected error occurred while ${action} the selector. Message: ${error.response?.statusText}. Please try again.`
        : `An unexpected error occurred while creating the selector. Please try again.`;

    addNotification(message, key, options);
};
