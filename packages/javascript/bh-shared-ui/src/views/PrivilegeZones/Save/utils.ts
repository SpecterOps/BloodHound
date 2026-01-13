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
import { OptionsObject } from 'notistack';

const getErrorMessage = (apiMessage: string, action: string, entity: string) => {
    switch (apiMessage) {
        case 'name must be unique':
            return `Error ${action} ${entity}: ${entity} names must be unique. Please provide a unique name for your new ${entity} and try again.`;

        case 'seeds are required':
            return `To save a ${entity} created using Cypher, the Cypher must be run first. Click "Run" to continue`;

        default:
            return `An unexpected error occurred while ${action} the ${entity}. Message: ${apiMessage}. Please try again.`;
    }
};

export const handleError = (
    error: unknown,
    action: 'creating' | 'updating' | 'deleting',
    entity: 'rule' | 'zone' | 'label',
    addNotification: (notification: string, key?: string, options?: OptionsObject) => void
) => {
    console.error(error);

    const key = `privilege-zones_${action}-${entity}`;

    const options: OptionsObject = { anchorOrigin: { vertical: 'top', horizontal: 'right' } };

    let message = `An unexpected error occurred while ${action} the ${entity}. Please try again.`;

    if (isAxiosError(error)) {
        const errorsList = error.response?.data?.errors ?? [];
        const apiMessage = errorsList.length ? errorsList[0].message : error.response?.statusText || undefined;

        if (apiMessage) {
            message = getErrorMessage(apiMessage, action, entity);
        }
    }

    addNotification(message, key, options);
};
